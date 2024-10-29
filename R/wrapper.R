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

  # Log the full path for debugging
  message("Attempting to load library at: ", lib_path)

  # Check if the path actually exists
  if (!file.exists(lib_path)) {
    stop("Library path does not exist: ", lib_path)
  }

  # Attempt to load the library
  tryCatch({
    dyn.load(lib_path)
    message("Successfully loaded library: ", libname)
  }, error = function(e) {
    stop("Failed to load the required shared library: ", e$message)
  })
}

.onLoad <- function(libname, pkgname) {
  message("Loading package: ", pkgname)
  message("Library name passed: ", libname)
  # Determine which library to load based on the platform
  os_type <- .Platform$OS.type
  sys_name <- Sys.info()[["sysname"]]

  if (os_type == "windows") {
    libname <- "libprismaid_windows_amd64.dll"
    lib_path <- system.file("libs/windows", libname, package = pkgname)
  } else if (sys_name == "Darwin") {
    libname <- "libprismaid_darwin_amd64.dylib"
    lib_path <- system.file("libs/macos", libname, package = pkgname)
  } else if (sys_name == "Linux") {
    libname <- "libprismaid_linux_amd64.so"
    lib_path <- system.file("libs/linux", libname, package = pkgname)
  } else {
    stop("Unsupported OS")
  }

  # Log the library path
  message("Attempting to load wrapper library at: ", lib_path)

  # Check if the library path exists
  if (!file.exists(lib_path)) {
    stop("Library path does not exist: ", lib_path)
  }

  # Load the library
  tryCatch({
    dyn.load(lib_path)
    message("Successfully loaded wrapper library: ", libname)
  }, error = function(e) {
    stop("Failed to load C wrapper library: ", e$message)
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
    return(result)
}
