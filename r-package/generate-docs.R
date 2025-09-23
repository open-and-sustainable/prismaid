#!/usr/bin/env Rscript

# Script to generate R documentation (.Rd files) from roxygen comments
# This script avoids loading the compiled package by using roxygen2's parse functions directly

cat("Generating documentation for prismaid R package...\n")

# Set working directory to package root
pkg_path <- getwd()
# If we're not already in the r-package directory, try to find it
if (!file.exists("DESCRIPTION")) {
  if (file.exists("r-package/DESCRIPTION")) {
    setwd("r-package")
  } else {
    # Already in the right place or need to stay where we are
    pkg_path <- "."
  }
}

# Check if roxygen2 is installed
if (!require("roxygen2", quietly = TRUE)) {
  stop("roxygen2 package is not installed. Please install it with: install.packages('roxygen2')")
}

cat("Working directory:", getwd(), "\n")

# Method 1: Try to generate docs without loading the package
tryCatch({
  cat("Attempting to generate documentation without loading package...\n")

  # Parse the package without loading the compiled code
  blocks <- roxygen2::parse_package(".", load = FALSE)

  # Create the rd roclet (documentation generator)
  rd <- roxygen2::rd_roclet()

  # Process the parsed blocks
  results <- roxygen2::roclet_process(rd, blocks)

  # Write the .Rd files
  roxygen2::roclet_output(rd, results)

  cat("Documentation generated successfully!\n")
  cat("Check the 'man/' directory for .Rd files\n")

}, error = function(e) {
  cat("Method 1 failed:", e$message, "\n")
  cat("Trying alternative method...\n")

  # Method 2: Try with source loading only
  tryCatch({
    roxygen2::roxygenise(".", load = "source", roclets = "rd")
    cat("Documentation generated successfully using source loading!\n")
    cat("Check the 'man/' directory for .Rd files\n")
  }, error = function(e2) {
    cat("Method 2 failed:", e2$message, "\n")
    cat("\n")
    cat("If both methods failed, you may need to:\n")
    cat("1. Ensure the shared library exists in inst/libs/<platform>/\n")
    cat("2. Or create a dummy file temporarily:\n")
    cat("   touch inst/libs/linux/libprismaid_linux_amd64.so\n")
    cat("3. Then run: roxygen2::roxygenise()\n")
    cat("4. Delete the dummy file afterwards\n")
  })
})

# List generated files
man_files <- list.files("man", pattern = "\\.Rd$", full.names = FALSE)
if (length(man_files) > 0) {
  cat("\nGenerated documentation files:\n")
  for (file in man_files) {
    cat("  - man/", file, "\n", sep = "")
  }
} else {
  cat("\nWarning: No .Rd files found in man/ directory\n")
}

# Clean up any compilation artifacts if they exist
so_files <- list.files("src", pattern = "\\.so$", full.names = TRUE)
o_files <- list.files("src", pattern = "\\.o$", full.names = TRUE)
if (length(so_files) > 0 || length(o_files) > 0) {
  cat("\nCleaning up compilation artifacts...\n")
  if (length(so_files) > 0) unlink(so_files)
  if (length(o_files) > 0) unlink(o_files)
  cat("Removed temporary .so and .o files\n")
}
