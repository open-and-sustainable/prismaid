// Package revaise provides optional support for documenting prismAId workflows
// with RevAIse review data objects.
//
// The package is intentionally used as an adapter layer between prismAId
// workflow outputs and a RevAIse Review record. It can load an existing JSON or
// YAML record, create a minimal record when none exists, merge stage-specific
// contributions, create a backup snapshot, and save the updated record
// atomically.
//
// RevAIse support is opt-in. Callers should pass a Config value parsed from a
// workflow's optional [revaise] TOML block. If Config.Enabled is false, update
// functions return without changing any files.
//
// # Workflow
//
// A typical workflow uses the same RevAIse record path across multiple
// prismAId stages:
//
//	[revaise]
//	enabled = true
//	record_file = "review.revaise.json"
//	format = "json"
//	schema_version = "0.7.1"
//
// The first enabled stage creates the record if it does not already exist. Each
// later stage reads that same record, creates a timestamped backup snapshot,
// merges its own contribution, and writes the updated record atomically. For
// example, a Zotero download can add a full-text output artifact, a pilot
// screening can add one screening round, a full screening pass can add a second
// screening round, and a review run can add a data extraction stage.
//
// Stage and run identities are important. Screening updates require
// Config.Screening.RoundID so repeated runs can update the same round instead of
// appending duplicates. Extraction updates require Config.Extraction.RunID so a
// pilot extraction and a full extraction can be documented as distinct runs.
//
// The resulting record is a RevAIse Review object that can be shared, versioned,
// inspected, and validated with RevAIse release artifacts. Use versioned
// RevAIse schema URLs for reproducible validation.
//
// The main entry points are:
//
//   - UpdateScreening, for screening stages and screening rounds.
//   - UpdateExtraction, for AI-assisted review/data extraction runs.
//   - UpdateStageOutputs, for stage outputs such as downloaded full-text files.
//
// The lower-level Update function can be used by future adapters to load,
// backup, merge, and save a RevAIse record with custom merge logic.
//
// # References
//
// RevAIse project:
//
//	https://github.com/open-and-sustainable/revaise-model
//
// RevAIse documentation:
//
//	https://revaise-model.readthedocs.io/
//
// RevAIse concept DOI on Zenodo:
//
//	https://doi.org/10.5281/zenodo.17054435
package revaise
