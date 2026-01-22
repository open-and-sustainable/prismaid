---
title: Convert Tool
layout: default
---

# Convert Tool

---

## Purpose and Capabilities

The prismAId Convert tool transforms documents from their native formats into plain text files that can be processed by Large Language Models (LLMs). This critical step bridges the literature acquisition phase and the systematic review analysis by:

1. **Standardizing formats**: Converting various document types into a consistent plain text format
2. **Extracting content**: Pulling textual content from complex formatted documents
3. **Preparing for analysis**: Creating files that are optimized for LLM processing

The tool currently supports three main document formats: PDF, DOCX, and HTML, making it versatile for handling various sources of scientific literature.

## Usage Methods

The Convert tool can be accessed through multiple interfaces to accommodate different workflows:

### Binary (Command Line)

```bash
# Convert all PDFs in a directory
./prismaid -convert-pdf ./papers

# Convert all DOCX files in a directory
./prismaid -convert-docx ./papers

# Convert all HTML files in a directory
./prismaid -convert-html ./papers

# Convert with Tika OCR fallback for difficult files
./prismaid -convert-pdf ./papers -tika-server localhost:9998

# Convert a single PDF (PDF-only flag)
./prismaid -convert-pdf ./papers -single-file /path/to/file.pdf

# OCR-only for a single PDF (requires Tika server)
./prismaid -convert-pdf ./papers -single-file /path/to/file.pdf -ocr-only -tika-server localhost:9998
```

### Go Package

```go
import "github.com/open-and-sustainable/prismaid/conversion"

// Convert files of specified formats in a directory (no OCR fallback)
err := conversion.Convert("./papers", "pdf,docx,html", conversion.ConvertOptions{})

// Convert with Tika OCR fallback
err := conversion.Convert("./papers", "pdf,docx,html", conversion.ConvertOptions{
    TikaServer: "localhost:9998",
})

// PDF-specific options (single file and OCR-only)
err := conversion.Convert("./papers", "pdf", conversion.ConvertOptions{
    TikaServer: "localhost:9998",
    PDF: conversion.PDFOptions{
        SingleFile: "/path/to/file.pdf",
        OCROnly:    true,
    },
})
```

### Python Package

```python
import prismaid

# Convert files of specified formats in a directory (no OCR fallback)
prismaid.convert("./papers", "pdf,docx,html")

# Convert with Tika OCR fallback - pass server address as optional parameter
prismaid.convert("./papers", "pdf,docx,html", "localhost:9998")

# OCR-only for PDFs
prismaid.convert("./papers", "pdf", "localhost:9998", ocr_only=True)
```

### R Package

```r
library(prismaid)

# Convert files of specified formats in a directory (no OCR fallback)
Convert("./papers", "pdf,docx,html")

# Convert with Tika OCR fallback - pass server address as optional parameter
Convert("./papers", "pdf,docx,html", "localhost:9998")

# OCR-only for PDFs
Convert("./papers", "pdf", "localhost:9998", "", TRUE)
```

### Julia Package

```julia
using PrismAId

# Convert files of specified formats in a directory (no OCR fallback)
PrismAId.convert("./papers", "pdf,docx,html")

# Convert with Tika OCR fallback - pass server address as optional parameter
PrismAId.convert("./papers", "pdf,docx,html", "localhost:9998")

# OCR-only for PDFs
PrismAId.convert("./papers", "pdf", "localhost:9998", "", true)
```

## Supported File Formats

### PDF (.pdf)

PDF (Portable Document Format) is the most common format for published scientific papers. The Convert tool uses advanced text extraction techniques to handle complex PDF structures:

- **Text Elements**: Extracts main body text, headings, and captions
- **Text Flow**: Attempts to maintain proper reading order
- **Multi-Column Handling**: Processes papers with multiple column layouts
- **Basic Table Detection**: Attempts to preserve tabular data

**Limitations**: Due to the nature of PDFs, which are essentially digital printouts, text extraction can be imperfect. Some formatting, mathematical equations, and specialized symbols may not convert accurately.

### DOCX (.docx)

