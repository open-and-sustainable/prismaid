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
