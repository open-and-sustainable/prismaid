# prismAId Changelog
All notable changes to this project will be documented in this file.
Releases use semantic versioning as in 'MAJOR.MINOR.PATCH'.
## Change entries
Added: For new features that have been added.
Changed: For changes in existing functionality.
Deprecated: For once-stable features removed in upcoming releases.
Removed: For features removed in this release.
Fixed: For any bug fixes.
Security: For vulnerabilities.

## [0.11.2] - 2026-02-13
### Added
- CLI PDF conversion retries OCR-only once for per-file errors or zero-byte outputs when Tika is available
### Fixed
- Updated R shared-lib header to match new ConvertR signature used by bindings
- Fixed Julia wrapper FFI pointer handling for exported calls
- Added Julia regression tests for wrapper pointer-conversion safety

## [0.11.1] - 2026-01-22
### Added
- PDF conversion now supports isolated per-file execution to reduce failures from process termination
- Added PDF-only options for conversion: single-file conversion and OCR-only mode (CLI and language bindings)
### Fixed
- Pagination size and limit in Zotero API call by download tool

## [0.11.0] - 2026-01-21
### Changed
- Updated alembica dependency from v0.1.1 to v0.3.0
  - Added support for cloud providers: AWS Bedrock, Azure AI, Vertex AI
  - Added support for self-hosted OpenAI-compatible endpoints
  - Extended configuration schema with optional fields for cloud/local deployments (base_url, endpoint_type, region, project_id, location, api_version)

### Added
- Cloud provider support in Review and Screening tools
  - AWS Bedrock: Configure with endpoint_type="bedrock" and region
  - Azure AI: Configure with endpoint_type="azure", base_url, and api_version
  - Vertex AI: Configure with endpoint_type="vertex", project_id, and location
  - Self-hosted: Configure with provider="SelfHosted" and base_url for OpenAI-compatible endpoints
- Extended LLM configuration in terminal init to support cloud providers with guided prompts
- Extended web review configurator to support cloud providers with dynamic field visibility
- Updated documentation with cloud provider examples and configuration details
- Updated README with cloud provider and self-hosted support in LLM list

## [0.10.1] - 2026-01-20
### Changed
- Updated CI/CD workflows to use macos-latest runner for ARM64 builds
  - Changed from macos-14 to macos-latest (currently macOS 15 ARM64)
  - Note: macos-15-intel is available for Intel builds until August 2027
- Updated alembica dependency from v0.0.8 to v0.1.1
  - **BREAKING**: Migrated from community OpenAI SDK to official OpenAI SDK
  - Updated Anthropic SDK from v1.2.1 to v1.19.0 with breaking changes
  - Updated Cohere SDK from v2.14.1 to v2.16.1

### Added
- Support for Perplexity AI provider (Sonar, Sonar Pro, Sonar Reasoning Pro, Sonar Deep Research)
- Support for OpenAI GPT-5 series models (GPT-5, GPT-5.1, GPT-5.2, GPT-5 Mini, GPT-5 Nano)
- Support for Anthropic Claude 4.5 series models (Claude 4.5 Opus, Claude 4.5 Sonnet, Claude 4.5 Haiku)
- Support for Google AI Gemini 2.5 series models (Gemini 2.5 Pro, Gemini 2.5 Flash, Gemini 2.5 Flash Lite)
- Support for Google AI Gemini 3 series preview models (Gemini 3 Pro Preview, Gemini 3 Flash Preview)
- Support for Cohere Command A Reasoning model (command-a-reasoning-08-2025)
- Apache Tika OCR fallback support in conversion tool for scanned PDFs and image-based documents
  - Automatic fallback to Apache Tika server with Tesseract OCR when standard conversion methods fail or return empty text
  - Optional `tika-server` parameter in CLI and all language bindings (Go, Python, R, Julia)
  - Included `tika-service.sh` script for easy Tika server management with Docker/Podman
  - Graceful degradation when Tika server is unavailable
  - Comprehensive documentation and testing for OCR fallback functionality

## [0.10.0] - 2025-11-22
### Changed
- New documentation domain
- New documentation template release and structure

### Update
- Documentation pages structure and composition to align with changed template structure

## [0.9.6] - 2025-10-03
### Fixed
- Configuration of R package to avoid compilation errors for unsupported CPUs
- Windows r-universe deployment configuration
- Infinite recursive DOI resolutions


## [0.9.5] - 2025-10-03
### Changed
- Configuration for R package compilation and deployment on R universe
- Modified download tool output to report consistent '_download' lists, with success, eventual reason for failure, and filename for stored PDFs

