---
title: Screening Tool
layout: default
---

# Screening Tool

---

<details>
<summary><strong>Page Contents</strong></summary>
<ul>
  <li><a href="#purpose-and-capabilities"><strong>Purpose and Capabilities</strong></a>: what the Screening tool does and why it's essential</li>
  <li><a href="#usage-methods"><strong>Usage Methods</strong></a>: how to use the tool across different platforms and programming languages</li>
  <li><a href="#configuration-file-structure"><strong>Configuration File Structure</strong></a>: detailed explanation of the TOML configuration</li>
  <li><a href="#screening-filters"><strong>Screening Filters</strong></a>: available filters and their options</li>
  <li><a href="#input-and-output-formats"><strong>Input and Output Formats</strong></a>: supported file formats and data structures</li>
  <li><a href="#best-practices"><strong>Best Practices</strong></a>: recommendations for effective screening</li>
  <li><a href="#workflow-integration"><strong>Workflow Integration</strong></a>: how the Screening tool fits into your systematic review process</li>
  <li><a href="#troubleshooting"><strong>Troubleshooting</strong></a>: solutions to common issues</li>
</ul>
</details>

---

## Purpose and Capabilities

The prismAId Screening tool automates the filtering phase of systematic literature reviews by identifying and tagging manuscripts for potential exclusion. This critical step occurs after the initial literature search but before downloading full texts, helping researchers focus on relevant literature by:

1. **Deduplication**: Identifies and removes duplicate manuscripts using various matching algorithms
2. **Language Filtering**: Detects manuscript language and filters based on accepted languages
3. **Article Type Classification**: Identifies article types (research articles, reviews, editorials, etc.) for selective inclusion/exclusion
4. **Batch Processing**: Efficiently processes large volumes of manuscripts with minimal manual intervention
5. **Transparent Tagging**: Provides clear reasons for exclusions and maintains complete audit trails
6. **AI-Assisted Analysis**: Optional integration with LLMs for enhanced classification accuracy

The Screening tool bridges the gap between literature search and paper acquisition, ensuring that only relevant, unique manuscripts are downloaded and proceed to the full review phase.

## Usage Methods

The Screening tool can be accessed through multiple interfaces to accommodate different workflows:

### Binary (Command Line)

```bash
# Run screening with a TOML configuration file
./prismaid -screening screening_config.toml
```

### Go Package

```go
import "github.com/open-and-sustainable/prismaid"

// Run screening with a TOML configuration string
tomlConfig := "..." // Your TOML configuration as a string
err := prismaid.Screening(tomlConfig)
```

### Python Package

```python
import prismaid

# Run screening with a TOML configuration file
with open("screening_config.toml", "r") as file:
    toml_config = file.read()
prismaid.screening(toml_config)
```

### R Package

```r
library(prismaid)

# Run screening with a TOML configuration file
toml_content <- paste(readLines("screening_config.toml"), collapse = "\n")
Screening(toml_content)
```

### Julia Package

```julia
using PrismAId

# Run screening with a TOML configuration file
toml_config = read("screening_config.toml", String)
PrismAId.screening(toml_config)
```

## Configuration File Structure

The Screening tool is driven by a TOML configuration file that defines all aspects of your screening process. Here's the complete structure:

### Project Section

```toml
[project]
name = "Manuscript Screening Example"       # Project title
author = "John Doe"                        # Project author
version = "1.0"                           # Configuration version
input_file = "/path/to/manuscripts.csv"    # Input CSV or TXT file
output_file = "/path/to/results"          # Output path (without extension)
text_column = "abstract"                  # Column with text/file paths
identifier_column = "doi"                 # Column with unique IDs
output_format = "csv"                     # "csv" or "json"
log_level = "medium"                      # "low", "medium", or "high"
```

### Filters Section

The filters section controls which screening criteria to apply:

