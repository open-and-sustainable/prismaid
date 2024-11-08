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
#' - `input_conversion`: Specifies manuscript conversion formats:
#'   - `""`: Default, non-active conversion.
#'   - `"pdf"`, `"docx"`, `"html"`, or combinations like `"pdf,docx"`.
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
#' - Configuration for LLMs, supporting multiple providers for ensemble reviews.
#' - Parameters include:
#'   - `provider`: The LLM service provider. Options: "OpenAI", "GoogleAI", "Cohere", or "Anthropic".
#'   - `api_key`: API key for the provider. If empty, environment variables will be checked.
#'   - `model`: Model name. Options vary by provider:
#'     - OpenAI: "gpt-3.5-turbo", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini", or "" (default).
#'     - GoogleAI: "gemini-1.5-flash", "gemini-1.5-pro", "gemini-1.0-pro", or "" (default).
#'     - Cohere: "command-r-plus", "command-r", "command-light", "command", or "" (default).
#'     - Anthropic: "claude-3-5-sonnet", "claude-3-opus", "claude-3-sonnet", "claude-3-haiku", or "" (default).
#'   - `temperature`: Controls model randomness. Range: 0 to 1 (or 0 to 2 for GoogleAI). Lower values reduce randomness.
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
