# File: R/wrapper.R

.onLoad <- function(libname, pkgname) {
  # Log package loading
  message("Loading package: ", pkgname)

  # Check platform support using the C function
  platform_support <- tryCatch({
    .Call("check_platform_support", PACKAGE = "prismaid")
  }, error = function(e) {
    "unknown"
  })

  if (platform_support == "supported") {
    # Determine the library file based on the platform
    os_type <- .Platform$OS.type
    sys_name <- Sys.info()[["sysname"]]
    arch <- Sys.info()[["machine"]]

    library_file <- NULL
    library_path <- NULL

    if (os_type == "windows" && grepl("x86_64|amd64", arch, ignore.case = TRUE)) {
      library_file <- "libprismaid_windows_amd64.dll"
      library_path <- system.file("libs", arch, library_file, package = pkgname)
      if (!file.exists(library_path)) {
        library_path <- system.file("libs/windows", library_file, package = pkgname)
      }
    } else if (sys_name == "Darwin" && arch == "arm64") {
      library_file <- "libprismaid_darwin_arm64.dylib"
      library_path <- system.file("libs", arch, library_file, package = pkgname)
      if (!file.exists(library_path)) {
        library_path <- system.file("libs/macos", library_file, package = pkgname)
      }
    } else if (sys_name == "Linux" && arch == "x86_64") {
      library_file <- "libprismaid_linux_amd64.so"
      library_path <- system.file("libs", arch, library_file, package = pkgname)
      if (!file.exists(library_path)) {
        library_path <- system.file("libs/linux", library_file, package = pkgname)
      }
    }

    # Try to load the native library if path exists
    if (!is.null(library_path) && file.exists(library_path)) {
      tryCatch({
        dyn.load(library_path)
        message("Successfully loaded native library: ", library_file)
      }, error = function(e) {
        warning("Failed to load native library, functions will return error messages: ", e$message)
      })
    } else if (!is.null(library_file)) {
      warning("Native library not found: ", library_file, ". Functions will return error messages.")
    }

  } else {
    # Unsupported platform - show informative message
    os_type <- .Platform$OS.type
    sys_name <- Sys.info()[["sysname"]]
    arch <- Sys.info()[["machine"]]

    packageStartupMessage(
      "prismaid: Limited functionality on this platform.\n",
      "Platform: ", sys_name, " ", arch, "\n",
      "Supported platforms: Linux x86_64, Windows x86_64, macOS ARM64\n",
      "Functions will return informative error messages.\n",
      "Consider using the command-line binary or other language bindings."
    )
  }
}