```toml
[filters]

[filters.deduplication]
enabled = true
use_ai = false                            # Use AI for similarity detection
compare_fields = ["title", "abstract", "doi", "authors", "year"]  # Fields to compare for duplication

[filters.language]
enabled = true
accepted_languages = ["en", "es", "fr"]   # ISO language codes
use_ai = false                            # Use AI for detection (recommended for better accuracy)

[filters.article_type]
enabled = true
use_ai = false                             # Use AI for classification (requires LLM config)

# Traditional publication type exclusions
exclude_reviews = true                    # Exclude all review types (review, systematic_review, meta_analysis)
exclude_editorials = true                 # Exclude editorials
exclude_letters = true                    # Exclude letters to editor
exclude_case_reports = false              # Exclude case reports
exclude_commentary = false                # Exclude commentary articles
exclude_perspectives = false              # Exclude perspective articles

# Methodological type exclusions (can overlap with publication types)
exclude_theoretical = false               # Exclude theoretical/conceptual papers
exclude_empirical = false                 # Exclude empirical studies with data
exclude_methods = false                   # Exclude methods/methodology papers

# Study scope exclusions (applies to empirical studies)
exclude_single_case = false               # Exclude single case studies (n=1, individual cases)
exclude_sample = false                    # Exclude sample studies (cohorts, cross-sectional, multiple subjects)

include_types = []                        # If specified, ONLY include these types
                                         # Available types: "research_article", "review", "systematic_review",
                                         # "meta_analysis", "editorial", "letter", "case_report", "commentary",
                                         # "perspective", "empirical_study", "theoretical_paper", "methods_paper",
                                         # "single_case_study", "sample_study"
```

### LLM Configuration (Optional)

For AI-assisted screening:

```toml
[filters.llm.1]
provider = "OpenAI"                       # AI provider
api_key = ""                              # API key (uses env if empty)
model = "gpt-4o-mini"                     # Model name
temperature = 0.01                        # Model temperature
tpm_limit = 0                             # Tokens per minute limit
rpm_limit = 0                             # Requests per minute limit
```

## Screening Filters

### Deduplication Filter

The deduplication filter identifies duplicate manuscripts using two methods:

#### 1. Simple Matching (Non-AI)
When `use_ai = false`, the filter uses intelligent field comparison:

**Priority Matching Rules:**
- **DOI Match**: If DOI fields exist and match exactly, records are considered duplicates
- **Combined Fields**: Checks for author + year + (title OR abstract) combinations
- **Single Character Tolerance**: Allows for minor variations (single character differences) in field comparisons
- **Text Normalization**: Automatically handles case differences, extra whitespace, and punctuation variations

**Best for:** Fast processing when records have consistent metadata or minor variations

#### 2. AI-Assisted Matching
When `use_ai = true` and LLM is configured, the filter uses semantic understanding:

**AI Capabilities:**
- Recognizes author name variations (initials vs full names, middle names)
- Handles character encoding issues (é→e, ü→u, Müller→Mueller)
- Understands minor title/abstract rephrasing
- Identifies duplicates despite formatting differences

**AI Prompt Used:**
```
You are a scientific reviewer tasked with identifying duplicate manuscripts in a research database. You are provided with specific fields from two different records to compare.

CONTEXT:
- You are comparing the following fields: [configured fields]
- Records may have variations due to:
  * Author name formats (initials vs full names, middle names, order variations)
  * Character encoding issues (é→e, ü→u, ñ→n, ø→o, incorrect UTF-8 representation)
  * Non-standard character replacements (Müller→Mueller, Gómez→Gomez, Søren→Soren)
  * Technical simplifications in database entries
  * Minor transcription differences
  * Abbreviated vs full journal names
  * Different citation styles or formats
  * Minor typos or punctuation differences

IMPORTANT CONSIDERATIONS:
- If DOI is provided and identical, they are definitely duplicates
- For author names: "Smith, J." and "Smith, John" likely refer to the same person
- Character variations: "Müller" and "Mueller" or "André" and "Andre" are likely the same
- For titles: ignore minor differences in capitalization, punctuation, or small words
- For years: same year is a strong indicator if other fields match
- For abstracts: similar content with different phrasing may still be duplicates

MANUSCRIPT 1:
[field data]

MANUSCRIPT 2:
[field data]

TASK: Determine if these represent the same publication.
Respond with ONLY a JSON object: {"duplicate": true} or {"duplicate": false}
```

**Output Format:**
Duplicates are marked in the output with:
- `tag_is_duplicate`: `true` for duplicates, `false` for originals
- `tag_duplicate_of`: Contains the ID of the original record (empty for non-duplicates)
- `include`: Set to `false` for duplicates
- `exclusion_reason`: "Duplicate of [ID]" for duplicates

### Language Detection Filter