## [0.9.4] - 2025-10-01
### Fixed
- DOI handling consistency across download tool:
  - Fixed inconsistent DOI URL construction in page content extraction, meta tag resolution, and Dimensions URL handling
  - All DOI resolution now properly uses convertDOIToURL() function instead of manual string concatenation
  - DOI cleaning for Unpaywall API integration to remove URL prefixes and ensure clean DOI identifiers
  - Proper handling of DOIs with various prefixes (doi:, DOI:, https://doi.org/, etc.)
  - Ensures reliable DOI resolution whether found in page content, meta tags, or API responses

## [0.9.3] - 2025-10-01
### Added
- Concurrent download functionality with intelligent rate limiting:
  - Global concurrency limit: Maximum 25 concurrent downloads system-wide
  - Per-host concurrency limit: Maximum 4 concurrent requests per publisher/domain
  - Prevents overwhelming individual publishers and reduces 429/403 throttling responses
- Early response validation for efficient bandwidth usage in download tool:
  - Validates Content-Type headers before downloading full response body
  - Checks first 4 bytes for %PDF signature to confirm valid PDF files
  - Aborts quickly on HTML error pages or invalid content
- Intelligent retry policy with exponential backoff in download tool:
  - Automatic retry on transient errors (5xx status codes, timeouts, connection resets)
  - Respects Retry-After headers from servers to avoid aggressive retrying
  - Exponential backoff with jitter (1s, 2s, 4s delays) to prevent retry storms
  - Maximum 3 retry attempts per download with smart error classification
  - Non-retryable errors (4xx client errors) fail immediately
- Unpaywall API integration as final fallback in download tool:
  - When downloads fail, automatically searches Unpaywall database for open access alternatives
  - Finds free, legal versions from 50,000+ publishers and repositories
  - Extracts DOIs from URLs or metadata to query open access database
  - Helps recover papers that are paywalled at original source but freely available elsewhere
- Improved CSV/TSV column detection with content analysis in download tool:
  - Smart distinction between journal titles and database sources in "Source" columns
  - Prioritizes specific column names (SourceTitle, Publication_Title) over generic ones
  - Content analysis prevents misidentifying database names as journal titles

### Changed
- Download tool now processes multiple PDFs concurrently instead of sequentially
- Significantly improved download performance for batch operations
- Enhanced error handling and reporting for failed downloads

## [0.9.2] - 2025-09-30
### Fixed
- Improved download tool features:
  - better format names of downloaded files when possible
  - avoid overwriting existing pdfs
  - direct links download
  - search for DOI in case of JavaScript pages reached form URLs in multiple column lists
  - better report of successes and failures in output list
- Updated documentation and testing accordingly


## [0.9.1] - 2025-09-26
### Fixed
- AI-assisted screening filters now properly respect rate limits by batching all prompts into single API calls
  - Deduplication filter: All manuscript comparisons processed in one batch instead of individual calls
  - Language detection filter: All manuscripts analyzed together instead of sequentially
  - Article type filter: All classifications performed in one batch request
  - Topic relevance filter: All relevance assessments bundled into single API call
- Added 30-second delay between consecutive AI-assisted filters to prevent rate limit breaches
### Changed
- Replaced `interface{}` with `any` throughout screening codebase for modern Go compatibility

## [0.9.0] - 2025-09-24
### Added
- New Screening tool for filtering manuscripts before download
  - Deduplication filter with exact, and semantic matching algorithms
  - Language detection filter with rule-based and AI-assisted detection
  - Article type classification filter (research articles, reviews, editorials, letters, etc. -- based on rules or AI)
  - Off-topic manuscripts detection filter (scores on keywords, concepts, and field, or AI)
  - Support for CSV and TSV input/output formats
  - TOML-based configuration following project patterns
  - Integration with command-line interface via `-screening` flag
  - Full API access through Go, Python, R, and Julia bindings
- Comprehensive documentation for the Screening tool
- Test coverage for screening functionality

### Changed
- Updated workflow order to: Search → Screen → Download → Convert → Review
- Updated documentation to reflect correct systematic review workflow
- Added Screening tool to navigation and all tool references
- Restructured projects directory to properly separate templates, tests, and users' workspace

## [0.8.1] - 2025-05-30
### Changed
- Updated dependencies, including alembica with support for multiple new models

## [0.8.0] - 2025-05-05
### Removed
- Removed Zotero integration and input conversion from TOML project configuration & template
- Removed Zotero integration and input conversion from terminal configuration init
- Removed Zotero integration and input conversion from review logic
- Removed Zotero integration and input conversion from TOML loading logic
- model features and rate limits in documentation, moved to `alembica` project
### Added
- Added Zotero integration and input conversion from terminal command flags
- Added Zotero integration and input conversion from main class programmatic access
- Added Zotero integration and input conversion from other languages access through the shared library and ports to Python, R, and Julia
### Changed
- Codebase structure
### Fixed
- Documentation inline and website to reflect changes in this release

##  [0.7.2] - 2025-04-28
### Fixed
- Wrong reference and inclusion of library for supported architecture (arm64) in Darwin, resulting in python package not working in MacOS

## [0.7.1] - 2025-04-15
### Added
- Functionality to download PDFs from text lists of URLs
- Testing and documentation of the new functionality

## [0.7.0] - 2025-04-14
### Removed
- Direct support of all LLMs and provider
### Added
- Update to the most recent possible version of all dependencies
- LLMs support though `alembica` project https://github.com/open-and-sustainable/alembica

## [joss] - 2025-04-11
### Changed
- Reviewed software release euivalent to 0.6.7 for JOSS pubblication: https://doi.org/10.21105/joss.07616

## [0.6.7] - 2025-01-31
### Added
- Instructions to use legacy versions <= 0.6.6 in Jupyter notebooks
### Removed
- User confirmation request for single model reviews

## [0.6.6] - 2025-01-18
### Added
- Support for new provider and model: DeepSeek Chat v3

## [0.6.5] - 2025-01-12
### Added
- Support for new model: Cohere's Command R7B

## [0.6.4] - 2024-11-23
### Added
- Julia package 'PrismAId', its documentation and deployment on Julia General registry
- Support for Anthropic's Claude 3.5 Haiku

## [0.6.3] - 2024-11-15
### Added
- secondary (fallback) pdf to txt conversion mechanism
- disclaimer in documentation to warn about pdf format problems
- release publishing in Matrix room for announcements
- setup of Matrix support room and reference in documentation
### Fixed
- release for arm64 platform packages in PyPI and R

## [0.6.2] - 2024-11-07
### Added
- Integration with Zotero collections and shared groups, with direct (API) download and conversion of literature pdfs
### Fixed
- updated documentation, terminal and web initializers to include Zotero integration

## [0.6.1] - 2024-10-31
### Added
- release of the project as R package published on r-universe
- web initializer of project configuration file
### Fixed
- updated documentation to include R package
### Changed
- deep refactoring of documentation to improve readability and information access

## [0.6.0] - 2024-10-25
### Added
- release of the project as Go package at github.com/Open-and-Sustainable/prismAId
- release of Python package at https://pypi.org/project/prismaid/
### Fixed
- updated documentation to include Go package download and installation
- updated documentation to include Python package

## [0.5.5] - 2024-10-23
### Added
- unit testing of each go package
- automated testing through CI/CD
### Changed
- light refactoring through interfaces for supporting testing without actual API access

## [0.5.4] - 2024-10-22
### Added
- inline documentation for package and public functions
- support for ensemble reviews
- CI/CD workflow to generate and attach binaries for each platform on release creation
### Changed
- deep code refactoring
- multiple package reorganization
### Fixed
- updated documentation to include model ensemble
- updated terminal init app to include model ensemble

## [0.5.3] - 2024-10-21
### Added
- full support for Anthropic AI models through Anthropic API
### Fixed
- updated documentation to include Anthropic

## [0.5.2] - 2024-10-15
### Added
- input file conversion from pdf, docx, and html
- generation of manuscript summaries
### Fixed
- updated documentation to include input document conversion and summary
- updated terminal init function to include input document conversion and summary

## [0.5.1] - 2024-09-30
### Added
- terminal app to create project configuration (.toml) file
- reference in documentation
### Fixed
- new costs and TPM limits on on GoogleAI Gemini 1.5 pro and flash
### Changed
- changed parameters of main class to accept activation of terminal app

## [0.5.0] - 2024-09-20
### Added
- full support for Cohere models
- check on consistency between selected supplier and models: both must be supported
- check on model input tokens limit, stops the execution if it exceeds the limit
### Fixed
- cleaned code structure of results and check packages
- updated documentation to include Cohere models and features

## [0.4.1] - 2024-08-25
### Added
- project parameter CoT justification
- automatic generation of CoT justification in OpenAI and Google AI models, as aditional prompt in same chat
- project parameter for duplicate runs (for debugging purposes)
- implementation and testing of duplication algorithm
- documentation of these features
- examples of these features
### Fixed
- output in JSON array formatting
- cleaned code structure by creating debug package and moving log setup there

## [0.4.0] - 2024-08-16
### Added
- full support for Google AI models
- support for RPM limits
- documentation for these new features
### Changed
- documentation website formatting and structure

## [0.3.2] - 2024-08-05
### Added
- support of GPT 4 Omni Mini model

## [0.3.1] - 2024-06-18
### Added
- in case of 'high' level of logging requested, log file named as project file
### Fixed
- documentation of the added feature

## [0.3] - 2024-06-17
### Added
- TPM limits support
- Documentation of this feature
### Fixed
- project.log saving in case of 'high' level of logging request

## [0.2] - 2024-05-29
### Added
- gpt-4o support & cost
- Technical FAQs
- Pages generation
### Fixed
- Cost minimization to include gpt-4o
- User manual

## [0.1.1] - 2024-05-21
### Fixed
- User Manual
- Readme

## [0.1.0] - 2024-05-16
### Added
- Input configuration file on command line
- Ouput file name form configuration file
- Tasks for building excutables
- Compiled executables for each OS and platform
### Fixed
- User Manual
- Moved log functions into main.go

## [Unreleased] - 2024-05-15
### Fixed
- Output to CSV
### Added
- Output to JSON

## [Unreleased] - 2024-05-13
### Added
- Configuration loading
- Cost assessement
- OpenAI API calls
- Output to CSV drafting

## [Unreleased] - 2024-05-03
### Added
- User manual placeholder
- Preliminary README
- Inheritance of cost module from testing project

## [Unreleased] - 2024-04-29
### Added
- Directory structure
- Test files structure

## [Unreleased] - 2024-04-22
### Added
- Changelog
- License
