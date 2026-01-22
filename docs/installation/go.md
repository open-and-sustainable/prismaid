---
title: Go Package 
layout: default
---

# Go Package

---

## Supported Platforms
- Linux: AMD64, Arm64
- macOS: AMD64, Arm64
- Windows: AMD64, Arm64

## Installation
To add the `prismaid` Go package to your project install it with:
```bash
go get "github.com/open-and-sustainable/prismaid"
```

## Use
Import and use the toolkit in your code:
```go
import "github.com/open-and-sustainable/prismaid"

// Run screening on manuscripts
err := prismaid.Screening(tomlConfigString)

// Download papers from Zotero
err := prismaid.DownloadZoteroPDFs(username, apiKey, collectionName, parentDir)

// Convert files to text
err := prismaid.Convert(inputDir, "pdf,docx,html", prismaid.ConvertOptions{})

// Run a systematic review
err := prismaid.Review(tomlConfigString)
```

Refer to full [documentation on pkg.go.dev](https://pkg.go.dev/github.com/open-and-sustainable/prismaid) for additional details.


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
