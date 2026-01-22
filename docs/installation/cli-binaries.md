---
title: CLI Binaries
layout: default
---

# CLI Binaries

---

## Supported Platforms
- Linux: AMD64, Arm64
- macOS: AMD64, Arm64
- Windows: AMD64, Arm64

## Installation
Download the appropriate executable for your OS from our [GitHub Releases](https://github.com/open-and-sustainable/prismaid/releases).

## Use
No coding is required. Use the command line interface to access all tools:

```bash
# Screen manuscripts to filter out duplicates and irrelevant papers
./prismaid -screening screening_config.toml

# Download papers from Zotero (requires a TOML config file)
# First create a file zotero_config.toml with:
#   user = "your_username"
#   api_key = "your_api_key"
#   group = "Your Collection"
./prismaid -download-zotero zotero_config.toml

# Download papers from a URL list
./prismaid -download-URL paper_urls.txt

# Convert files to text (separate commands for each format)
./prismaid -convert-pdf ./papers
./prismaid -convert-docx ./papers
./prismaid -convert-html ./papers

# PDF-only options
./prismaid -convert-pdf ./papers -single-file /path/to/file.pdf
./prismaid -convert-pdf ./papers -ocr-only -tika-server localhost:9998

# Initialize a new project configuration interactively
./prismaid -init

# Run a systematic review
./prismaid -project your_project.toml
```


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
