---
title: RevAIse Integration
layout: default
---

# RevAIse Integration

---

prismAId can optionally document review workflows using [RevAIse](https://github.com/open-and-sustainable/revaise-model) data objects. RevAIse support is disabled by default. When enabled, prismAId reads a user-provided RevAIse review record, merges the current stage output into it, and saves the updated record.

Normal prismAId outputs are unchanged. Screening results, extraction tables, downloaded files, and logs are still written as usual.

## Shared Record Workflow

Use the same `record_file` path across stages to build one cumulative RevAIse review record:

1. Zotero download adds a full-text artifact reference.
2. A pilot screening run adds one screening round.
3. A full screening run adds another screening round.
4. A pilot extraction run adds one data extraction stage.
5. A full extraction run adds another extraction stage or updates the configured extraction run.

If the RevAIse file already exists, prismAId updates it instead of overwriting it from scratch.

## Backups

When RevAIse updates are enabled, backups are enabled by default. Before prismAId modifies an existing record, it writes a timestamped snapshot to `.revaise-backups` next to the record file.

Example:

```text
review.revaise.json
.revaise-backups/review.revaise.20260603T143012.123456789Z.bak.json
```

To disable backups explicitly:

```toml
[revaise]
enabled = true
record_file = "review.revaise.json"
backup = false
```

If `backup` is omitted, backups remain enabled.

## Common Configuration

Add this block to any supported stage configuration:

```toml
[revaise]
enabled = true
record_file = "review.revaise.json"
format = "json"
schema_version = "0.7.1"
```

`format` may be `json` or `yaml`. If omitted, prismAId detects the format from `record_file`; JSON is the default.

### Human Oversight

prismAId records how much a human reviewed the AI output in every AI-assistance entry. Set `human_oversight_level` to describe your process:

```toml
[revaise]
enabled = true
record_file = "review.revaise.json"
human_oversight_level = "FULL_REVIEW"
```

Allowed values are `FULL_REVIEW`, `SAMPLE_REVIEW`, `CONFIDENCE_BASED`, `EXCEPTION_ONLY`, `MINIMAL`, and `NONE`. The default is `NONE`, because prismAId itself performs no human review; raise it to reflect the review you actually carry out.

## Screening Rounds

Each screening run needs an explicit round identity. Reusing the same `round_id` updates the existing round. Using a new `round_id` appends a new round.

The reviewer recorded for a round is identified by `reviewer_id` and carries the role set in `reviewer_role` (default `SCREENER`; other RevAIse roles include `DATA_EXTRACTOR`, `REVIEWER`, and `LEAD_REVIEWER`).

Pilot screening:

```toml
[revaise]
enabled = true
record_file = "review.revaise.json"

[revaise.stage]
stage_type = "screening_title_abstract"
stage_label = "Title and abstract screening"

[revaise.screening_round]
round_id = "ta_pilot_001"
round_type = "TITLE_ABSTRACT"
round_number = 1
round_label = "Pilot title and abstract screening"
reviewer_id = "prismaid"
```

Full screening on already screened records:

```toml
[revaise]
enabled = true
record_file = "review.revaise.json"

[revaise.stage]
stage_type = "screening_title_abstract"
stage_label = "Title and abstract screening"

[revaise.screening_round]
round_id = "ta_full_001"
round_type = "TITLE_ABSTRACT"
round_number = 2
round_label = "Full title and abstract screening"
reviewer_id = "prismaid"
```

## Review and Extraction Runs

The review tool maps prismAId structured extraction results to a RevAIse `data_extraction` stage. Each run needs an explicit `run_id`.

Pilot extraction:

```toml
[revaise]
enabled = true
record_file = "review.revaise.json"

[revaise.stage]
stage_type = "data_extraction"
stage_label = "Pilot AI-assisted extraction"

[revaise.extraction_run]
run_id = "pilot_extraction_001"
label = "Pilot extraction on calibration papers"
form_id = "slca_extraction_form_v1"
form_name = "SLCA extraction form"
form_version = "1"
extractor_id = "prismaid"
```

Full extraction:

```toml
[revaise]
enabled = true
record_file = "review.revaise.json"

[revaise.stage]
stage_type = "data_extraction"
stage_label = "Full AI-assisted extraction"

[revaise.extraction_run]
run_id = "full_extraction_001"
label = "Full extraction on included studies"
form_id = "slca_extraction_form_v1"
form_name = "SLCA extraction form"
form_version = "1"
extractor_id = "prismaid"
```

## Zotero Download

Zotero download configuration can also update the shared RevAIse record. The current integration records `output_dir` as a full-text artifact. It does not yet fetch full parent-item Zotero bibliographic metadata.

```toml
[zotero]
user = "your_zotero_user_id"
api_key = "your_api_key"
group = "Your Group/Collection"
output_dir = "papers/zotero"

[revaise]
enabled = true
record_file = "review.revaise.json"

[revaise.stage]
stage_type = "search"
stage_label = "Zotero full-text download"
```

## Authoring the record directly

Beyond the automatic stage updates above, prismAId exposes tools to build and check a review record by hand — useful for the stages prismAId does not run itself (registration, search, risk of bias, synthesis) and for assembling a record across stages:

- **Seed a record** — create a record with a valid review header and, optionally, empty stubs for the manual stages, ready to fill in.
- **Look up the data model** — describe the released RevAIse schema (field names, required slots, enum values), fetched live from RevAIse, so fields are filled correctly without touching the LinkML source.
- **Merge a stage** — add a stage to an existing record, or fill a seeded stub.
- **Validate a record** — check a record against the released data-model JSON Schema (structural validity — distinct from [protocol conformance](../conformance)).

A typical flow: **seed → look up the schema to fill the stubs → merge each stage as it is completed → validate → [check conformance](../conformance)**, planned with [Protocol Guidance](../guidance).

These are available on every channel:

| Channel | Seed | Describe schema | Merge stage | Validate |
|---|---|---|---|---|
| CLI | `-generate-record params.json` | `-revaise-schema <type>` | `-merge-record rec.json -merge-stage s.json` | `-validate-record rec.json` |
| Go | `GenerateRevAIseRecord` | `RevAIseSchema` | `MergeRecordStage` | `ValidateRecord` |
| Python | `generate_revaise_record` | `revaise_schema` | `merge_record_stage` | `validate_record` |
| R | `GenerateRevAIseRecord` | `RevAIseSchema` | `MergeRecordStage` | `ValidateRecord` |
| Julia | `generate_revaise_record` | `revaise_schema` | `merge_record_stage` | `validate_record` |
| MCP | `prismaid_generate_revaise_record` | `prismaid_revaise_schema` | `prismaid_merge_record_stage` | `prismaid_validate_record` |

The describe-schema and validate tools fetch the released data model live from RevAIse, so they require network access; nothing is vendored.

## Supported Updates

Current RevAIse hooks:

- Screening TOML: `ScreeningStage`, `ScreeningRound`, screening decisions, inclusion and exclusion lists, screening statistics, literature records.
- Review TOML: `ExtractionStage`, extraction form, extracted studies, extracted data points, AI assistance metadata, extraction output artifact.
- Zotero download TOML: full-text output artifact for `zotero.output_dir`.

Unsupported hooks are ignored unless implemented in a future release.

## Next Steps

Once stages are recorded in a RevAIse review record, [check it for protocol conformance](../conformance) against a reporting protocol such as PRISMA 2020, and use [Protocol Guidance](../guidance) to see the protocol's full requirement checklist.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
