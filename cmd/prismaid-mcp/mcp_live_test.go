package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestMCPReviewLive drives the full authoring-to-execution path over the stdio
// MCP transport: it generates a review configuration, validates it, and runs the
// review against a real OpenAI endpoint. It is skipped in short mode and when no
// API key is available, so it never runs in CI without explicit credentials.
func TestMCPReviewLive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MCP live test in short mode")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping MCP live test: missing OPENAI_API_KEY")
	}

	cmdPath, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve command path: %v", err)
	}

	cli, err := client.NewStdioMCPClient("go", nil, "run", cmdPath)
	if err != nil {
		t.Fatalf("start MCP client: %v", err)
	}
	defer cli.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "prismaid-mcp-live-test", Version: "0.1.0"}
	initReq.Params.Capabilities = mcp.ClientCapabilities{}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if _, err := cli.Initialize(ctx, initReq); err != nil {
		t.Fatalf("initialize MCP client: %v", err)
	}

	inputDir := t.TempDir()
	manuscript := filepath.Join(inputDir, "manuscript.txt")
	if err := os.WriteFile(manuscript, []byte(
		"This randomized controlled trial evaluated a new therapy in 200 patients over 12 months."), 0o600); err != nil {
		t.Fatalf("write manuscript: %v", err)
	}
	resultsFile := filepath.Join(inputDir, "results")

	// 1) Generate a review configuration from structured parameters.
	genArgs := map[string]any{
		"Name":            "MCP live test",
		"Author":          "prismaid-mcp",
		"Version":         "1.0",
		"InputDirectory":  inputDir,
		"ResultsFileName": resultsFile,
		"OutputFormat":    "json",
		"LogLevel":        "low",
		"LLMs": []map[string]any{
			{"Provider": "OpenAI", "APIKey": apiKey, "Model": "gpt-4o-mini", "Temperature": 0.01},
		},
		"Persona":        "You are a systematic-review assistant.",
		"Task":           "Extract the study design from the manuscript.",
		"ExpectedResult": "A JSON object with the requested fields.",
		"ReviewItems": []map[string]any{
			{"Key": "study_design", "Values": []string{"rct", "observational", "review"}},
		},
	}
	toml := callString(ctx, t, cli, "prismaid_generate_review_config", genArgs, "toml")

	// 2) Validate the generated configuration.
	var validation struct {
		Valid bool `json:"valid"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	callInto(ctx, t, cli, "prismaid_validate_config",
		map[string]any{"config_type": "review", "toml": toml}, &validation)
	if !validation.Valid {
		msg := ""
		if validation.Error != nil {
			msg = validation.Error.Message
		}
		t.Fatalf("generated configuration did not validate: %s", msg)
	}

	// 3) Run the review.
	var review struct {
		Result *struct {
			OutputFile           string `json:"OutputFile"`
			ManuscriptsProcessed int    `json:"ManuscriptsProcessed"`
			ReviewItems          int    `json:"ReviewItems"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	callInto(ctx, t, cli, "prismaid_review", map[string]any{"toml": toml}, &review)
	if review.Error != nil {
		t.Fatalf("review tool error: %s", review.Error.Message)
	}
	if review.Result == nil {
		t.Fatalf("review returned no result")
	}
	if review.Result.ManuscriptsProcessed != 1 {
		t.Fatalf("expected 1 manuscript processed, got %d", review.Result.ManuscriptsProcessed)
	}
	if review.Result.OutputFile == "" {
		t.Fatalf("review returned an empty output file")
	}
}

