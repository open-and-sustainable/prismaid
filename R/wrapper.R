# File: R/wrapper.R

# Function to safely load the shared library
safeLoadLibrary <- function(lib_path) {
  tryCatch({
    dyn.load(lib_path)
  }, error = function(e) {
    stop("Failed to load the required shared library: ", e$message)
  })
}

# Load the correct shared library depending on the operating system
if (.Platform$OS.type == "unix") {
    if (Sys.info()["sysname"] == "Linux") {
        safeLoadLibrary(paste0(system.file("libs", package = "prismaid"), "/libprismaid_linux_amd64.so"))
    } else if (Sys.info()["sysname"] == "Darwin") {
        safeLoadLibrary(paste0(system.file("libs", package = "prismaid"), "/libprismaid_macos_amd64.so"))
    }
} else if (.Platform$OS.type == "windows") {
    safeLoadLibrary(paste0(system.file("libs", package = "prismaid"), "/libprismaid_windows_amd64.dll"))
}

#' Run Review
#'
#' This function interfaces with a shared library to perform a review process on the input data.
#' @param input_string A string representing the input data.
#' @return A string indicating the result of the review process.
#' @export
#' @examples
#' RunReview("example input")
RunReview <- function(input_string) {
    # Convert input to a form suitable for C
    input_char <- charToRaw(input_string)
    result <- .C("RunReviewR", as.character(input_char), PACKAGE = "prismaid")
    # Convert result back to R character string
    output <- rawToChar(result$value)
    return(output)
}