#' Run Review
#'
#' This function interfaces with a shared library to perform a review process on the input data.
#'
#' @description
#' The input data must be structured in a TOML format, consisting of several sections and parameters.
#'
#' @details
#' **\[project\]**
#' - `name`: A string representing the project title. Example: "Use of LLM for systematic review".
#' - `author`: The name of the project author. Example: "John Doe".
#' - `version`: The version number for the project configuration. Example: "1.0".
#'
#' **\[project.configuration\]**
#' - `input_directory`: The file path to the directory containing manuscripts to be reviewed. Example: "/path/to/txt/files".
#' - `results_file_name`: The path and base name for saving results (file extension will be added automatically). Example: "/path/to/save/results".
#' - `output_format`: The format for output results. Options: "csv" (default) or "json".
#' - `log_level`: Determines logging verbosity:
#'   - `"low"`: Minimal logging (default).
#'   - `"medium"`: Logs to standard output.
#'   - `"high"`: Logs to a file (see user manual for details).
#' - `duplication`: Runs model queries twice for debugging purposes. Options: "yes" or "no" (default).
#' - `cot_justification`: Requests chain-of-thought justification from the model. Options: "yes" or "no" (default).
#' - `summary`: Generates and saves summaries of manuscripts. Options: "yes" or "no" (default).
#'
#' **\[project.llm\]**
#' - Configuration for LLMs, supporting multiple instances (llm.1, llm.2, etc.) for ensemble reviews.
#' - Parameters for each LLM include:
#'   - `provider`: The LLM service provider. Options: "OpenAI", "GoogleAI", "Cohere", "Anthropic", "DeepSeek", or "Perplexity".
#'   - `api_key`: API key for the provider. If empty, environment variables will be checked.
#'   - `model`: Model name. Options vary by provider:
#'     - OpenAI: "gpt-5-nano", "gpt-5-mini", "gpt-5.2", "gpt-5.1", "gpt-5", "o4-mini", "o3-mini", "o3", "o1-mini", "o1", "gpt-4.1-nano", "gpt-4.1-mini", "gpt-4.1", "gpt-4o-mini", "gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo", or "" (default for cost optimization).
#'     - GoogleAI: "gemini-3-flash-preview", "gemini-3-pro-preview", "gemini-2.5-flash-lite", "gemini-2.5-flash", "gemini-2.5-pro", "gemini-2.0-flash-lite", "gemini-2.0-flash", "gemini-1.5-flash", "gemini-1.5-pro", or "" (default for cost optimization).
#'     - Cohere: "command-a-reasoning-08-2025", "command-a-03-2025", "command-r-08-2024", "command-r7b-12-2024", "command-r-plus", "command-r", "command-light", "command", or "" (default for cost optimization).
#'     - Anthropic: "claude-4-5-haiku", "claude-4-5-sonnet", "claude-4-5-opus", "claude-4-0-opus", "claude-4-0-sonnet", "claude-3-7-sonnet", "claude-3-5-sonnet", "claude-3-5-haiku", "claude-3-opus", "claude-3-sonnet", "claude-3-haiku", or "" (default for cost optimization).
#'     - DeepSeek: "deepseek-chat", "deepseek-reasoner", or "" (default for cost optimization).
#'     - Perplexity: "sonar-deep-research", "sonar-reasoning-pro", "sonar-pro", "sonar", or "" (default for cost optimization).
#'   - `temperature`: Controls model randomness. Range: 0 to 1 (0 to 2 for GoogleAI). Lower values reduce randomness.
#'   - `tpm_limit`: Tokens per minute limit before delaying prompts. Default: 0 (no delay).
#'   - `rpm_limit`: Requests per minute limit before delaying prompts. Default: 0 (no delay).
#'
#' **\[prompt\]**
#' - Defines the main components of the prompt for reviews.
#' - `persona`: Optional text specifying the model's role. Example: "You are an experienced scientist...".
#' - `task`: Required text framing the task for the model. Example: "Map the concepts discussed in a scientific paper...".
#' - `expected_result`: Required text describing the expected output structure in JSON.
#' - `definitions`: Optional text defining concepts to clarify instructions. Example: "'Interest rate' is defined as...".
#' - `example`: Optional example to illustrate concepts.
#' - `failsafe`: Specifies a fallback if the concepts cannot be identified. Example: "Respond with an empty '' value if concepts are unclear".
#'
#' **\[review\]**
#' - Defines the keys and possible values in the JSON object for the review.
#' - Example entries:
#'   - \[review.1\]: `key = "interest rate"`, `values = [""]`
#'   - \[review.2\]: `key = "regression models"`, `values = ["yes", "no"]`
#'   - \[review.3\]: `key = "geographical scale"`, `values = ["world", "continent", "river basin"]`
#'
#' @param input_string A string representing the input data.
#' @return A string indicating the result of the review process.
#' @export
#' @examples
#' RunReview("example input")
RunReview <- function(input_string) {
    # Directly pass the string as R character to .Call)
    result <- .Call("RunReviewR_wrap", input_string, PACKAGE = "prismaid")
    return(result)
}

