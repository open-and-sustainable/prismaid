package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	prismaid "github.com/open-and-sustainable/prismaid"
)

// ---- request types -----------------------------------------------------------

// TOMLRequest carries a prismAId TOML configuration string.
type TOMLRequest struct {
	TOML string `json:"toml" jsonschema_description:"prismAId TOML configuration" jsonschema:"required"`
}

// ValidateConfigRequest selects a configuration schema and the TOML to check.
type ValidateConfigRequest struct {
	ConfigType string `json:"config_type" jsonschema_description:"Configuration schema: review, screening, or zotero" jsonschema:"required"`
	TOML       string `json:"toml" jsonschema_description:"prismAId TOML configuration" jsonschema:"required"`
}

// ConformanceRequest carries a RevAIse review-record JSON and a protocol name.
type ConformanceRequest struct {
	RecordJSON string `json:"record_json" jsonschema_description:"RevAIse review-record JSON" jsonschema:"required"`
	Protocol   string `json:"protocol" jsonschema_description:"Protocol identifier, e.g. prisma-2020" jsonschema:"required"`
}

// ProtocolRequest carries a protocol name.
type ProtocolRequest struct {
	Protocol string `json:"protocol" jsonschema_description:"Protocol identifier, e.g. prisma-2020" jsonschema:"required"`
}

// ConvertRequest carries the inputs for a file conversion run.
type ConvertRequest struct {
	InputDir   string `json:"input_dir" jsonschema_description:"Directory containing files to convert" jsonschema:"required"`
	Formats    string `json:"formats" jsonschema_description:"Comma-separated formats to process, e.g. pdf,docx,html" jsonschema:"required"`
	TikaServer string `json:"tika_server,omitempty" jsonschema_description:"Optional Apache Tika server address for OCR fallback, e.g. localhost:9998"`
}

// URLListRequest carries the path to a text file of URLs, one per line.
type URLListRequest struct {
	Path string `json:"path" jsonschema_description:"Path to a text file listing URLs, one per line" jsonschema:"required"`
}

// ---- response types ----------------------------------------------------------

// ErrorInfo describes a tool-level failure without ending the MCP call.
type ErrorInfo struct {
	Code    int    `json:"code" jsonschema_description:"Error code"`
	Message string `json:"message" jsonschema_description:"Error message"`
}

// ValidationResponse reports whether a configuration validated.
type ValidationResponse struct {
	Valid bool       `json:"valid" jsonschema_description:"Whether the configuration is valid"`
	Error *ErrorInfo `json:"error,omitempty" jsonschema_description:"Validation error, if any"`
}

// ConfigResponse returns a generated TOML configuration.
type ConfigResponse struct {
	TOML string `json:"toml" jsonschema_description:"Generated TOML configuration"`
}

// ConformanceResponse returns the protocol conformance report.
type ConformanceResponse struct {
	Report *prismaid.ConformanceReport `json:"report,omitempty" jsonschema_description:"Conformance report: protocol, conforms, and unmet constraints"`
	Error  *ErrorInfo                  `json:"error,omitempty" jsonschema_description:"Error details, if any"`
}

// GuidanceResponse returns a protocol's requirement checklist.
type GuidanceResponse struct {
	Guidance *prismaid.ConformanceGuidance `json:"guidance,omitempty" jsonschema_description:"Protocol requirement checklist and metadata"`
	Error    *ErrorInfo                    `json:"error,omitempty" jsonschema_description:"Error details, if any"`
}

// ReviewResponse returns the outcome of a review run.
type ReviewResponse struct {
	Result *prismaid.ReviewResult `json:"result,omitempty" jsonschema_description:"Review run summary"`
	Error  *ErrorInfo             `json:"error,omitempty" jsonschema_description:"Error details, if any"`
}

