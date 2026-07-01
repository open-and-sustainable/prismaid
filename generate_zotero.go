package prismaid

import (
	"fmt"
	"strings"
)

// ZoteroRevAIse describes an optional [revaise] block documenting the Zotero
// download as a RevAIse search stage. A nil *ZoteroRevAIse omits it. The Zotero
// download records only a stage output (the downloaded full texts) and no AI
// assistance, so there is no human-oversight field here.
type ZoteroRevAIse struct {
	RecordFile    string
	Format        string
	SchemaVersion string
	StageLabel    string
}

// ZoteroConfigParams holds the inputs for GenerateZoteroConfig.
type ZoteroConfigParams struct {
	User      string
	APIKey    string
	Group     string
	OutputDir string

	RevAIse *ZoteroRevAIse
}

// GenerateZoteroConfig builds a Zotero-download TOML configuration from params.
// Like the other generators it produces well-formed, deterministic TOML and
// leaves completeness to ValidateConfig("zotero", ...).
func GenerateZoteroConfig(p ZoteroConfigParams) string {
	var b strings.Builder

	b.WriteString("[zotero]\n")
	fmt.Fprintf(&b, "user = %q\n", p.User)
	fmt.Fprintf(&b, "api_key = %q\n", p.APIKey)
	fmt.Fprintf(&b, "group = %q\n", p.Group)
	fmt.Fprintf(&b, "output_dir = %q\n", p.OutputDir)

	result := strings.TrimSpace(b.String())
	if p.RevAIse != nil {
		result += "\n\n" + zoteroRevAIseSection(*p.RevAIse)
	}
	return result
}

func zoteroRevAIseSection(r ZoteroRevAIse) string {
	var b strings.Builder
	b.WriteString("[revaise]\n")
	b.WriteString("enabled = true\n")
	fmt.Fprintf(&b, "record_file = %q\n", r.RecordFile)
	fmt.Fprintf(&b, "format = %q\n", r.Format)
	fmt.Fprintf(&b, "schema_version = %q\n", r.SchemaVersion)
	b.WriteString("\n[revaise.stage]\n")
	b.WriteString("stage_type = \"search\"\n")
	fmt.Fprintf(&b, "stage_label = %q\n", r.StageLabel)
	return strings.TrimSpace(b.String())
}
