---
title: R Package
layout: default
---

# R Package

---

## Supported Platforms
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
zotero_content <- paste(readLines("zotero.toml"), collapse = "\n")
DownloadZotero(zotero_content)

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

# Validate a configuration without executing it
# config type is one of "review", "screening", or "zotero"
ValidateConfig("review", toml_content)

# Check a RevAIse review record against a reporting protocol (returns a JSON string)
report <- CheckConformance(record_json, "prisma-2020")

# Get a protocol's full requirement checklist, to plan a conforming review
guidance <- ProtocolGuidance("prisma-2020")
```

See [Protocol Conformance](../conformance) and [Protocol Guidance](../guidance) for what these do.


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