Microsoft Word documents (.docx) are common for manuscripts in development or preprints:

- **Text Extraction**: Preserves most text formatting and structure
- **List Handling**: Maintains numbered and bulleted lists
- **Table Support**: Extracts content from tables
- **Document Structure**: Preserves headings and document organization

**Limitations**: Some complex formatting elements like text boxes or embedded objects may not convert perfectly.

### HTML (.html)

HTML files are often used for web-published articles or open-access content:

- **Text Content**: Extracts main article content
- **Structural Elements**: Preserves headings and sectioning
- **List Elements**: Maintains ordered and unordered lists
- **Basic Table Support**: Extracts tabular data

**Limitations**: Dynamic content, JavaScript-generated text, or complex layouts may not be fully captured.

## Conversion Process

The Convert tool follows a standardized process:

1. **File Discovery**: The tool scans the specified directory for files of the requested format(s)
2. **Content Extraction**: For each file, the appropriate extraction method is applied based on file type
3. **Text Processing**: Extracted text is processed to remove unnecessary elements and normalize formatting
4. **Output Generation**: A plain text (.txt) file is created for each input document, maintaining the same filename but with a .txt extension

### Notes on Process Isolation and OCR Retry

- The CLI `-convert-pdf` command runs per-file conversions in separate processes. If a single-file conversion fails or produces a zero-byte `.txt`, it will retry once using OCR-only when Tika is available.
- Library bindings (Go package, Python/Julia/R) run in-process and do not spawn isolated child processes. For those integrations, handle retries or post-checks in your own workflow if needed.

## OCR Fallback with Apache Tika

For challenging documents that fail standard conversion methods (such as scanned PDFs, image-based files, or corrupted documents), prismAId offers an optional OCR (Optical Character Recognition) fallback using Apache Tika.

### What is Apache Tika?

Apache Tika is a powerful content analysis toolkit that can extract text from over a thousand different file types. When configured with Tesseract OCR, it can:

- Extract text from scanned PDF documents
- Process image-based files (PNG, JPEG, TIFF, etc.)
- Handle documents that fail with standard extraction methods
- Provide more robust text extraction for complex or corrupted files

### Setting Up Tika Server

prismAId includes a helper script (`tika-service.sh` in the repository root) to start a local Tika server using Docker or Podman:

```bash
# Start Tika server with OCR support
./tika-service.sh start

# Check server status
./tika-service.sh status

# View server logs
./tika-service.sh logs

# Stop the server
./tika-service.sh stop
```

The server will be available at `http://localhost:9998` by default.

**Alternative Setup with Docker:**

If you prefer to manage the container manually:

```bash
# Pull and run Tika server with OCR support
docker run -d -p 9998:9998 --name tika-ocr apache/tika:latest-full

# Or with Podman
podman run -d -p 9998:9998 --name tika-ocr apache/tika:latest-full
```

### Using OCR Fallback

Once the Tika server is running, enable OCR fallback by providing the server address:

**Command Line:**
```bash
# Basic usage with single format
./prismaid -convert-pdf ./papers -tika-server localhost:9998

# Note: You can only convert one format at a time via CLI
# To convert multiple formats, run the command multiple times:
./prismaid -convert-pdf ./papers -tika-server localhost:9998
./prismaid -convert-docx ./papers -tika-server localhost:9998
./prismaid -convert-html ./papers -tika-server localhost:9998
```

**Go Package:**
```go
import "github.com/open-and-sustainable/prismaid/conversion"

// Convert single format with Tika OCR fallback
err := conversion.Convert("./papers", "pdf", conversion.ConvertOptions{
    TikaServer: "localhost:9998",
})

// Convert multiple formats in one call
err := conversion.Convert("./papers", "pdf,docx,html", conversion.ConvertOptions{
    TikaServer: "localhost:9998",
})

// Disable Tika by leaving TikaServer empty
err := conversion.Convert("./papers", "pdf", conversion.ConvertOptions{})

// OCR-only for a single PDF
err := conversion.Convert("./papers", "pdf", conversion.ConvertOptions{
    TikaServer: "localhost:9998",
    PDF: conversion.PDFOptions{
        SingleFile: "/path/to/file.pdf",
        OCROnly:    true,
    },
})
```

