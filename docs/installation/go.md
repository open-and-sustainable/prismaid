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
The execution functions return a result value summarizing the run alongside the error:

```go
import "github.com/open-and-sustainable/prismaid"

// Run screening on manuscripts
screeningResult, err := prismaid.Screening(tomlConfigString)

// Download papers from Zotero
zoteroResult, err := prismaid.DownloadZotero(zoteroTomlConfigString)

// Download files listed in a text or CSV file
urlResult, err := prismaid.DownloadURLList("urls.txt")

// Convert files to text
convertResult, err := prismaid.Convert(inputDir, "pdf,docx,html", prismaid.ConvertOptions{})

// Run a systematic review
reviewResult, err := prismaid.Review(tomlConfigString)
```

Configuration helpers and protocol conformance:

```go
// Validate a configuration without executing it
// configType is one of "review", "screening", or "zotero"
err := prismaid.ValidateConfig("review", tomlConfigString)

// Generate a configuration programmatically (also GenerateScreeningConfig, GenerateZoteroConfig).
// Compose with ValidateConfig to confirm the result before use.
reviewTOML := prismaid.GenerateReviewConfig(prismaid.ReviewConfigParams{ /* project, LLMs, prompt, review items */ })

// Check a RevAIse review record against a reporting protocol's shapes
report, err := prismaid.CheckConformance(recordJSON, "prisma-2020")
```

Refer to full [documentation on pkg.go.dev](https://pkg.go.dev/github.com/open-and-sustainable/prismaid) for additional details.


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