The language detection filter identifies the primary language of manuscripts and filters based on accepted languages. It can operate in two modes: rule-based or AI-assisted.

#### Configuration

```toml
[filters.language]
enabled = true                             # Enable/disable language filter
accepted_languages = ["en", "es", "fr"]   # ISO 639-1 language codes to accept
use_ai = false                            # Use AI for detection (requires LLM config)
```

#### Working Mechanism

**Processing Order:**
1. Language detection runs after deduplication (skips already excluded duplicates)
2. Analyzes each manuscript's title, abstract, and journal fields
3. Determines primary language
4. Excludes manuscripts not in the accepted languages list

**Field Priority:**
- **Title language has priority** over abstract language
- Many scientific databases translate abstracts to English while keeping original titles
- Journal names can indicate regional publications (e.g., "Revista Española", "Deutsche Zeitschrift")

#### Rule-Based Detection (use_ai = false)

When `use_ai = false`, the filter uses pattern matching:

**Detection Method:**
- Analyzes character scripts (Latin, Cyrillic, CJK, Arabic, Hebrew, Greek)
- Checks for common words in major languages (English, Spanish, French, German, Italian, Portuguese, Dutch, Russian, Chinese, Japanese, Arabic)
- Fast and privacy-preserving (no external API calls)
- Works offline without dependencies

**Output Tags:**
- `tag_detected_language`: Final detected language (prioritizes title)
- `tag_title_language`: Language detected in title field
- `tag_abstract_language`: Language detected in abstract field

**Limitations:**
- May struggle with short titles or mixed-language content
- Limited to major languages with predefined word lists
- Less accurate for technical/scientific text with many Latin terms

#### AI-Assisted Detection (use_ai = true)

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

**Output Tags:**
- `tag_detected_language`: Final detected language from AI
- `tag_ai_detected_language`: Language detected by AI (for debugging)

#### Example Configurations

**Basic rule-based filtering (accept only English):**
```toml
[filters.language]
enabled = true
accepted_languages = ["en"]
use_ai = false
```

**Multi-language with AI detection:**
```toml
[filters.language]
enabled = true
accepted_languages = ["en", "es", "fr", "de"]
use_ai = true

[filters.llm.1]
provider = "OpenAI"
api_key = ""  # Uses environment variable
model = "gpt-4o-mini"
temperature = 0.01
```

#### Exclusion Handling

Manuscripts excluded by language filter will have:
- `include`: Set to `false`
- `exclusion_reason`: "Language not accepted: [detected_language]"
- Language detection tags preserved for review

#### Performance Considerations

**Rule-based:**
- Very fast (milliseconds per manuscript)
- No API costs
- Consistent results
- Best for: English-only filtering, quick screening, offline use

**AI-assisted:**
- Slower (depends on API latency)
- API costs apply
- More accurate for edge cases
- Best for: Multi-language collections, manuscripts with mixed languages, regional publications

### Article Type Classification Filter

Classifies manuscripts into multiple overlapping categories. A single manuscript can belong to several types simultaneously (e.g., a paper can be both "research_article" AND "empirical_study" AND "sample_study").

#### Operating Modes

The filter can operate in two modes:

##### 1. Rule-based Classification (`use_ai = false`)
- Uses keyword patterns and heuristics to classify articles
- Analyzes text for specific indicators (e.g., "systematic review", "editorial", "we conducted", "participants")
- Calculates confidence scores based on keyword frequency and placement
- Fast and deterministic, no API calls required
- Best for: Large datasets, offline processing, consistent reproducible results

##### 2. AI-Assisted Classification (`use_ai = true`)
- Uses configured LLM models for semantic understanding of title and abstract
- Understands context and meaning beyond keyword matching
- Can identify subtle distinctions between article types
- Provides more accurate classification for complex or ambiguous cases
- Best for: High-accuracy requirements, nuanced classification needs

**Traditional Publication Types:**
- **Research Articles**: Original research with methods and results
- **Review Articles**: Literature reviews, narrative reviews  
- **Systematic Reviews**: Following structured protocols (e.g., PRISMA)
- **Meta-Analyses**: Statistical synthesis of multiple studies
- **Editorials**: Opinion pieces by editors
- **Letters**: Correspondence to editors
- **Case Reports**: Reports of individual patient cases or specific instances
- **Commentary**: Comments on published work
- **Perspectives**: Author viewpoints and opinion pieces