**Python Package:**
```python
import prismaid

# Convert single format with Tika OCR fallback
prismaid.convert("./papers", "pdf", "localhost:9998")

# Convert multiple formats
prismaid.convert("./papers", "pdf,docx,html", "localhost:9998")

# Disable Tika by passing empty string (default)
prismaid.convert("./papers", "pdf", "")
```

**R Package:**
```r
library(prismaid)

# Convert single format with Tika OCR fallback
Convert("./papers", "pdf", "localhost:9998")

# Convert multiple formats
Convert("./papers", "pdf,docx,html", "localhost:9998")

# Disable Tika by passing empty string (default)
Convert("./papers", "pdf", "")
```

**Julia Package:**
```julia
using PrismAId

# Convert single format with Tika OCR fallback
PrismAId.convert("./papers", "pdf", "localhost:9998")

# Convert multiple formats
PrismAId.convert("./papers", "pdf,docx,html", "localhost:9998")

# Disable Tika by passing empty string (default)
PrismAId.convert("./papers", "pdf", "")
```

### How OCR Fallback Works

When you specify a Tika server address:

1. **Server Availability Check**: prismAId checks if the Tika server is reachable
2. **Primary Conversion**: prismAId first attempts standard conversion methods (fast, no network)
   - For PDFs: tries `ledongthuc/pdf` library, then falls back to `pdfcpu` library
   - For DOCX: uses `go-docx` library
   - For HTML: uses `html2text` library
3. **Automatic Fallback**: If standard methods fail (error) OR return empty text, AND Tika is available, the file is automatically sent to the Tika server
4. **OCR Processing**: Tika performs OCR on the document if needed (scanned PDFs, images, etc.)
5. **Text Extraction**: Extracted text is saved as a .txt file

### OCR-Only Full Directory Example

Use this when you want to force OCR for every PDF in a directory (no standard PDF extraction):

**Command Line:**
```bash
./prismaid -convert-pdf ./papers -ocr-only -tika-server localhost:9998
```

**Go Package:**
```go
import "github.com/open-and-sustainable/prismaid/conversion"

err := conversion.Convert("./papers", "pdf", conversion.ConvertOptions{
    TikaServer: "localhost:9998",
    PDF: conversion.PDFOptions{
        OCROnly: true,
    },
})
```

The fallback is transparent - you'll see log messages indicating when Tika is being used:
```
Tika server available at localhost:9998 - will use as OCR fallback
Standard conversion failed for scanned.pdf, attempting Tika OCR fallback
Successfully extracted text from scanned.pdf using Tika OCR
```

**Graceful Degradation**: If you specify a Tika server but it's not available, prismAId will log an info message and continue with standard conversion only - it won't fail.

### When to Use OCR Fallback

Consider using the Tika OCR fallback when:

- Working with older publications that may be scanned images
- Dealing with documents from varied sources with inconsistent quality
- Processing documents that failed standard conversion
- Handling image-based files alongside text-based documents
- Working with non-standard or corrupted PDF files

### Performance Considerations

OCR processing is computationally intensive:

- **Processing Time**: OCR can take 10-60 seconds per page depending on image quality and page complexity
- **Server Resources**: The Tika server requires approximately 2-4 GB of RAM
- **Network**: Using a local Tika server (`localhost:9998`) is much faster than remote servers

**Recommendation**: Start the Tika server before processing large batches of documents, and leave it running for your entire session.

### Tika Server Address Format

The Tika server address should be in the format `host:port` (without `http://`):

**Valid formats:**
- `localhost:9998` (most common for local server)
- `127.0.0.1:9998`
- `0.0.0.0:9998`
- `192.168.1.100:9998` (remote server on your network)
- `server.local:9998`

