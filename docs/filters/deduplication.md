# Deduplication Filter

## Overview

The deduplication filter identifies and removes duplicate manuscripts from your dataset using intelligent field comparison and optional AI assistance.

## Configuration

### Basic Configuration

```toml
[filters.deduplication]
enabled = true
use_ai = false
compare_fields = ["title", "authors", "abstract", "doi"]
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enabled` | boolean | false | Enable/disable the filter |
| `use_ai` | boolean | false | Use AI for semantic duplicate detection |
| `compare_fields` | array | ["title", "abstract"] | Fields to compare for duplication |

## How It Works

### Simple Matching (Non-AI)

When `use_ai = false`, the filter uses intelligent field comparison:

**Priority Matching Rules:**
- **DOI Match**: If DOI fields exist and match exactly, records are considered duplicates
- **Combined Fields**: Checks for author + year + (title OR abstract) combinations
- **Single Character Tolerance**: Allows for minor variations (single character differences) in field comparisons
- **Text Normalization**: Automatically handles case differences, extra whitespace, and punctuation variations

**Best for:** Fast processing when records have consistent metadata or minor variations

### AI-Assisted Matching

When `use_ai = true` and LLM is configured, the filter uses semantic understanding:

**AI Capabilities:**
- Recognizes author name variations (initials vs full names, middle names)
- Handles character encoding issues (é→e, ü→u, Müller→Mueller)
- Understands minor title/abstract rephrasing
- Identifies duplicates despite formatting differences

**AI Prompt Used:**
The AI compares manuscripts considering:
- Author name formats (initials vs full names, middle names, order variations)
- Character encoding issues (é→e, ü→u, ñ→n, ø→o, incorrect UTF-8 representation)
- Non-standard character replacements (Müller→Mueller, Gómez→Gomez, Søren→Soren)
- Technical simplifications in database entries
- Minor transcription differences
- Abbreviated vs full journal names
- Different citation styles or formats
- Minor typos or punctuation differences

## Output Fields

The filter adds these fields to each manuscript record:

| Field | Type | Description |
|-------|------|-------------|
| `tag_is_duplicate` | boolean | `true` for duplicates, `false` for originals |
| `tag_duplicate_of` | string | ID of the original record (empty for non-duplicates) |
| `include` | boolean | Set to `false` for duplicates |
| `exclusion_reason` | string | "Duplicate of [ID]" for duplicates |

## Example Configurations

### Basic Deduplication
```toml
[filters.deduplication]
enabled = true
use_ai = false
compare_fields = ["doi", "title"]
```

### Comprehensive Deduplication
```toml
[filters.deduplication]
enabled = true
use_ai = false
compare_fields = ["title", "authors", "abstract", "doi", "year"]
```

### AI-Enhanced Deduplication
```toml
[filters.deduplication]
enabled = true
use_ai = true
compare_fields = ["title", "authors", "abstract"]

[[filters.llm]]
provider = "OpenAI"
api_key = ""  # Uses environment variable
model = "gpt-4o-mini"
temperature = 0.01
```

## Best Practices

1. **Field Selection**: Include multiple fields for better accuracy
2. **DOI Priority**: Always include DOI if available for exact matching
3. **Author Fields**: Include author names to catch same-title different-author papers
4. **AI Usage**: Use AI mode when dealing with:
   - Multiple database sources with different formatting
   - International datasets with character encoding variations
   - Historical data with inconsistent metadata

## Performance Considerations

- **Non-AI Mode**: Very fast, processes thousands of records per second
- **AI Mode**: Limited by API rate limits, typically 10-100 records per second
- **Memory Usage**: Minimal, uses streaming processing

## Filter Order

Deduplication is applied first in the screening pipeline to maximize efficiency by removing duplicates before other processing.