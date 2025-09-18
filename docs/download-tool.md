---
title: Download Tool
layout: default
---

# Download Tool

---

<details>
<summary><strong>Page Contents</strong></summary>
<ul>
  <li><a href="#purpose-and-capabilities"><strong>Purpose and Capabilities</strong></a>: what the Download tool does and why it's useful</li>
  <li><a href="#usage-methods"><strong>Usage Methods</strong></a>: how to use the tool across different platforms and programming languages</li>
  <li><a href="#downloading-from-url-lists"><strong>Downloading from URL Lists</strong></a>: instructions for batch downloading papers from URLs</li>
  <li><a href="#zotero-integration"><strong>Zotero Integration</strong></a>: comprehensive guide for using the Zotero Download functionality</li>
  <li><a href="#best-practices"><strong>Best Practices</strong></a>: recommendations for effective use of the Download tool</li>
  <li><a href="#troubleshooting"><strong>Troubleshooting</strong></a>: solutions to common Download tool issues</li>
  <li><a href="#workflow-integration"><strong>Workflow Integration</strong></a>: how the Download tool fits into your systematic review process</li>
</ul>
</details>

---

## Purpose and Capabilities

The prismAId Download tool is designed to streamline the literature acquisition phase of your systematic review. It offers two primary functionalities:

1. **URL-based Downloads**: Batch download papers from a list of URLs for efficient collection of literature.
2. **Zotero Integration**: Direct access to papers stored in your Zotero library or group collections.

This tool bridges the gap between literature identification and analysis by automating the tedious process of gathering papers, allowing you to focus on the review itself rather than manual collection tasks.

## Usage Methods

The Download tool can be accessed through multiple interfaces to fit your preferred workflow:

### Binary (Command Line)

```bash
# For URL list downloads
./prismaid -download-URL path/to/urls.txt

# For Zotero downloads (requires a TOML config file)
# First create a file zotero_config.toml with:
#   user = "your_username"
#   api_key = "your_api_key"
#   group = "Your Collection"
./prismaid -download-zotero zotero_config.toml
```

### Go Package

```go
import "github.com/open-and-sustainable/prismaid"

// Download from URL list
err := prismaid.DownloadURLList("path/to/urls.txt")

// Download from Zotero
err := prismaid.DownloadZoteroPDFs("username", "apiKey", "collectionName", "./papers")
```

### Python Package

```python
import prismaid

# Download from URL list
prismaid.download_url_list("path/to/urls.txt")

# Download from Zotero
prismaid.download_zotero_pdfs("username", "api_key", "collection_name", "./papers")
```

### R Package

```r
library(prismaid)

# Download from URL list
DownloadURLList("path/to/urls.txt")

# Download from Zotero
DownloadZoteroPDFs("username", "api_key", "collection_name", "./papers")
```

### Julia Package

```julia
using PrismAId

# Download from URL list
PrismAId.download_url_list("path/to/urls.txt")

# Download from Zotero
PrismAId.download_zotero_pdfs("username", "api_key", "collection_name", "./papers")
```

## Downloading from URL Lists

The URL list download feature allows you to batch download papers from a text file containing URLs, one per line.

### Creating Your URL List

1. Create a simple text file (e.g., `paper_urls.txt`)
2. Add one paper URL per line, for example:
   ```
   https://arxiv.org/pdf/2303.08774.pdf
   https://www.science.org/doi/pdf/10.1126/science.1236498
   https://www.nature.com/articles/s41586-021-03819-2.pdf
   ```
3. Save the file

### Running the Download