#' Download PDFs from Zotero
#'
#' This function downloads PDFs from a Zotero collection to a specified directory.
#'
#' @description
#' Downloads PDF documents from a Zotero collection using the Zotero API.
#'
#' @param username Your Zotero username/user ID
#' @param api_key Your Zotero API key
#' @param collection_name The name of the Zotero collection
#' @param parent_dir Directory where PDFs will be saved
#' @return A string indicating the result of the download process
#' @export
#' @examples
#' \dontrun{
#' DownloadZoteroPDFs("user123", "apikey456", "My Collection", "/path/to/pdfs")
#' }
DownloadZoteroPDFs <- function(username, api_key, collection_name, parent_dir) {
    result <- .Call("DownloadZoteroPDFsR_wrap", username, api_key, collection_name, parent_dir, PACKAGE = "prismaid")
    return(result)
}

#' Download Files from URL List
#'
#' This function downloads files from URLs listed in a file.
#'
#' @description
#' Reads a file containing a list of URLs (one per line) and downloads the files.
#'
#' @param path Path to the file containing URLs
#' @return A string indicating the result of the download process
#' @export
#' @examples
#' \dontrun{
#' DownloadURLList("/path/to/url_list.txt")
#' }
DownloadURLList <- function(path) {
    result <- .Call("DownloadURLListR_wrap", path, PACKAGE = "prismaid")
    return(result)
}

#' Convert Files to Different Formats
#'
#' This function converts files in a directory to specified formats.
#'
#' @description
#' Converts files in the input directory to the requested formats. If standard conversion methods fail
#' and a Tika server address is provided, files are automatically sent to the Tika server for OCR-based
#' text extraction as a fallback.
#'
#' @param input_dir Directory containing files to convert
#' @param selected_formats Comma-separated list of target formats (e.g., "pdf,docx,html")
#' @param tika_address Tika server address for OCR fallback (e.g., "localhost:9998"). Empty string disables OCR fallback. Defaults to "".
#' @param single_file Convert only the specified PDF (PDF format only). Defaults to "".
#' @param ocr_only Force OCR for PDFs via Tika (PDF format only). Defaults to FALSE.
#' @return A string indicating the result of the conversion process
#' @export
#' @examples
#' \dontrun{
#' # Convert without Tika OCR
#' Convert("/path/to/files", "pdf,docx")
#'
#' # Convert with Tika OCR fallback
#' Convert("/path/to/files", "pdf", "localhost:9998")
#'
#' # OCR-only for PDFs
#' Convert("/path/to/files", "pdf", "localhost:9998", "", TRUE)
#' }
Convert <- function(input_dir, selected_formats, tika_address = "", single_file = "", ocr_only = FALSE) {
    ocr_only_value <- if (isTRUE(ocr_only)) "true" else "false"
    result <- .Call("ConvertR_wrap", input_dir, selected_formats, tika_address, single_file, ocr_only_value, PACKAGE = "prismaid")
    return(result)
}