**Methodological Types (can overlap with publication types):**
- **Empirical Study**: Research based on observation or experimentation with data collection
- **Theoretical Paper**: Conceptual or theoretical work without empirical data
- **Methods Paper**: Paper presenting new methods, techniques, or protocols

**Study Scope Classifications (applies to empirical studies):**
- **Single Case Study**: In-depth analysis of a single case, patient, organization, or instance (n=1)
- **Sample Study**: Study involving multiple cases, participants, or subjects (includes cohort studies, cross-sectional studies, case-control studies, surveys)

**Important Notes:**
- Article types are not mutually exclusive - papers receive all applicable classifications
- Exclusion filters check against ALL assigned types
- For example, if `exclude_reviews = true`, it will exclude papers classified as "review", "systematic_review", OR "meta_analysis"
- The `include_types` filter allows specifying exactly which types to accept, overriding exclusion rules

#### AI-Assisted Classification Details

When `use_ai = true`, the filter sends the manuscript's title and abstract to the configured LLM with a comprehensive prompt that:

1. **Requests multiple overlapping classifications** - The AI is explicitly instructed that papers can belong to multiple categories
2. **Evaluates three dimensions**:
   - Traditional publication types (research_article, review, editorial, etc.)
   - Methodological types (empirical_study, theoretical_paper, methods_paper)
   - Study scope (single_case_study, sample_study)
3. **Returns structured JSON** with all applicable types

**AI Prompt Components:**

The AI receives:
- **Manuscript title**: Primary indicator of article type and scope
- **Abstract**: Full content for semantic analysis
- **Classification instructions**: Explicit guidance on overlapping categories

The prompt instructs the AI to:
```
1. Analyze the manuscript for ALL applicable categories
2. Identify traditional publication type (research, review, editorial, etc.)
3. Determine methodological approach (empirical, theoretical, methods)
4. Assess study scope if empirical (single case vs. sample study)
5. Return structured JSON with:
   - primary_type: The most specific/important classification
   - all_types: Complete list of applicable types
   - methodological_types: Empirical/theoretical/methods classification
   - scope_types: Single case or sample study designation
```

**Example AI Response:**
```json
{
  "primary_type": "research_article",
  "all_types": ["research_article", "empirical_study", "sample_study"],
  "methodological_types": ["empirical_study"],
  "scope_types": ["sample_study"]
}
```

**Fallback Behavior:**
- If AI classification fails or no LLM is configured, automatically falls back to rule-based classification
- Ensures consistent operation regardless of AI availability
- Logs warnings when falling back to maintain transparency

## Filter Interaction and Processing Order

The screening tool applies filters sequentially, which optimizes performance and ensures clear exclusion tracking:

### Processing Pipeline

```
Input Manuscripts
    ↓
[1] DEDUPLICATION FILTER
    ├─ Identifies duplicates
    ├─ Marks with: tag_is_duplicate=true
    └─ Sets: include=false, exclusion_reason="Duplicate of [ID]"
    ↓
[2] LANGUAGE FILTER
    ├─ Skips already excluded records
    ├─ Detects language (title priority)
    └─ Excludes non-accepted languages
    ↓
[3] ARTICLE TYPE FILTER
    ├─ Skips already excluded records
    ├─ Classifies article types
    └─ Excludes specified types
    ↓
Final Output (CSV/JSON)
```

### Key Principles

1. **Sequential Processing**: Filters run in order - deduplication → language → article type
2. **Exclusion Preservation**: Once excluded, a manuscript is not reprocessed by subsequent filters
3. **Single Exclusion Reason**: Each manuscript shows only the first reason for exclusion
4. **Performance Optimization**: Skipping excluded records reduces API calls and processing time
5. **Tag Accumulation**: Included manuscripts may have tags from multiple filters

### Example Filter Interaction

Given this configuration:
```toml
[filters.deduplication]
enabled = true
use_ai = false
compare_fields = ["doi", "title"]

[filters.language]
enabled = true
accepted_languages = ["en"]
use_ai = false

[filters.article_type]
enabled = true
exclude_editorials = true
exclude_theoretical = true                # Focus on empirical work only
```

Processing flow for a duplicate Spanish editorial:
1. **Deduplication**: Marked as duplicate → excluded (exclusion_reason: "Duplicate of 123")
2. **Language**: Skipped (already excluded) → no language detection performed
3. **Article Type**: Skipped (already excluded) → no type classification performed