Using your preferred method from the [Usage Methods](#usage-methods) section, point the tool to your URL list. For example, with the binary:

```bash
./prismaid -download-URL paper_urls.txt
```

### Output

Papers will be downloaded to the current working directory by default. Each paper is saved with a filename derived from its URL.

## Zotero Integration

The Zotero integration allows direct access to papers stored in your Zotero library or group collections.

### Getting Your Zotero Credentials

1. **Find Your User ID**:
   - Go to the [Zotero Settings](https://www.zotero.org/settings) page
   - Navigate to the **Security** tab, then to the **Applications** section
   - Your **user ID** is displayed at the top:

   <div style="text-align: center;">
       <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/zotero_user.png" alt="Zotero User ID" style="width: 600px;">
   </div>

2. **Generate an API Key**:
   - Click "Create new private key"
   - **Enable** "Allow library access"
   - Set **permissions** to "Read Only" for all groups under "Default Group Permissions"
   - Provide a name for the key, such as "prismAId"
   - Click "Save Key" and copy your new API key

   <div style="text-align: center;">
       <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/zotero_apikey.png" alt="Zotero API Key" style="width: 600px;">
   </div>

### Specifying Collections and Groups

The collection parameter uses a filesystem-like representation for your Zotero library structure:

- For a top-level collection: `"Collection Name"`
- For a parent collection with a sub-collection: `"Parent Collection/Sub Collection"`
- For a group with a collection: `"Group Name/Collection Name"`

### Creating a Zotero Config File

For the binary method, create a TOML configuration file (e.g., `zotero_config.toml`):

```toml
user = "12345678"  # Your Zotero user ID
api_key = "AbCdEfGhIjKlMnOpQrStUv"  # Your Zotero API key
group = "Systematic Review/Climate Papers"  # Your collection path
```

### Running the Download

Choose your preferred method from the [Usage Methods](#usage-methods) section. For example, with the binary:

```bash
./prismaid -download-zotero zotero_config.toml
```

Or with Python:

```python
import prismaid
prismaid.download_zotero_pdfs("12345678", "AbCdEfGhIjKlMnOpQrStUv", "Systematic Review/Climate Papers")
```

## Best Practices

To get the most out of the Download tool:

1. **Organize before downloading**:
   - When using Zotero, organize papers into collections based on their relevance to your review
   - For URL lists, verify all URLs are accessible before batch downloading

2. **Check for duplicates**:
   - Zotero can help identify duplicate entries before downloading
   - Use consistent file naming in URL lists to avoid duplicate downloads

3. **Verify accessibility**:
   - Ensure you have access rights to all papers before downloading
   - Some journals may require institutional access or subscriptions

4. **Structure your downloads**:
   - Use separate output directories for different paper categories
   - Consider naming conventions that will help with the next workflow steps

5. **Batch processing**:
   - For large reviews, consider downloading in batches to manage resources
   - This approach also allows for quality checks along the way

## Troubleshooting

### Common Issues with URL Downloads

- **Invalid URLs**: Ensure all URLs in your list are properly formatted and accessible
- **Access Restrictions**: Some papers may require login credentials or institutional access
- **Network Issues**: Check your internet connection if downloads fail consistently
- **Timeout Errors**: Large files may cause timeout issues; try increasing timeout settings if available

### Common Issues with Zotero Integration

- **Authentication Errors**:
  - Verify your user ID and API key are correct
  - Ensure the API key has appropriate permissions
  - Check that the API key hasn't expired

- **Collection Not Found**:
  - Verify the collection/group path exists and uses the correct format
  - Collection names are case sensitive
  - Check for typos in the collection path

- **Missing PDFs in Zotero**:
  - Some entries in Zotero may not have attached PDFs
  - The tool can only download papers that have PDF attachments in Zotero

- **Rate Limiting**:
  - Zotero API has rate limits; if you receive 429 errors, slow down your requests
  - For large collections, consider downloading in smaller batches

## Workflow Integration

The Download tool is designed to fit seamlessly into your systematic review workflow:

1. **Literature Search**:
   - Search databases and identify potentially relevant papers
   - Export search results to CSV or reference manager

2. **Screening** ([Screening Tool](screening-tool)):
   - Filter out duplicates, wrong languages, and irrelevant article types
   - Create a refined list of papers to acquire

3. **Literature Acquisition** (Download Tool):
   - Download only the screened papers from Zotero collections or URL lists
   - Organize papers in a structured directory

4. **Format Conversion** ([Convert Tool](convert-tool)):
   - Convert downloaded papers to text format for analysis

5. **Systematic Review** ([Review Tool](review-tool)):
   - Process papers and extract structured information

By automating the literature acquisition step with the Download tool, you can significantly reduce the time and effort required for systematic reviews while ensuring a comprehensive and well-organized collection of literature.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