**Invalid formats (will not work):**
- `http://localhost:9998` ❌ (don't include protocol)
- `localhost:9998/tika` ❌ (don't include path)
- `localhost` ❌ (must include port)

## Best Practices

To achieve optimal conversion results:

1. **Pre-conversion check**:
   - Check if PDF files are text-based or scanned images
   - If dealing with scanned documents, set up Tika server for OCR support
   - Verify that documents are not password-protected or damaged
   - Check that files are complete and correctly formatted

2. **Directory organization**:
   - Keep original files and converted text files organized in separate directories
   - Use consistent file naming conventions to maintain traceability

3. **Post-conversion verification**:
   - **IMPORTANT**: Always manually check a sample of converted documents to ensure quality
   - Pay special attention to papers with complex formatting, equations, or non-standard characters
   - Consider spot-checking longer documents to verify that all content was properly extracted

4. **Handling special cases**:
   - For papers with significant mathematical content, consider additional manual editing
   - For papers with important tables or figures, supplementary notes may be needed
   - Non-English papers may require special attention to character encoding

## Limitations and Considerations

**<span style="color: red; font-weight: bold;">IMPORTANT:</span>** The conversion process has inherent limitations that users should be aware of:

1. **PDF Limitations**:
   - PDFs store formatting rather than semantic structure, making perfect extraction challenging
   - Multi-column layouts may occasionally be extracted in incorrect order
   - Figures and their captions may be separated or misplaced
   - Mathematical equations often convert poorly to plain text
   - Headers, footers, and page numbers may appear in the middle of content

2. **Text Recognition Issues**:
   - Non-standard fonts may cause character recognition problems
   - Ligatures and special characters might not be preserved correctly
   - Text in images cannot be extracted without OCR (use Tika fallback for scanned PDFs)

3. **Structural Information Loss**:
   - Formatting that conveys meaning (bold, italic, etc.) is lost in plain text
   - Document hierarchy may not be perfectly preserved
   - References to figures or tables by location ("see Figure 2 below") may lose context

4. **Special Content**:
   - Tables are particularly challenging and may lose their structure
   - Citations and references may not maintain their formatting
   - Footnotes may be displaced from their reference points

## Troubleshooting

### Common Issues and Solutions

1. **Empty or Very Short Output Files**:
   - **Issue**: The conversion produced an empty or minimal text file
   - **Possible causes**:
     - The PDF is a scanned image without text layers
     - The document is corrupt or password-protected
     - The file contains primarily non-textual elements
   - **Solution**: Enable Tika OCR fallback with `-tika-server localhost:9998` - it will automatically trigger when standard methods return empty text

2. **Garbled Text**:
   - **Issue**: Output contains random characters or illegible text
   - **Possible causes**:
     - Non-standard encoding
     - Custom fonts without proper mapping
     - Copy protection mechanisms
   - **Solution**: Try opening the original in different applications and copying text manually, or contact the publisher for an accessible version

3. **Incomplete Conversion**:
   - **Issue**: Only part of the document was converted
   - **Possible causes**:
     - File corruption
     - Complex document structure that confused the parser
   - **Solution**: Try alternative conversion tools or split large documents into smaller sections

4. **Character Encoding Issues**:
   - **Issue**: Special characters appear incorrectly
   - **Possible causes**:
     - Mismatched character encoding
     - Non-standard character sets
   - **Solution**: Manually correct critical passages or try using different encoding options if your programming language interface allows it

## Workflow Integration

The Convert tool is a critical bridge in the systematic review workflow:

1. **Literature Search**:
   - Search databases and identify potentially relevant papers
   - Export search results to CSV or reference manager

2. **Screening** ([Screening Tool](screening-tool)):
   - Filter out duplicates, wrong languages, and irrelevant article types
   - Create a refined list of papers to acquire

3. **Literature Acquisition** ([Download Tool](download-tool)):
   - Download only the screened papers from Zotero collections or URL lists

4. **Format Conversion** (Convert Tool):
   - Convert downloaded papers to text format for analysis
   - Verify conversion quality before proceeding

5. **Systematic Review** ([Review Tool](review-tool)):
   - Process the converted text files to extract structured information

The Convert tool's output directly feeds into the Review tool, making the quality of conversion a critical factor in the success of your systematic review. Always allocate sufficient time for post-conversion verification to ensure your review is based on accurately extracted text.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