Result: Single exclusion reason preserved, no unnecessary processing.

## Output Format

The screening tool saves results with comprehensive information about each manuscript and the applied filters:

### CSV Output Structure

The CSV output includes the following column types:

1. **Original Data Columns**: All columns from the input file are preserved
2. **Tag Columns**: Added with prefix `tag_` containing filter results:
   - `tag_is_duplicate`: `true` if duplicate, `false` or empty otherwise
   - `tag_duplicate_of`: ID of the original record if duplicate
   - `tag_detected_language`: Primary language detected (prioritizes title)
   - `tag_title_language`: Language detected in title (when non-AI mode)
   - `tag_abstract_language`: Language detected in abstract (when non-AI mode)
   - `tag_article_type`: Classified article type (e.g., research_article, empirical_study, single_case_study)
3. **Status Columns**:
   - `include`: `true` for included records, `false` for excluded
   - `exclusion_reason`: Explanation for exclusion (e.g., "Duplicate of 123", "Language not accepted: fr")

### Filter Processing Order

Filters are applied sequentially, and excluded records are not reprocessed:

1. **Deduplication**: Marks duplicates, sets `include=false` with reason
2. **Language**: Skips already excluded records, processes only included ones
3. **Article Type**: Skips already excluded records, processes only included ones

This ensures:
- Exclusion reasons are preserved from the first filter that excludes a record
- Processing efficiency by not running unnecessary filters on excluded records
- Clear traceability of why each record was excluded

### Language Detection Priority

When using non-AI language detection:
- **Title language takes priority** over abstract language
- Many journals translate abstracts to English while keeping original titles
- Both `title_language` and `abstract_language` tags are saved for transparency
- The final `detected_language` uses title language when available and valid

### Example CSV Output

```csv
title,abstract,doi,tag_is_duplicate,tag_duplicate_of,tag_detected_language,tag_title_language,tag_abstract_language,include,exclusion_reason
"Climate Study","Research on climate...","10.1234",false,,en,en,en,true,
"Climate Study","Research on climate...","10.1234",true,1,,,,,false,"Duplicate of 1"
"Étude climatique","Cette recherche...","10.5678",false,,fr,fr,fr,false,"Language not accepted: fr"
```

## Practical Examples

### Example 1: Basic English-Only Screening

**Scenario**: Screen manuscripts keeping only English research articles, removing duplicates.

```toml
[project]
name = "English Literature Review"
author = "Research Team"
version = "1.0"
input_file = "./manuscripts.csv"
output_file = "./screened_results"
text_column = "abstract"
identifier_column = "id"
output_format = "csv"
log_level = "medium"

[filters.deduplication]
enabled = true
use_ai = false
compare_fields = ["doi", "title", "authors"]

[filters.language]
enabled = true
accepted_languages = ["en"]
use_ai = false

[filters.article_type]
enabled = true
use_ai = false                            # Using rule-based classification
exclude_reviews = false                   # Keep reviews for literature review
exclude_editorials = true
exclude_letters = true
exclude_case_reports = false
exclude_commentary = false
exclude_perspectives = false
exclude_theoretical = false
exclude_empirical = false
exclude_methods = false
exclude_single_case = true                # Focus on studies with multiple subjects
exclude_sample = false
```

### Example 2: Multi-Language Screening with AI

**Scenario**: Accept manuscripts in English, Spanish, and Portuguese, using AI for accurate detection.

```toml
[project]
name = "Latin American Climate Research"
input_file = "./la_climate_papers.csv"
output_file = "./filtered_papers"

[filters.deduplication]
enabled = true
use_ai = true  # AI helps with author name variations
compare_fields = ["title", "authors", "year"]

[filters.language]
enabled = true
accepted_languages = ["en", "es", "pt"]
use_ai = true  # Better for regional language variants

[filters.article_type]
enabled = true
use_ai = true  # AI classification for better accuracy
exclude_reviews = false
exclude_editorials = true

[filters.llm.1]
provider = "OpenAI"
api_key = ""  # Uses OPENAI_API_KEY env variable
model = "gpt-4o-mini"
temperature = 0.01
```

### Example 3: Strict Deduplication for Systematic Review

