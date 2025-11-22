---
title: Setup Overview
layout: default
---

# Setup Overview

---

## Supported Systems
prismAId is accessible across multiple platforms, offering flexibility based on user preference and system requirements:

1. **Binaries**: Standalone executables for Windows, macOS, and Linux, requiring no coding skills.

2. **Go Package**: Full functionality for Go-based projects.

3. **Python Package** on PyPI: For integration in Python scripts and Jupyter notebooks.

4. **R Package** on R-universe: Compatible with R and RStudio environments.

5. **Julia Package** from the Julia official package registry: For integration in Julia workflows and Jupyter notebooks.

## Toolkit Overview

prismAId offers several specialized tools to support systematic reviews:

1. **Screening Tool**: Filter manuscripts to identify items for exclusion
   - Remove duplicates using exact or fuzzy matching
   - Filter by language detection
   - Classify and filter by article type (research, review, editorial, etc.)

2. **Review Tool**: Process systematic literature reviews based on TOML configurations
   - Configure review criteria, AI model settings, and output formats
   - Extract structured information from scientific papers
   - Generate comprehensive review summaries

3. **Download Tool**: Acquire papers for your review
   - Download PDFs directly from Zotero collections
   - Download files from URL lists

4. **Convert Tool**: Transform documents into analyzable text
   - Convert PDFs, DOCX, and HTML files to plain text
   - Prepare documents for AI processing

### Workflow Overview
1. **AI Model Provider Account and API Key**:
    - Register for an account with [OpenAI](https://www.openai.com/), [GoogleAI](https://aistudio.google.com), [Cohere](https://cohere.com/), or [Anthropic](https://www.anthropic.com/) and obtain an API key.
    - Generate an API key from the provider's dashboard.
2. **Install prismAId**:
    - Follow the installation instructions below based on your preferred system.
3. **Prepare Papers for Review:**
    - Download papers using the Download tool
    - Convert papers to text format using the Convert tool
4. **Define and Run the Review Project:**
    - Set up a configuration file (.toml) for your review project
    - Use the Review tool to process your papers and extract information

### Option 1. Go Package

**(Supported: Linux, macOS, Windows; AMD64, Arm64)**

To add the `prismaid` Go package to your project:
1. Install with:
```bash
go get "github.com/open-and-sustainable/prismaid"
```

2. Import and use the toolkit in your code:
```go
import "github.com/open-and-sustainable/prismaid"

// Run screening on manuscripts
err := prismaid.Screening(tomlConfigString)

// Download papers from Zotero
err := prismaid.DownloadZoteroPDFs(username, apiKey, collectionName, parentDir)

// Convert files to text
err := prismaid.Convert(inputDir, "pdf,docx,html")

// Run a systematic review
err := prismaid.Review(tomlConfigString)
```

Refer to full [documentation on pkg.go.dev](https://pkg.go.dev/github.com/open-and-sustainable/prismaid) for additional details.


### Option 3. Python Package

**(Supported: Linux and Windows AMD64, macOS Arm64)**

Install the `prismaid` package from [PYPI](https://pypi.org/project/prismaid/) with:
```bash
pip install prismaid
```

This Python package provides access to all prismAId tools:
```python
import prismaid

# Run screening on manuscripts
with open("screening.toml", "r") as file:
    screening_config = file.read()
prismaid.screening(screening_config)

# Download papers from Zotero
prismaid.download_zotero_pdfs("username", "api_key", "collection_name", "./papers")  # Full name

# Download from URL list
prismaid.download_url_list("urls.txt")

# Convert files to text
prismaid.convert("./papers", "pdf,docx,html")

# Run a systematic review
with open("project.toml", "r") as file:
    toml_config = file.read()
prismaid.review(toml_config)
```

### Option 4. R Package

**(Supported: Linux AMD64, macOS Arm64)**

Install the `prismaid` R package from [R-universe](https://open-and-sustainable.r-universe.dev/prismaid) using:
```r
install.packages("prismaid", repos = c("https://open-and-sustainable.r-universe.dev", "https://cloud.r-project.org"))
```

Access all prismAId tools from R:
```r
library(prismaid)

# Run screening on manuscripts
screening_content <- paste(readLines("screening.toml"), collapse = "\n")
Screening(screening_content)  # Note the capitalization


# Download papers from Zotero
DownloadZoteroPDFs("username", "api_key", "collection_name", "./papers")  # Full name

# Download from URL list
DownloadURLList("urls.txt")

# Convert files to text
Convert("./papers", "pdf,docx,html")  # Note the capitalization

# Run a systematic review
toml_content <- paste(readLines("project.toml"), collapse = "\n")
RunReview(toml_content)  # Note the capitalization
```

### Option 5. Julia Package

**(Supported: Linux and Windows AMD64, macOS Arm64)**

Install the `PrismAId` package using Julia's package manager:
```julia
using Pkg
Pkg.add("PrismAId")
```

Access all prismAId tools from Julia:
```julia
using PrismAId

# Run screening on manuscripts
screening_config = read("screening.toml", String)
PrismAId.screening(screening_config)

# Download papers from Zotero
PrismAId.download_zotero_pdfs("username", "api_key", "collection_name", "./papers")  # Full name

# Download from URL list
PrismAId.download_url_list("urls.txt")

# Convert files to text
PrismAId.convert("./papers", "pdf,docx,html")

# Run a systematic review
toml_config = read("project.toml", String)
PrismAId.run_review(toml_config)  # Correct function name
```


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
