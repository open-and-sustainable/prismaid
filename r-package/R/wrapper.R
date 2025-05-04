# File: R/wrapper.R

.onLoad <- function(libname, pkgname) {
  # Log package loading
  message("Loading package: ", pkgname)

  # Determine the library file based on the platform
  os_type <- .Platform$OS.type
  sys_name <- Sys.info()[["sysname"]]

  library_file <- NULL
  library_path <- NULL

  if (os_type == "windows") {
    library_file <- "libprismaid_windows_amd64.dll"
    library_path <- system.file("libs/windows", library_file, package = pkgname)
  } else if (sys_name == "Darwin") {
    library_file <- "libprismaid_darwin_arm64.dylib"
    library_path <- system.file("libs/macos", library_file, package = pkgname)
  } else if (sys_name == "Linux") {
    library_file <- "libprismaid_linux_amd64.so"
    library_path <- system.file("libs/linux", library_file, package = pkgname)
  } else {
    stop("Unsupported OS: ", sys_name)
  }

  # Load the dynamic library with an explicit path
  dyn.load(library_path)

  # Log the library path
  message("Attempting to load library from: ", library_path)

  # Check if the library path exists
  if (!file.exists(library_path)) {
    stop("Library path does not exist: ", library_path)
  }

  # Load the library
  tryCatch({
    dyn.load(library_path)
    message("Successfully loaded library: ", library_file)
  }, error = function(e) {
    stop("Failed to load C wrapper library: ", e$message)
  })
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
#'   - `provider`: The LLM service provider. Options: "OpenAI", "GoogleAI", "Cohere", or "Anthropic".
#'   - `api_key`: API key for the provider. If empty, environment variables will be checked.
#'   - `model`: Model name. Options vary by provider:
#'     - OpenAI: "gpt-3.5-turbo", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini", or "" (default for cost optimization).
#'     - GoogleAI: "gemini-1.5-flash", "gemini-1.5-pro", "gemini-1.0-pro", or "" (default for cost optimization).
#'     - Cohere: "command-r7b-12-2024", "command-r-plus", "command-r", "command-light", "command", or "" (default for cost optimization).
#'     - Anthropic: "claude-3-5-sonnet", "claude-3-5-haiku", "claude-3-opus", "claude-3-sonnet", "claude-3-haiku", or "" (default for cost optimization).
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
#' Converts files in the input directory to the requested formats.
#'
#' @param input_dir Directory containing files to convert
#' @param selected_formats Comma-separated list of target formats (e.g., "pdf,docx,html")
#' @return A string indicating the result of the conversion process
#' @export
#' @examples
#' \dontrun{
#' Convert("/path/to/files", "pdf,docx")
#' }
Convert <- function(input_dir, selected_formats) {
    result <- .Call("ConvertR_wrap", input_dir, selected_formats, PACKAGE = "prismaid")
    return(result)
}
