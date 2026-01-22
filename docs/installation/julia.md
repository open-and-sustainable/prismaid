---
title: Julia Package
layout: default
---


# Julia Package

---

## Supported Platforms
- Linux: AMD64
- Windows: AMD64
- macOS: Arm64

## Installation
Install the `PrismAId` package using Julia's package manager:
```julia
using Pkg
Pkg.add("PrismAId")
```

## Use
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

# Convert with Tika OCR fallback
PrismAId.convert("./papers", "pdf", "localhost:9998")

# OCR-only for PDFs
PrismAId.convert("./papers", "pdf", "localhost:9998", "", true)

# Run a systematic review
toml_config = read("project.toml", String)
PrismAId.run_review(toml_config)  # Correct function name
```


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
