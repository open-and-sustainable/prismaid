---
title: R Package
layout: default
---

# R Package

---

## Supported Platform
- Linux: AMD64
- macOS: Arm64

## Installation
Install the `prismaid` R package from [R-universe](https://open-and-sustainable.r-universe.dev/prismaid) using:
```r
install.packages("prismaid", repos = c("https://open-and-sustainable.r-universe.dev", "https://cloud.r-project.org"))
```

## Use
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

# Convert with Tika OCR fallback
Convert("./papers", "pdf", "localhost:9998")

# OCR-only for PDFs
Convert("./papers", "pdf", "localhost:9998", "", TRUE)

# Run a systematic review
toml_content <- paste(readLines("project.toml"), collapse = "\n")
RunReview(toml_content)  # Note the capitalization
```


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
