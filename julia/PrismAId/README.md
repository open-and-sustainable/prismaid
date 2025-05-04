# ![prismAId Logo](https://raw.githubusercontent.com/ricboer0/prismAId/main/figures/prismAId_logo.png) PrismAId

`PrismAId` is a Julia package designed to facilitate access to the [prismAId](https://github.com/open-and-sustainable/prismaid) tools directly from Julia code and workflows.

## Installation
To install `PrismAId` using Julia's package manager and official registry, run the following commands in your Julia REPL:
```julia
using Pkg
Pkg.add("PrismAId")
```

## Usage

`PrismAId` provides four main functions to interact with the underlying shared library:

1. `run_review`: Execute a systematic review based on a TOML configuration
2. `download_zotero_pdfs`: Download PDFs from a Zotero collection
3. `download_url_list`: Download files from a list of URLs
4. `convert`: Convert files to text format

### Quick Start Example: Running a Review

1. Start by loading the `PrismAId` package:
   ```julia
   using PrismAId
   ```

2. Prepare your review project configuration in TOML format:
   ```julia
   toml_test = """
       [project]
       name = "Test of prismAId"
       ...
       """
   ```

3. Run the review process:
   ```julia
   PrismAId.run_review(toml_test)
   ```

When you run the review project, you'll be prompted with cost information and asked to confirm before proceeding.

### Downloading PDFs from Zotero

To download PDFs from a Zotero collection:

```julia
# Parameters: username, API key, collection name, destination directory
PrismAId.download_zotero_pdfs(
    "your_username",
    "your_api_key",
    "Collection/Subcollection",
    "/path/to/output/directory"
)
```

### Downloading Files from a URL List

To download files from a list of URLs (one URL per line):

```julia
# Parameter: path to file containing URLs
PrismAId.download_url_list("/path/to/url_list.txt")
```

### Converting Files to Text Format

To convert files from various formats (PDF, DOCX, HTML) to text:

```julia
# Parameters: directory containing files, comma-separated list of formats
PrismAId.convert("/path/to/files", "pdf,docx,html")
```

This will process all files with the specified extensions in the directory and create corresponding .txt files.

## Important Notes

**ATTENTION**: Interaction with `PrismAId` functionalities is mediated through a C shared library, which can make debugging challenging. It is recommended to set the `log_level` to `high` in your project configuration to ensure comprehensive logging of any issues encountered during the review process, with logs stored in the specified output directory.

## Documentation

Comprehensive documentation for `PrismAId`, including detailed descriptions of its functionalities, installation guide, usage examples, and configuration settings, is available online. You can access the complete documentation by visiting the following URL:

[prismAId Documentation](https://open-and-sustainable.github.io/prismaid)

## License
PrismAId is made available under the GNU Affero General Public License v3 (AGPL v3).
