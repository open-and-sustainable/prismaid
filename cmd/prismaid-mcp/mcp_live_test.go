package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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