#' Screen Manuscripts for Systematic Review
#'
#' This function screens manuscripts to identify items for exclusion based on various criteria.
#'
#' @description
#' Processes a list of manuscripts applying multiple filters to identify which should be
#' excluded from a systematic review. Supports both rule-based and AI-assisted screening.
#'
#' @details
#' The input data must be structured in a TOML format with the following sections:
#'
#' **\[project\]**
#' - `name`: Project title. Example: "Screening for climate change literature".
#' - `author`: Name of the project author. Example: "Jane Smith".
#' - `version`: Version number for the configuration. Example: "1.0".
#' - `input_file`: Path to CSV/JSON file containing manuscripts to screen. Example: "/path/to/manuscripts.csv".
#' - `output_file`: Path where screening results will be saved. Example: "/path/to/screening_results".
#' - `text_column`: Name of column containing text or path to text files. Example: "abstract" or "text_file_path".
#' - `identifier_column`: Column name for unique manuscript identifiers. Example: "doi" or "id".
#' - `output_format`: Format for results. Options: "csv" or "json".
#' - `log_level`: Logging verbosity. Options: "low", "medium", or "high".
#'
#' **\[filters.deduplication\]**
#' - `enabled`: Whether to apply deduplication. Options: true or false.
#' - `use_ai`: Use AI for semantic similarity detection. Options: true or false.
#' - `compare_fields`: List of fields to compare. Example: ["title", "abstract", "doi"].
#'
#' **\[filters.language\]**
#' - `enabled`: Whether to filter by language. Options: true or false.
#' - `accepted_languages`: List of accepted language codes. Example: ["en", "es", "fr"].
#' - `use_ai`: Use AI for language detection. Options: true or false.
#'
#' **\[filters.article_type\]**
#' - `enabled`: Whether to filter by article type. Options: true or false.
#' - `use_ai`: Use AI for article classification. Options: true or false.
#' - `exclude_reviews`: Exclude review articles. Options: true or false.
#' - `exclude_editorials`: Exclude editorial articles. Options: true or false.
#' - `exclude_letters`: Exclude letters to editor. Options: true or false.
#' - `exclude_theoretical`: Exclude theoretical papers. Options: true or false.
#' - `exclude_empirical`: Exclude empirical studies. Options: true or false.
#' - `exclude_methods`: Exclude methodology papers. Options: true or false.
#' - `exclude_single_case`: Exclude single case studies. Options: true or false.
#' - `exclude_sample`: Exclude sample-based studies. Options: true or false.
#' - `include_types`: Specific article types to include. Example: ["research", "case_study"].
#'
#' **\[filters.topic_relevance\]**
#' - `enabled`: Whether to filter by topic relevance. Options: true or false.
#' - `use_ai`: Use AI for relevance scoring. Options: true or false.
#' - `topics`: List of topic descriptions. Example: ["climate change impacts", "adaptation strategies"].
#' - `min_score`: Minimum relevance score (0-1) to include. Example: 0.7.
#' - `score_weights.keyword_match`: Weight for keyword matching (0-1). Example: 0.3.
#' - `score_weights.concept_match`: Weight for concept matching (0-1). Example: 0.4.
#' - `score_weights.field_relevance`: Weight for field relevance (0-1). Example: 0.3.
#'
#' **\[filters.llm\]** (Optional, required if any filter has `use_ai = true`)
#' - Configuration for AI models, supporting multiple instances (llm.1, llm.2, etc.).
#' - Parameters for each LLM:
#'   - `provider`: LLM service provider. Options: "OpenAI", "GoogleAI", "Cohere", "Anthropic", "DeepSeek", or "Perplexity".
#'   - `api_key`: API key for the provider. If empty, environment variables will be checked.
#'   - `model`: Model name (see RunReview documentation for available models per provider).
#'   - `temperature`: Controls randomness (0-1, or 0-2 for GoogleAI).
#'   - `tpm_limit`: Tokens per minute limit. Default: 0 (no limit).
#'   - `rpm_limit`: Requests per minute limit. Default: 0 (no limit).
#'
#' @param input_string A string containing the TOML configuration for screening.
#' @return A string indicating the result of the screening process.
#' @export
#' @examples
#' \dontrun{
#' config <- '
#' [project]
#' name = "Climate Literature Screening"
#' author = "Research Team"
#' version = "1.0"
#' input_file = "/data/manuscripts.csv"
#' output_file = "/results/screening"
#' text_column = "abstract"
#' identifier_column = "doi"
#' output_format = "csv"
#' log_level = "medium"
#'
#' [filters.deduplication]
#' enabled = true
#' use_ai = false
#' compare_fields = ["title", "doi"]
#'
#' [filters.language]
#' enabled = true
#' accepted_languages = ["en"]
#' use_ai = false
#'
#' [filters.article_type]
#' enabled = true
#' use_ai = false
#' exclude_reviews = true
#' exclude_editorials = true
#' '
#' Screening(config)
#' }
Screening <- function(input_string) {
    result <- .Call("ScreeningR_wrap", input_string, PACKAGE = "prismaid")
    return(result)
}