// ScreeningResponse returns the outcome of a screening run.
type ScreeningResponse struct {
	Result *prismaid.ScreeningResult `json:"result,omitempty" jsonschema_description:"Screening run summary"`
	Error  *ErrorInfo                `json:"error,omitempty" jsonschema_description:"Error details, if any"`
}

// ConvertResponse returns the outcome of a conversion run.
type ConvertResponse struct {
	Result *prismaid.ConvertResult `json:"result,omitempty" jsonschema_description:"Conversion run summary"`
	Error  *ErrorInfo              `json:"error,omitempty" jsonschema_description:"Error details, if any"`
}

// ZoteroResponse returns the outcome of a Zotero download run.
type ZoteroResponse struct {
	Result *prismaid.ZoteroResult `json:"result,omitempty" jsonschema_description:"Zotero download summary"`
	Error  *ErrorInfo             `json:"error,omitempty" jsonschema_description:"Error details, if any"`
}

// URLListResponse returns the outcome of a URL-list download run.
type URLListResponse struct {
	Result *prismaid.URLListResult `json:"result,omitempty" jsonschema_description:"URL-list download summary"`
	Error  *ErrorInfo              `json:"error,omitempty" jsonschema_description:"Error details, if any"`
}

func main() {
	errLogger := log.New(os.Stderr, "prismaid-mcp: ", log.LstdFlags)

	srv := server.NewMCPServer(
		"prismaid-mcp",
		"0.15.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Design and setup: safe, offline, no API keys.
	srv.AddTool(
		mcp.NewTool("prismaid_validate_config",
			mcp.WithDescription("Validate a prismAId configuration (review, screening, or zotero) without executing it. Offline and read-only."),
			mcp.WithInputSchema[ValidateConfigRequest](),
			mcp.WithOutputSchema[ValidationResponse](),
		),
		mcp.NewStructuredToolHandler(handleValidateConfig),
	)
	srv.AddTool(
		mcp.NewTool("prismaid_generate_review_config",
			mcp.WithDescription("Generate a well-formed review-tool TOML configuration from structured parameters."),
			mcp.WithInputSchema[prismaid.ReviewConfigParams](),
			mcp.WithOutputSchema[ConfigResponse](),
		),
		mcp.NewStructuredToolHandler(handleGenerateReviewConfig),
	)
	srv.AddTool(
		mcp.NewTool("prismaid_generate_screening_config",
			mcp.WithDescription("Generate a well-formed screening-tool TOML configuration from structured parameters."),
			mcp.WithInputSchema[prismaid.ScreeningConfigParams](),
			mcp.WithOutputSchema[ConfigResponse](),
		),
		mcp.NewStructuredToolHandler(handleGenerateScreeningConfig),
	)
	srv.AddTool(
		mcp.NewTool("prismaid_generate_zotero_config",
			mcp.WithDescription("Generate a well-formed Zotero-download TOML configuration from structured parameters."),
			mcp.WithInputSchema[prismaid.ZoteroConfigParams](),
			mcp.WithOutputSchema[ConfigResponse](),
		),
		mcp.NewStructuredToolHandler(handleGenerateZoteroConfig),
	)

	// Protocol conformance: offline symbolic check.
	srv.AddTool(
		mcp.NewTool("prismaid_check_conformance",
			mcp.WithDescription("Check a RevAIse review record against a reporting protocol's SHACL shapes (e.g. prisma-2020). Offline; the verdict comes from the shapes, not the model."),
			mcp.WithInputSchema[ConformanceRequest](),
			mcp.WithOutputSchema[ConformanceResponse](),
		),
		mcp.NewStructuredToolHandler(handleCheckConformance),
	)
	srv.AddTool(
		mcp.NewTool("prismaid_list_protocols",
			mcp.WithDescription("List the protocol identifiers accepted by prismaid_check_conformance."),
		),
		handleListProtocols,
	)
	srv.AddTool(
		mcp.NewTool("prismaid_protocol_guidance",
			mcp.WithDescription("Return a protocol's full requirement checklist (grouped by record class) so a user can plan a conforming review before running anything. Advisory; does not constrain tool order."),
			mcp.WithInputSchema[ProtocolRequest](),
			mcp.WithOutputSchema[GuidanceResponse](),
		),
		mcp.NewStructuredToolHandler(handleProtocolGuidance),
	)

	// Execution: reads and writes files, uses network and LLM API keys from the
	// environment. Configuration file paths are resolved inside the server's own
	// filesystem.
	srv.AddTool(
		mcp.NewTool("prismaid_review",
			mcp.WithDescription("Run a systematic review from a TOML configuration. Reads and writes files and calls LLM APIs using keys from the environment."),
			mcp.WithInputSchema[TOMLRequest](),
			mcp.WithOutputSchema[ReviewResponse](),
		),
		mcp.NewStructuredToolHandler(handleReview),
	)
	srv.AddTool(
		mcp.NewTool("prismaid_screening",
			mcp.WithDescription("Screen manuscripts from a TOML configuration. Reads and writes files; may call LLM APIs when AI-assisted filters are enabled."),
			mcp.WithInputSchema[TOMLRequest](),
			mcp.WithOutputSchema[ScreeningResponse](),
		),
		mcp.NewStructuredToolHandler(handleScreening),
	)
	srv.AddTool(
		mcp.NewTool("prismaid_convert",
			mcp.WithDescription("Convert files (pdf, docx, html) in a directory to plain text."),
			mcp.WithInputSchema[ConvertRequest](),
			mcp.WithOutputSchema[ConvertResponse](),
		),
		mcp.NewStructuredToolHandler(handleConvert),
	)
	srv.AddTool(
		mcp.NewTool("prismaid_download_zotero",
			mcp.WithDescription("Download attachments from a Zotero collection using a TOML configuration."),
			mcp.WithInputSchema[TOMLRequest](),
			mcp.WithOutputSchema[ZoteroResponse](),
		),
		mcp.NewStructuredToolHandler(handleDownloadZotero),
	)
	srv.AddTool(
		mcp.NewTool("prismaid_download_url_list",
			mcp.WithDescription("Download files from a text file of URLs, one per line."),
			mcp.WithInputSchema[URLListRequest](),
			mcp.WithOutputSchema[URLListResponse](),
		),
		mcp.NewStructuredToolHandler(handleDownloadURLList),
	)

	if err := server.ServeStdio(srv, server.WithErrorLogger(errLogger)); err != nil {
		errLogger.Fatalf("server error: %v", err)
	}
}

func handleValidateConfig(ctx context.Context, request mcp.CallToolRequest, args ValidateConfigRequest) (ValidationResponse, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	err := runWithTimeout(ctx, func() error {
		return prismaid.ValidateConfig(args.ConfigType, args.TOML)
	})
	if err != nil {
		return ValidationResponse{Valid: false, Error: errorInfo(400, err.Error())}, nil
	}
	return ValidationResponse{Valid: true}, nil
}

func handleGenerateReviewConfig(ctx context.Context, request mcp.CallToolRequest, args prismaid.ReviewConfigParams) (ConfigResponse, error) {
	return ConfigResponse{TOML: prismaid.GenerateReviewConfig(args)}, nil
}

func handleGenerateScreeningConfig(ctx context.Context, request mcp.CallToolRequest, args prismaid.ScreeningConfigParams) (ConfigResponse, error) {
	return ConfigResponse{TOML: prismaid.GenerateScreeningConfig(args)}, nil
}

func handleGenerateZoteroConfig(ctx context.Context, request mcp.CallToolRequest, args prismaid.ZoteroConfigParams) (ConfigResponse, error) {
	return ConfigResponse{TOML: prismaid.GenerateZoteroConfig(args)}, nil
}

func handleCheckConformance(ctx context.Context, request mcp.CallToolRequest, args ConformanceRequest) (ConformanceResponse, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var report prismaid.ConformanceReport
	err := runWithTimeout(ctx, func() error {
		var err error
		report, err = prismaid.CheckConformance(args.RecordJSON, args.Protocol)
		return err
	})
	if err != nil {
		return ConformanceResponse{Error: errorInfo(400, err.Error())}, nil
	}
	return ConformanceResponse{Report: &report}, nil
}

func handleListProtocols(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	protocols, err := prismaid.ConformanceProtocols()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultStructured(protocols, ""), nil
}

func handleProtocolGuidance(ctx context.Context, request mcp.CallToolRequest, args ProtocolRequest) (GuidanceResponse, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var guidance prismaid.ConformanceGuidance
	err := runWithTimeout(ctx, func() error {
		var err error
		guidance, err = prismaid.ProtocolGuidance(args.Protocol)
		return err
	})
	if err != nil {
		return GuidanceResponse{Error: errorInfo(400, err.Error())}, nil
	}
	return GuidanceResponse{Guidance: &guidance}, nil
}

func handleReview(ctx context.Context, request mcp.CallToolRequest, args TOMLRequest) (ReviewResponse, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var result prismaid.ReviewResult
	err := runWithTimeout(ctx, func() error {
		var err error
		result, err = prismaid.Review(args.TOML)
		return err
	})
	if err != nil {
		return ReviewResponse{Error: errorInfo(400, err.Error())}, nil
	}
	return ReviewResponse{Result: &result}, nil
}

func handleScreening(ctx context.Context, request mcp.CallToolRequest, args TOMLRequest) (ScreeningResponse, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var result prismaid.ScreeningResult
	err := runWithTimeout(ctx, func() error {
		var err error
		result, err = prismaid.Screening(args.TOML)
		return err
	})
	if err != nil {
		return ScreeningResponse{Error: errorInfo(400, err.Error())}, nil
	}
	return ScreeningResponse{Result: &result}, nil
}

func handleConvert(ctx context.Context, request mcp.CallToolRequest, args ConvertRequest) (ConvertResponse, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var result prismaid.ConvertResult
	err := runWithTimeout(ctx, func() error {
		var err error
		result, err = prismaid.Convert(args.InputDir, args.Formats, prismaid.ConvertOptions{TikaServer: args.TikaServer})
		return err
	})
	if err != nil {
		return ConvertResponse{Error: errorInfo(400, err.Error())}, nil
	}
	return ConvertResponse{Result: &result}, nil
}

func handleDownloadZotero(ctx context.Context, request mcp.CallToolRequest, args TOMLRequest) (ZoteroResponse, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var result prismaid.ZoteroResult
	err := runWithTimeout(ctx, func() error {
		var err error
		result, err = prismaid.DownloadZotero(args.TOML)
		return err
	})
	if err != nil {
		return ZoteroResponse{Error: errorInfo(400, err.Error())}, nil
	}
	return ZoteroResponse{Result: &result}, nil
}

func handleDownloadURLList(ctx context.Context, request mcp.CallToolRequest, args URLListRequest) (URLListResponse, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var result prismaid.URLListResult
	err := runWithTimeout(ctx, func() error {
		var err error
		result, err = prismaid.DownloadURLList(args.Path)
		return err
	})
	if err != nil {
		return URLListResponse{Error: errorInfo(400, err.Error())}, nil
	}
	return URLListResponse{Result: &result}, nil
}

func errorInfo(code int, message string) *ErrorInfo {
	return &ErrorInfo{Code: code, Message: message}
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	timeoutSeconds := os.Getenv("PRISMAID_MCP_TIMEOUT_SECONDS")
	if timeoutSeconds == "" {
		return context.WithCancel(ctx)
	}

	seconds, err := strconv.Atoi(timeoutSeconds)
	if err != nil || seconds <= 0 {
		return context.WithCancel(ctx)
	}

	return context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
}

func runWithTimeout(ctx context.Context, fn func() error) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- fn()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