**Scenario**: Aggressive deduplication for systematic review, accepting only primary research articles.

```toml
[project]
name = "Systematic Review Screening"
log_level = "high"  # Detailed logging for audit trail

[filters.deduplication]
enabled = true
use_ai = false  # Faster for large datasets
compare_fields = ["doi", "title", "authors", "year", "abstract"]

[filters.language]
enabled = true
accepted_languages = ["en"]
use_ai = false

[filters.article_type]
enabled = true
use_ai = false                            # Using rule-based classification
exclude_reviews = true                    # No reviews (includes systematic reviews and meta-analyses)
exclude_editorials = true                 # No editorials
exclude_letters = true                    # No letters
exclude_case_reports = true               # No case reports
exclude_commentary = true                 # No commentary
exclude_perspectives = true               # No perspectives
exclude_theoretical = true                # Only empirical work
exclude_empirical = false                 # Keep empirical studies
exclude_methods = false                   # Keep methods papers
exclude_single_case = true                # Only studies with samples
exclude_sample = false                    # Keep sample studies
include_types = ["empirical_study", "sample_study"]  # Focus on empirical research with samples

### Example 3b: Same Screening with AI Classification

**Scenario**: Same requirements but using AI for more accurate article type classification.

```toml
[project]
name = "Systematic Review Screening with AI"
log_level = "high"

[filters.deduplication]
enabled = true
use_ai = false
compare_fields = ["doi", "title", "authors", "year", "abstract"]

[filters.language]
enabled = true
accepted_languages = ["en"]
use_ai = true  # AI for better language detection

[filters.article_type]
enabled = true
use_ai = true  # AI for comprehensive type classification
exclude_reviews = true
exclude_editorials = true
exclude_letters = true
exclude_case_reports = true
exclude_commentary = true
exclude_perspectives = true
exclude_theoretical = true
exclude_single_case = true
include_types = ["empirical_study", "sample_study"]

[filters.llm.1]
provider = "OpenAI"
api_key = ""
model = "gpt-4o-mini"
temperature = 0.01
```

### Example 4: Minimal Filtering for Broad Inclusion

**Scenario**: Keep most manuscripts, only remove obvious duplicates.

```toml
[project]
name = "Broad Literature Search"

[filters.deduplication]
enabled = true
use_ai = false
compare_fields = ["doi"]  # Only exact DOI matches

[filters.language]
enabled = false  # Accept all languages

[filters.article_type]
enabled = false  # Accept all article types
```

Classification uses multiple indicators:
- Keywords and phrases
- Document structure
- Section headings
- Statistical content
- Length analysis

## Input and Output Formats

### Input Formats

#### CSV Format
```csv
doi,title,abstract,full_text_path
10.1234/example1,"Study Title 1","Abstract text...","./texts/paper1.txt"
10.1234/example2,"Study Title 2","Abstract text...","./texts/paper2.txt"
```

#### TSV/TXT Format
```tsv
doi	title	abstract	full_text_path
10.1234/example1	Study Title 1	Abstract text...	./texts/paper1.txt
10.1234/example2	Study Title 2	Abstract text...	./texts/paper2.txt
```

### Output Formats

#### CSV Output
Includes original columns plus:
- `tag_is_duplicate`: Boolean indicating duplication
- `tag_duplicate_of`: ID of original if duplicate
- `tag_detected_language`: Detected language code
- `tag_article_type`: Classified article type
- `include`: Boolean for inclusion/exclusion
- `exclusion_reason`: Reason if excluded

#### JSON Output
```json
{
  "total_records": 100,
  "included_records": 75,
  "excluded_records": 25,
  "records": [
    {
      "id": "10.1234/example1",
      "original_data": {...},
      "tags": {
        "is_duplicate": false,
        "detected_language": "en",
        "article_type": "research_article"
      },
      "include": true
    }
  ],
  "statistics": {
    "duplicates_found": 10,
    "language_excluded": 8,
    "article_type_excluded": 7
  }
}
```

## Best Practices

### Data Preparation
1. **Ensure consistent formatting**: Clean data before screening
2. **Include key fields**: Title, abstract, and identifiers at minimum
3. **Use unique identifiers**: DOIs, PMIDs, or custom IDs
4. **Verify file paths**: If using external text files, ensure paths are correct

### Filter Configuration
1. **Start conservative**: Begin with high thresholds and adjust as needed
2. **Order matters**: Filters apply sequentially (dedup → language → type)
3. **Test on subset**: Run on a small sample first to verify settings
4. **Document decisions**: Keep notes on why certain filters were chosen

### Performance Optimization
1. **Batch processing**: Process large datasets in chunks if needed
2. **Local text files**: Store full text locally when possible
3. **API limits**: Configure rate limits to avoid API throttling
4. **Incremental screening**: Save progress and resume if interrupted

### Quality Assurance
1. **Review exclusions**: Manually check a sample of excluded items
2. **Adjust thresholds**: Fine-tune based on false positives/negatives
3. **Multiple passes**: Consider running with different settings
4. **Keep originals**: Always maintain unfiltered backup

## Workflow Integration

The Screening tool fits into the systematic review workflow:

```
1. Literature Search
   ↓
