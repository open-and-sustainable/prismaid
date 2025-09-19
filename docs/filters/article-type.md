# Article Type Classification Filter

## Overview

The article type classification filter categorizes manuscripts into multiple overlapping types. A single manuscript can belong to several categories simultaneously (e.g., a paper can be both "research_article" AND "empirical_study" AND "sample_study").

## Configuration

### Basic Configuration

```toml
[filters.article_type]
enabled = true
use_ai = false

# Traditional publication type exclusions
exclude_reviews = true
exclude_editorials = true
exclude_letters = true
exclude_case_reports = false
exclude_commentary = false
exclude_perspectives = false

# Methodological type exclusions
exclude_theoretical = false
exclude_empirical = false
exclude_methods = false

# Study scope exclusions
exclude_single_case = false
exclude_sample = false

include_types = []  # If specified, ONLY include these types
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enabled` | boolean | false | Enable/disable the filter |
| `use_ai` | boolean | false | Use AI for classification (requires LLM config) |
| `exclude_reviews` | boolean | false | Exclude all review types (review, systematic_review, meta_analysis) |
| `exclude_editorials` | boolean | false | Exclude editorial articles |
| `exclude_letters` | boolean | false | Exclude letters to editor |
| `exclude_case_reports` | boolean | false | Exclude case report articles |
| `exclude_commentary` | boolean | false | Exclude commentary articles |
| `exclude_perspectives` | boolean | false | Exclude perspective/opinion articles |
| `exclude_theoretical` | boolean | false | Exclude theoretical/conceptual papers |
| `exclude_empirical` | boolean | false | Exclude empirical studies with data |
| `exclude_methods` | boolean | false | Exclude methods/methodology papers |
| `exclude_single_case` | boolean | false | Exclude single case studies (n=1) |
| `exclude_sample` | boolean | false | Exclude sample studies (multiple subjects) |
| `include_types` | array | [] | If specified, ONLY include these types |

## Article Type Categories

### Traditional Publication Types

- **research_article**: Original research with methods and results
- **review**: Literature reviews, narrative reviews
- **systematic_review**: Following structured protocols (e.g., PRISMA)
- **meta_analysis**: Statistical synthesis of multiple studies
- **editorial**: Opinion pieces by editors
- **letter**: Correspondence to editors
- **case_report**: Reports of individual patient cases or specific instances
- **commentary**: Comments on published work
- **perspective**: Author viewpoints and opinion pieces

### Methodological Types

Can overlap with publication types:
- **empirical_study**: Research based on observation or experimentation with data collection
- **theoretical_paper**: Conceptual or theoretical work without empirical data
- **methods_paper**: Paper presenting new methods, techniques, or protocols

### Study Scope Classifications

Applies to empirical studies:
- **single_case_study**: In-depth analysis of a single case, patient, organization, or instance (n=1)
- **sample_study**: Study involving multiple cases, participants, or subjects (includes cohort studies, cross-sectional studies, case-control studies, surveys)

## Classification Methods

### Rule-Based Classification (use_ai = false)

Uses keyword patterns and heuristics:
- Analyzes text for specific indicators (e.g., "systematic review", "editorial", "we conducted", "participants")
- Calculates confidence scores based on keyword frequency and placement
- Fast and deterministic, no API calls required
- Best for: Large datasets, offline processing, consistent reproducible results

### AI-Assisted Classification (use_ai = true)

Uses LLM for semantic understanding:
- Understands context and meaning beyond keyword matching
- Identifies subtle distinctions between article types
- Provides more accurate classification for complex cases
- Returns structured classification with all applicable types

**AI Analysis Dimensions:**
1. Traditional publication types (research, review, editorial, etc.)
2. Methodological approach (empirical, theoretical, methods)
3. Study scope if empirical (single case vs. sample study)

## Output Fields

The filter adds these fields to each manuscript record:

| Field | Type | Description |
|-------|------|-------------|
| `tag_article_types` | string/object | JSON with primary_type and all_types arrays |
| `tag_primary_type` | string | The most specific/important classification |
| `tag_all_types` | array | Complete list of applicable types |
| `tag_methodological_types` | array | Empirical/theoretical/methods classification |
| `tag_scope_types` | array | Single case or sample study designation |
| `exclusion_reason` | string | Explanation if manuscript is excluded |

## Important Notes

- **Article types are NOT mutually exclusive** - papers receive all applicable classifications
- **Exclusion filters check ALL assigned types** - if `exclude_reviews = true`, it excludes papers classified as "review", "systematic_review", OR "meta_analysis"
- **The `include_types` filter overrides exclusions** - specifies exactly which types to accept

## Example Configurations

### Research Articles Only

```toml
[filters.article_type]
enabled = true
use_ai = false
exclude_reviews = true
exclude_editorials = true
exclude_letters = true
exclude_case_reports = true
exclude_commentary = true
exclude_perspectives = true
```

### Empirical Studies with Multiple Subjects

```toml
[filters.article_type]
enabled = true
use_ai = false
exclude_theoretical = true
exclude_single_case = true
include_types = ["empirical_study", "sample_study"]
```

### High-Accuracy Classification with AI

```toml
[filters.article_type]
enabled = true
use_ai = true
exclude_reviews = false
exclude_editorials = true
exclude_letters = true

[[filters.llm]]
provider = "OpenAI"
api_key = ""  # Uses environment variable
model = "gpt-4o-mini"
temperature = 0.01
```

### Accept Only Specific Types

```toml
[filters.article_type]
enabled = true
use_ai = false
include_types = ["research_article", "systematic_review", "meta_analysis"]
# This will ONLY include these three types, regardless of other settings
```

## Classification Examples

| Manuscript Type | Classifications Applied |
|----------------|------------------------|
| Randomized controlled trial | research_article, empirical_study, sample_study |
| Systematic review with meta-analysis | systematic_review, meta_analysis |
| Case report of rare disease | case_report, single_case_study |
| New statistical method paper | methods_paper, research_article |
| Theoretical framework paper | theoretical_paper, research_article |
| Editorial on research ethics | editorial |
| Cohort study of 1000 patients | research_article, empirical_study, sample_study |

## Performance Considerations

### Rule-Based Mode
- **Speed**: Very fast, processes hundreds of manuscripts per second
- **Accuracy**: Good for clear article types with distinct keywords
- **Cost**: Free, no API calls

### AI-Assisted Mode
- **Speed**: Limited by API rate limits
- **Accuracy**: Better for nuanced classification and edge cases
- **Cost**: API costs apply

## Best Practices

1. **Understand Overlapping Types**: Remember that manuscripts can have multiple types
2. **Use Exclusions Carefully**: Excluding one type may affect related types
3. **Consider `include_types`**: Use for precise control over accepted types
4. **AI for Complex Cases**: Use AI mode when dealing with:
   - Interdisciplinary research
   - Novel publication formats
   - Non-standard article structures
5. **Review Classifications**: Check the assigned types in output to verify accuracy

## Filter Order

Article type classification is applied third in the screening pipeline, after deduplication and language detection. This ensures:
- No duplicate processing
- Only manuscripts in accepted languages are classified
- Classification tags available for final inclusion decisions