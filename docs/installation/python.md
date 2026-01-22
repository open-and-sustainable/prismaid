---
title: Python Package
layout: default
---

# Python Package

---

## Supported Platforms
- Linux: AMD64
- Windows: AMD64
- macOS: Arm64

## Installation
Install the `prismaid` package from [PYPI](https://pypi.org/project/prismaid/) with:
```bash
pip install prismaid
```

## Use
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

# Convert with Tika OCR fallback
prismaid.convert("./papers", "pdf", "localhost:9998")

# OCR-only for PDFs
prismaid.convert("./papers", "pdf", "localhost:9998", ocr_only=True)

# Run a systematic review
with open("project.toml", "r") as file:
    toml_config = file.read()
prismaid.review(toml_config)
```

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
