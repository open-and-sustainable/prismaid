# Complete prismAId Workflow Example

This example demonstrates the complete systematic review workflow using all prismAId tools in the correct order.

## Workflow Overview

The standard systematic review workflow follows these steps:
1. **Literature Search** → 2. **Screening** → 3. **Download** → 4. **Convert** → 5. **Review**

## Step 1: Literature Search

Export your search results from databases (PubMed, Web of Science, Scopus, etc.) to a CSV file:

```csv
doi,title,abstract,authors,year,journal
10.1234/2023.001,"Climate Change Impacts on Biodiversity","This study examines...",Smith et al.,2023,Nature
10.1234/2023.002,"Duplicate: Climate Change Impacts on Biodiversity","This study examines...",Smith et al.,2023,Nature
10.1234/2023.003,"Review of Climate Models","This systematic review...",Jones et al.,2023,Science
10.1234/2023.004,"Editorial: Climate Action Now","Editorial discussing...",Editor,2023,Climate Journal
10.1234/2023.005,"Efectos del Cambio Climático","Este estudio analiza...",García et al.,2023,Revista Clima
```

Save this as `search_results.csv`.

## Step 2: Screening

Create a screening configuration file `screening_config.toml`:

```toml
[project]
name = "Climate Research Screening"
author = "Research Team"
version = "1.0"
input_file = "./search_results.csv"
output_file = "./screened_papers"
text_column = "abstract"
identifier_column = "doi"
output_format = "csv"
log_level = "medium"

[filters]

[filters.deduplication]
enabled = true
method = "fuzzy"           # Use fuzzy matching to catch near-duplicates
threshold = 0.90           # 90% similarity threshold
compare_fields = ["title", "abstract"]

[filters.language]
enabled = true
accepted_languages = ["en"]  # Only accept English papers
use_ai = false               # Use rule-based detection

[filters.article_type]
enabled = true
exclude_reviews = true       # Exclude review articles
exclude_editorials = true    # Exclude editorials
exclude_letters = true       # Exclude letters to editor
include_types = []           # Accept all other types
```

Run the screening:

```bash
./prismaid -screening screening_config.toml
```

Expected output:
- `screened_papers.csv` with tagged and filtered results
- Summary showing excluded papers (duplicates, non-English, reviews/editorials)

## Step 3: Download

### Option A: Download from URL list

Extract URLs from screened papers and create `paper_urls.txt`:

```text
https://example.com/paper1.pdf
https://example.com/paper2.pdf
https://example.com/paper3.pdf
```

Download papers:

```bash
./prismaid -download-URL paper_urls.txt
```

### Option B: Download from Zotero

Create `zotero_config.toml`:

```toml
user = "your_zotero_username"
api_key = "your_api_key_here"
group = "Climate Research Collection"
```

Download papers:

```bash
./prismaid -download-zotero zotero_config.toml
```

## Step 4: Convert

Convert downloaded PDFs to text:

```bash
./prismaid -convert-pdf ./papers
```

For mixed formats:

```bash
./prismaid -convert-pdf ./papers
./prismaid -convert-docx ./papers
./prismaid -convert-html ./papers
```

This creates `.txt` files alongside the original documents.

## Step 5: Review

Create review configuration `review_config.toml`:

```toml
[project]
name = "Climate Impact Information Extraction"
author = "Research Team"
version = "1.0"

[project.configuration]
input_directory = "./papers"  # Directory with .txt files
results_file_name = "./results/climate_review"
output_format = "csv"
log_level = "medium"
duplication = "no"
cot_justification = "yes"  # Include AI reasoning
summary = "no"

[project.llm]
[project.llm.1]
provider = "OpenAI"
api_key = ""  # Uses environment variable if empty
model = "gpt-4o-mini"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

[prompt]
persona = "You are an expert climate scientist conducting a systematic review."
task = "Extract key information about climate change impacts from the attached research paper."
expected_result = "Provide a JSON object with the following information:"
definitions = "Impact severity should be classified as 'low', 'medium', or 'high' based on the magnitude of effects described."
example = ""
failsafe = "If information is not clearly stated in the document, respond with an empty '' value."

[review]
[review.1]
key = "study_location"
values = [""]

[review.2]
key = "climate_factor"
values = ["temperature", "precipitation", "extreme_weather", "sea_level", "other"]

[review.3]
key = "ecosystem_type"
values = ["forest", "marine", "freshwater", "arctic", "desert", "urban", "agricultural", "other"]

[review.4]
key = "impact_severity"
values = ["low", "medium", "high", ""]

[review.5]
key = "time_horizon"
values = ["short_term", "medium_term", "long_term", ""]

[review.6]
key = "adaptation_measures"
values = [""]

[review.7]
key = "mitigation_strategies"
values = [""]
```

