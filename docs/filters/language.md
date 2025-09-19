# Language Detection Filter

## Overview

The language detection filter identifies the primary language of manuscripts and filters based on accepted languages. It can operate in two modes: rule-based pattern matching or AI-assisted semantic detection.

## Configuration

### Basic Configuration

```toml
[filters.language]
enabled = true
accepted_languages = ["en", "es", "fr"]
use_ai = false
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enabled` | boolean | false | Enable/disable the filter |
| `accepted_languages` | array | ["en"] | ISO 639-1 language codes to accept |
| `use_ai` | boolean | false | Use AI for detection (requires LLM config) |

## How It Works

### Processing Order

1. Language detection runs after deduplication (skips already excluded duplicates)
2. Analyzes each manuscript's title, abstract, and journal fields
3. Determines primary language
4. Excludes manuscripts not in the accepted languages list

### Field Priority

- **Title language has priority** over abstract language
- Many scientific databases translate abstracts to English while keeping original titles
- Journal names can indicate regional publications (e.g., "Revista Española", "Deutsche Zeitschrift")

## Detection Methods

### Rule-Based Detection (use_ai = false)

When `use_ai = false`, the filter uses pattern matching:

**Detection Method:**
- Analyzes character scripts (Latin, Cyrillic, CJK, Arabic, Hebrew, Greek)
- Checks for common words in major languages (English, Spanish, French, German, Italian, Portuguese, Dutch, Russian, Chinese, Japanese, Arabic)
- Fast and privacy-preserving (no external API calls)
- Works offline without dependencies

**Limitations:**
- May struggle with short titles or mixed-language content
- Limited to major languages with predefined word lists
- Less accurate for technical/scientific text with many Latin terms

### AI-Assisted Detection (use_ai = true)

When `use_ai = true` and LLM is configured, the filter uses semantic understanding:

**Detection Method:**
- Sends title, abstract, and journal fields to configured LLM
- Uses specialized prompt that understands scientific manuscript conventions
- Recognizes that abstracts are often translated while titles remain in original language
- Handles character encoding variations (é→e, ü→u, ñ→n)
- Identifies primary language even in mixed-language documents

**Graceful Fallback:**
- If no LLM is configured → falls back to rule-based detection
- If API call fails → falls back to rule-based detection
- If response parsing fails → falls back to rule-based detection
- Always provides a result, never fails completely

## Output Fields

The filter adds these fields to each manuscript record:

| Field | Type | Description |
|-------|------|-------------|
| `tag_detected_language` | string | Final detected language (prioritizes title) |
| `tag_title_language` | string | Language detected in title field |
| `tag_abstract_language` | string | Language detected in abstract field |
| `tag_ai_detected_language` | string | Language detected by AI (when use_ai=true) |
| `exclusion_reason` | string | "Language not accepted: [language]" if excluded |

## Supported Language Codes

Common ISO 639-1 language codes:
- `en` - English
- `es` - Spanish
- `fr` - French
- `de` - German
- `it` - Italian
- `pt` - Portuguese
- `nl` - Dutch
- `ru` - Russian
- `zh` - Chinese
- `ja` - Japanese
- `ko` - Korean
- `ar` - Arabic

## Example Configurations

### English-Only Screening
```toml
[filters.language]
enabled = true
accepted_languages = ["en"]
use_ai = false
```

### Multi-Language European Collection
```toml
[filters.language]
enabled = true
accepted_languages = ["en", "es", "fr", "de", "it", "pt"]
use_ai = false
```

### Multi-Language with AI Detection
```toml
[filters.language]
enabled = true
accepted_languages = ["en", "es", "fr", "de"]
use_ai = true

[[filters.llm]]
provider = "OpenAI"
api_key = ""  # Uses environment variable
model = "gpt-4o-mini"
temperature = 0.01
```

### Accept All Languages (Detection Only)
```toml
[filters.language]
enabled = true
accepted_languages = []  # Empty means accept all
use_ai = false  # Still detects and tags language
```

## Performance Considerations

### Rule-Based Mode
- **Speed**: Very fast (milliseconds per manuscript)
- **Accuracy**: Good for major languages with distinct patterns
- **Cost**: Free, no API calls

### AI-Assisted Mode
- **Speed**: Depends on API latency
- **Accuracy**: Better for edge cases and mixed languages
- **Cost**: API costs apply

## Best Practices

1. **Title Priority**: Trust title language over abstract due to translation practices
2. **Journal Context**: Consider journal names as language indicators
3. **AI for Edge Cases**: Use AI mode when dealing with:
   - Regional publications
   - Mixed-language collections
   - Manuscripts with technical Latin terms
4. **Fallback Strategy**: AI mode always falls back to rule-based on errors

## Filter Order

Language detection is applied second in the screening pipeline, after deduplication but before article type classification. This ensures:
- No duplicate processing
- Language tags available for downstream filters
- Efficient exclusion of non-target language papers