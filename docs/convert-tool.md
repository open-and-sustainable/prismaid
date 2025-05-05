---
title: Convert Tool
layout: default
---

# Convert Tool

---

<details>
<summary><strong>Page Contents</strong></summary>
<ul>
  <li><a href="#purpose-and-capabilities"><strong>Purpose and Capabilities</strong></a>: what the Convert tool does and why it's necessary</li>
  <li><a href="#usage-methods"><strong>Usage Methods</strong></a>: how to use the tool across different platforms and programming languages</li>
  <li><a href="#supported-file-formats"><strong>Supported File Formats</strong></a>: details on each file format the tool can process</li>
  <li><a href="#conversion-process"><strong>Conversion Process</strong></a>: how the tool works and what to expect</li>
  <li><a href="#best-practices"><strong>Best Practices</strong></a>: recommendations for effective conversions</li>
  <li><a href="#limitations-and-considerations"><strong>Limitations and Considerations</strong></a>: important factors to be aware of</li>
  <li><a href="#troubleshooting"><strong>Troubleshooting</strong></a>: solutions to common conversion issues</li>
  <li><a href="#workflow-integration"><strong>Workflow Integration</strong></a>: how the Convert tool fits into your systematic review process</li>
</ul>
</details>

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
```

### Go Package

```go
import "github.com/open-and-sustainable/prismaid"

// Convert files of specified formats in a directory
err := prismaid.Convert("./papers", "pdf,docx,html")
```

### Python Package

```python
import prismaid

# Convert files of specified formats in a directory
prismaid.convert("./papers", "pdf,docx,html")
```

### R Package

```r
library(prismaid)

# Convert files of specified formats in a directory
Convert("./papers", "pdf,docx,html")
```

### Julia Package

```julia
using PrismAId

# Convert files of specified formats in a directory
PrismAId.convert("./papers", "pdf,docx,html")
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

## Best Practices

To achieve optimal conversion results:

1. **Pre-conversion check**:
   - Ensure PDF files are text-based, not scanned images
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
   - Text in images cannot be extracted (including scanned PDF documents)

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
   - **Solution**: Use OCR software to convert image-based PDFs, or manually type/transcribe critical content

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

1. **Literature Identification**:
   - Search databases and identify relevant papers

2. **Literature Acquisition** ([Download Tool](download-tool)):
   - Download papers from Zotero collections or URL lists

3. **Format Conversion** (Convert Tool):
   - Convert downloaded papers to text format for analysis
   - Verify conversion quality before proceeding

4. **Review Configuration**:
   - Set up your review project configuration

5. **Systematic Review** ([Review Tool](review-tool)):
   - Process the converted text files to extract structured information

The Convert tool's output directly feeds into the Review tool, making the quality of conversion a critical factor in the success of your systematic review. Always allocate sufficient time for post-conversion verification to ensure your review is based on accurately extracted text.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