Run the review:

```bash
./prismaid -project review_config.toml
```

## Complete Workflow Script

Create a bash script `run_complete_workflow.sh`:

```bash
#!/bin/bash

echo "Starting systematic review workflow..."

# Step 1: Assume search_results.csv exists
echo "Step 1: Literature search results in search_results.csv"

# Step 2: Screening
echo "Step 2: Screening manuscripts..."
./prismaid -screening screening_config.toml

# Step 3: Download (example with URL list)
echo "Step 3: Downloading papers..."
# Extract URLs from screened results (custom script needed)
./prismaid -download-URL paper_urls.txt

# Step 4: Convert
echo "Step 4: Converting papers to text..."
./prismaid -convert-pdf ./papers

# Step 5: Review
echo "Step 5: Running systematic review..."
./prismaid -project review_config.toml

echo "Workflow complete! Check results directory for output."
```

## Expected Results

### After Screening (screened_papers.csv):
- Original papers: 5
- Excluded duplicates: 1
- Excluded non-English: 1
- Excluded reviews/editorials: 2
- Papers for download: 1

### After Review (climate_review.csv):
- Structured data extraction from each paper
- Columns for each review item
- Chain-of-thought justifications (if enabled)
- Ready for statistical analysis

## Tips for Success

1. **Quality Control**: Manually review a sample of screened exclusions
2. **Incremental Processing**: Test with small batches first
3. **API Management**: Monitor API usage and costs
4. **Documentation**: Keep configuration files with results for reproducibility
5. **Backup**: Always maintain original search results and downloaded papers

## Troubleshooting

### Common Issues:

**Screening issues:**
- Adjust similarity thresholds based on your dataset
- Check text encoding for non-English detection
- Review article type classifications manually

**Download failures:**
- Verify URLs are accessible
- Check Zotero API permissions
- Ensure sufficient disk space

**Conversion problems:**
- Some PDFs may have text extraction issues
- Check converted files before review
- Consider OCR for scanned documents

**Review errors:**
- Verify text files exist and are readable
- Check API keys and rate limits
- Start with smaller models for testing

## Integration with Analysis Tools

Export results for further analysis:

### Python
```python
import pandas as pd

# Load screening results
screening_df = pd.read_csv('screened_papers.csv')
excluded = screening_df[screening_df['include'] == False]
print(f"Excluded papers: {len(excluded)}")

# Load review results
review_df = pd.read_csv('climate_review.csv')
print(f"Reviewed papers: {len(review_df)}")

# Analysis example
severity_counts = review_df['impact_severity'].value_counts()
print("Impact severity distribution:", severity_counts)
```

### R
```r
library(tidyverse)

# Load results
screening_results <- read_csv("screened_papers.csv")
review_results <- read_csv("climate_review.csv")

# Analyze exclusions
exclusion_reasons <- screening_results %>%
  filter(include == FALSE) %>%
  count(exclusion_reason)

# Analyze extracted information
impact_summary <- review_results %>%
  group_by(ecosystem_type, impact_severity) %>%
  summarise(count = n())
```

## Conclusion

This workflow demonstrates how prismAId tools work together to streamline systematic reviews:
- **Screening** reduces the corpus to relevant papers
- **Download** efficiently acquires selected papers
- **Convert** prepares papers for AI analysis
- **Review** extracts structured information

The modular design allows you to use tools independently or as a complete pipeline, adapting to your specific research needs.