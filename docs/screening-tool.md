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
method = "fuzzy"                          # "exact", "fuzzy", or "semantic"
threshold = 0.85                          # Similarity threshold (0.0-1.0)
compare_fields = ["title", "abstract"]    # Fields to compare

[filters.language]
enabled = true
accepted_languages = ["en", "es", "fr"]   # ISO language codes
use_ai = false                            # Use AI for detection

[filters.article_type]
enabled = true
exclude_reviews = true                    # Exclude review articles
exclude_editorials = true                 # Exclude editorials
exclude_letters = true                    # Exclude letters
include_types = []                        # Specific types to include
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

The deduplication filter identifies duplicate manuscripts using three methods:

#### 1. Exact Matching
- Compares specified fields exactly (after normalization)
- Fastest method with 100% precision
- Best for: Structured data with consistent formatting

#### 2. Fuzzy Matching
- Uses string similarity algorithms (Jaccard, Levenshtein)
- Configurable threshold (0.0 to 1.0)
- Best for: Catching near-duplicates with minor variations

#### 3. Semantic Matching (Future Enhancement)
- Uses embeddings for semantic similarity
- Catches conceptually similar manuscripts
- Best for: Identifying paraphrased or translated duplicates

### Language Detection Filter

The language detection filter can operate in two modes:

#### Rule-Based Detection
- Analyzes character scripts (Latin, Cyrillic, CJK, Arabic, etc.)
- Checks for common words in various languages
- Fast and privacy-preserving
- Supports major world languages

#### AI-Based Detection
- Uses LLMs for more accurate detection
- Handles mixed-language documents
- Identifies regional variants
- Requires API configuration

### Article Type Classification Filter

Classifies manuscripts into categories:

- **Research Articles**: Original research with methods and results
- **Review Articles**: Literature reviews, narrative reviews
- **Systematic Reviews**: Following structured protocols
- **Meta-Analyses**: Statistical analysis of multiple studies
- **Editorials**: Opinion pieces by editors
- **Letters**: Correspondence to editors
- **Case Reports**: Individual patient cases
- **Commentary**: Comments on published work
- **Perspectives**: Author viewpoints

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