// TestMCPChunkingPlan exercises the complete offline chunking workflow through
// the local stdio MCP server. It verifies nested generator inputs, review
// validation, and planning against a temporary source file without an API key
// or a model call.
func TestMCPChunkingPlan(t *testing.T) {
	cmdPath, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve command path: %v", err)
	}

	cli, err := client.NewStdioMCPClient("go", nil, "run", cmdPath)
	if err != nil {
		t.Fatalf("start MCP client: %v", err)
	}
	defer cli.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "prismaid-mcp-chunking-test", Version: "0.1.0"}
	initReq.Params.Capabilities = mcp.ClientCapabilities{}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if _, err := cli.Initialize(ctx, initReq); err != nil {
		t.Fatalf("initialize MCP client: %v", err)
	}

	inputDir := t.TempDir()
	manuscript := filepath.Join(inputDir, "long-manuscript.txt")
	paragraph := "This randomized controlled trial evaluated a new therapy in 200 patients over 12 months. "
	if err := os.WriteFile(manuscript, []byte(strings.Repeat(paragraph+"\n\n", 400)), 0o600); err != nil {
		t.Fatalf("write manuscript: %v", err)
	}

	// 1) Supply the nested chunking plan through the MCP generator schema.
	genArgs := map[string]any{
		"Name":            "MCP chunking test",
		"Author":          "prismaid-mcp",
		"Version":         "1.0",
		"InputDirectory":  inputDir,
		"ResultsFileName": filepath.Join(inputDir, "results"),
		"OutputFormat":    "json",
		"LogLevel":        "low",
		"LLMs": []map[string]any{
			{"Provider": "OpenAI", "Model": "gpt-4o-mini", "Temperature": 0.01},
		},
		"Persona":        "You are a systematic-review assistant.",
		"Task":           "Extract the study design from the manuscript.",
		"ExpectedResult": "A JSON object with the requested fields.",
		"ReviewItems": []map[string]any{
			{"Key": "study_design", "Values": []string{"rct", "observational", "not_specified"}},
		},
		"Chunking": map[string]any{
			"Enabled":            true,
			"InputContextTokens": 512,
			"OverlapTokens":      32,
			"MergeRules": []map[string]any{
				{
					"Key":      "study_design",
					"Rule":     "categorical",
					"Defaults": []string{"not_specified"},
					"TieBreak": "first",
				},
			},
		},
	}
	toml := callString(ctx, t, cli, "prismaid_generate_review_config", genArgs, "toml")
	if !strings.Contains(toml, "[project.configuration.chunking]") ||
		!strings.Contains(toml, "[project.configuration.chunking.merge.\"study_design\"]") {
		t.Fatalf("generated TOML is missing the chunking plan:\n%s", toml)
	}

	// 2) Validate the enabled plan through the MCP validation tool.
	var validation ValidationResponse
	callInto(ctx, t, cli, "prismaid_validate_config",
		map[string]any{"config_type": "review", "toml": toml}, &validation)
	if !validation.Valid {
		message := ""
		if validation.Error != nil {
			message = validation.Error.Message
		}
		t.Fatalf("generated chunking configuration did not validate: %s", message)
	}

	// 3) Plan the temporary manuscript through the MCP planning tool. This only
	// reads the text and must report that the low user-authored limit requires
	// more than one chunk.
	var response ChunkingPlanResponse
	callInto(ctx, t, cli, "prismaid_plan_review_chunking", map[string]any{"toml": toml}, &response)
	if response.Error != nil {
		t.Fatalf("chunking planner returned an error: %s", response.Error.Message)
	}
	if response.Plan == nil || len(response.Plan.Documents) != 1 {
		t.Fatalf("expected a plan for one manuscript, got %#v", response.Plan)
	}
	document := response.Plan.Documents[0]
	if document.Filename != "long-manuscript" {
		t.Fatalf("unexpected planned filename: %q", document.Filename)
	}
	if document.ChunkCount < 2 || len(document.ChunkPromptTokens) != document.ChunkCount {
		t.Fatalf("expected a multi-chunk plan with per-chunk sizes, got %#v", document)
	}
}

// callString calls a tool and returns a single string field from its structured
// output.
func callString(ctx context.Context, t *testing.T, cli *client.Client, name string, args map[string]any, field string) string {
	t.Helper()
	var out map[string]any
	callInto(ctx, t, cli, name, args, &out)
	value, ok := out[field].(string)
	if !ok || value == "" {
		t.Fatalf("%s: missing or empty %q in output", name, field)
	}
	return value
}

// callInto calls a tool and decodes its structured output into target.
func callInto(ctx context.Context, t *testing.T, cli *client.Client, name string, args map[string]any, target any) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args

	result, err := cli.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("%s: call failed: %v", name, err)
	}
	if result.StructuredContent == nil {
		t.Fatalf("%s: response has no structured content", name)
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("%s: marshal structured content: %v", name, err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("%s: decode structured content: %v", name, err)
	}
}
