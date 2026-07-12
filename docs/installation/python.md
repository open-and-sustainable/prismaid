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
with open("zotero.toml", "r") as file:
    zotero_config = file.read()
prismaid.download_zotero(zotero_config)

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

# Validate a configuration without executing it
# config type is one of "review", "screening", or "zotero"
prismaid.validate_config("review", toml_config)

# Check a RevAIse review record against a reporting protocol (returns a dict)
report = prismaid.check_conformance(record_json, "prisma-2020")

# Get a protocol's full requirement checklist, to plan a conforming review
guidance = prismaid.protocol_guidance("prisma-2020")
```

See [Protocol Conformance](../conformance) and [Protocol Guidance](../guidance) for what these do.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