2. Export Results (CSV/TSV)
   ↓
3. **SCREENING TOOL**
   - Deduplication
   - Language filtering
   - Type classification
   ↓
4. Manual Review (reduced set)
   ↓
5. Download Tool (acquire selected papers)
   ↓
6. Convert Tool (PDF/DOCX/HTML to text)
   ↓
7. Review Tool (extract information)
```

### Integration with Other prismAId Tools

1. **After Literature Search**: Screen search results before downloading
2. **Before Download Tool**: Filter to reduce papers to acquire
3. **Before Convert Tool**: Only selected papers need conversion
4. **Before Review Tool**: Ensure only relevant papers are reviewed

### Example Workflow

```bash
# 1. Export search results to CSV
# (from PubMed, Web of Science, etc.)

# 2. Run screening on search results
./prismaid -screening screening_config.toml

# 3. Download only included papers
# (use filtered list from screening output)
./prismaid -download-URL filtered_urls.txt

# 4. Convert downloaded papers to text
./prismaid -convert-pdf ./papers

# 5. Run review on converted texts
./prismaid -project review_config.toml
```

## Troubleshooting

### Common Issues and Solutions

#### Issue: High false positive rate in deduplication
**Solution**: 
- Increase similarity threshold (e.g., from 0.85 to 0.95)
- Use more specific comparison fields
- Switch from fuzzy to exact matching for structured data

#### Issue: Language detection errors
**Solution**:
- Enable AI-based detection for mixed-language documents
- Check text encoding (UTF-8 recommended)
- Ensure sufficient text sample (at least 100 characters)

#### Issue: Incorrect article type classification
**Solution**:
- Review classification rules and indicators
- Use AI-based classification for ambiguous cases
- Manually tag a training set for validation

#### Issue: Memory issues with large datasets
**Solution**:
- Process in smaller batches
- Use file paths instead of embedding full text
- Increase system memory allocation

#### Issue: API rate limits exceeded
**Solution**:
- Configure tpm_limit and rpm_limit in LLM settings
- Use multiple API keys with round-robin
- Implement exponential backoff

### Error Messages

**"text_column 'X' not found in CSV"**
- Verify column name matches exactly (case-sensitive)
- Check for extra spaces in column headers

**"at least one filter must be enabled"**
- Enable at least one screening filter in configuration

**"Could not read file X"**
- Verify file paths are relative to current directory
- Check file permissions

### Performance Tips

1. **For speed**: Use exact matching and rule-based methods
2. **For accuracy**: Use fuzzy/semantic matching and AI assistance
3. **For large datasets**: Use file paths instead of inline text
4. **For reproducibility**: Save configuration files with results

## Advanced Features

### Custom Field Mapping
Map non-standard column names:
```toml
[project]
text_column = "manuscript_abstract"  # Your column name
identifier_column = "paper_id"       # Your ID column
```

### Multi-Language Projects
Accept multiple languages:
```toml
[filters.language]
accepted_languages = ["en", "es", "pt", "fr", "it"]
```

### Ensemble AI Screening
Use multiple models for consensus:
```toml
[filters.llm.1]
provider = "OpenAI"
model = "gpt-4o-mini"

[filters.llm.2]
provider = "GoogleAI"
model = "gemini-1.5-flash"
```

### Detailed Logging
High verbosity for debugging:
```toml
[project]
log_level = "high"  # Saves detailed log file
```

---

For more information on systematic review workflows, see the [Review Tool](review-tool) documentation.