# File: R/wrapper.R

# Function to safely load the shared library
safeLoadLibrary <- function() {
  # Determine the operating system
  if (.Platform$OS.type == "unix") {
    os_type <- tolower(Sys.info()[["sysname"]])
    if (os_type == "darwin") {
      libname <- "libprismaid_darwin_amd64.dylib"  # macOS
    } else {
      libname <- "libprismaid_linux_amd64.so"     # Linux and other Unix-like systems
    }
  } else if (.Platform$OS.type == "windows") {
    libname <- "libprismaid_windows_amd64.dll"      # Windows
  } else {
    stop("Unsupported OS")
  }

  # Construct the full path to the library
  lib_path <- system.file("libs", os_type, libname, package = "prismaid")

  # Check if the path actually exists
  if (!file.exists(lib_path)) {
    stop("Library path does not exist: ", lib_path)
  }

  # Attempt to load the library
  tryCatch({
    dyn.load(lib_path)
  }, error = function(e) {
    stop("Failed to load the required shared library: ", e$message)
  })
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
    # Directly pass the string as R character to .Call)
    result <- .Call("RunReviewR_wrap", input_string, PACKAGE = "prismaid")
    return(output)
}

.onLoad <- function(libname, pkgname) {
  safeLoadLibrary()
}
