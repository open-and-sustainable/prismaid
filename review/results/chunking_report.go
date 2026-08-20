package results

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/open-and-sustainable/prismaid/review/chunking"
)

// ChunkingDocumentReport is the persistent audit record for one document in a
// chunked review run. It deliberately keeps merge diagnostics out of the
// configured extraction schema.
type ChunkingDocumentReport struct {
	Filename          string              `json:"filename"`
	CounterMethod     string              `json:"counter_method"`
	FullPromptTokens  int                 `json:"full_prompt_tokens"`
	ChunkPromptTokens []int               `json:"chunk_prompt_tokens"`
	ChunkCount        int                 `json:"chunk_count"`
	Coercions         []chunking.Coercion `json:"coercions"`
	Conflicts         []chunking.Conflict `json:"conflicts"`
	Failures          []chunking.Failure  `json:"failures"`
}

// ChunkingReport records the plan and merge outcome for every source document.
type ChunkingReport struct {
	Documents []ChunkingDocumentReport `json:"documents"`
}

// SaveChunkingReport writes the JSON sidecar used to audit chunking and merge
// decisions. The normal results output is saved separately and is never
// overwritten by this report.
func SaveChunkingReport(resultsFileName string, report ChunkingReport) (string, error) {
	path := resultsFileName + ".chunking-report.json"
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode chunking report: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return "", fmt.Errorf("write chunking report: %w", err)
	}
	return path, nil
}
