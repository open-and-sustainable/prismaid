---
title: Installation & Setup
layout: default
---

# Installation & Setup

 ---
*Page Contents:*
- [**Supported Systems**](#supported-systems): platforms and languages supported by prismAId
- [**Workflow Overview**](#workflow-overview): the recommended process for utilizing prismAId features
- [**Step-by-Step Installation**](#step-by-step-installations): instructions for installation on any platform
- [**Additional Setup Information**](#additional-setup-information): supplementary guidance to help you get started

 ---

## Supported Systems
prismAId is accessible across multiple platforms, offering flexibility based on user preference and system requirements:

1. **Go Package**: Full functionality for Go-based projects.

2. **Binaries**: Standalone executables for Windows, macOS, and Linux, requiring no coding skills.

3. **Python Package** on PyPI: For integration in Python scripts and Jupyter notebooks.

4. **R Package** on R-universe: Compatible with R and RStudio environments.

5. **Julia Package** from the Github repo: For integration in Julia workflows and Jupyter notebooks.

## Workflow Overview
1. **AI Model Provider Account and API Key**:
    - Register for an account with [OpenAI](https://www.openai.com/), [GoogleAI](https://aistudio.google.com), [Cohere](https://cohere.com/), or [Anthropic](https://www.anthropic.com/) and obtain an API key from your provider’s dashboard.
    - Generate an API key from the the provider dashboard.
2. **Install prismAId**:
    - Follow the installation instructions below based on your preferred system from the Supported Systems section.
3. **Prepare Papers for Review:**
    - Ensure papers are in .txt format, or use prismAId `input_conversion` flag in `[project.configuration]` to convert PDF, DOCX, and HTML files to plain text.
4. **Define the Review Project:**
    - Set up a configuration file (.toml) specifying project parameters, including the AI model, input data, and output preferences. This configuration defines the scope and details of your systematic review.


## Step-by-Step Installation

### Option 1. Go Package

**(Supported: Linux, macOS, Windows; AMD64, Arm64)**

To add the `prismaid` Go package to your project:
1. Install with:
```bash
go get "github.com/open-and-sustainable/prismaid"
```

2. Import when needed:
```go
import "github.com/open-and-sustainable/prismaid"
```

Refer to full [documentation on pkg.go.dev](https://pkg.go.dev/github.com/open-and-sustainable/prismaid) for additional details.

### Option 2. Binaries

**(Supported: Linux, macOS, Windows; AMD64, Arm64)**

Download the appropriate executable for your OS from our [GitHub Releases](https://github.com/open-and-sustainable/prismaid/releases). No coding is required.

prismAId uses a human-readable `.toml` project configuration file for setup. You can find a template and example in the [GitHub repository](https://github.com/open-and-sustainable/prismaid/tree/main/projects). Once your `.toml` file is ready, execute the project with:
```bash
# For Windows
./prismAId_windows_amd64.exe --project your_project.toml
```

### Option 3. Python Package

**(Supported: Linux and Windows AMD64, macOS Arm64)**

Install the `prismaid` package from [PYPI](https://pypi.org/project/prismaid/) with:
```bash
pip install prismaid
```
This Python package provides an interface that wraps a C shared library, allowing configuration and review processing within Python scripts or Jupyter notebooks. Once installed, import prismAId and use it to load and execute review projects, as shown in the example below:
```python
import prismaid

# Example usage: load and run a review project configuration
with open("proj_test.toml", "r") as file:
    input_str = file.read()
error_ptr = prismaid.RunReviewPython(input_str.encode('utf-8'))

# Handle errors if they occur
if error_ptr:
    print("Error:", error_ptr.decode('utf-8'))
else:
    print("RunReview completed successfully")
```

**NOTE**: when using prismAId legacy versions <= 0.6.6 in Jupyter notebooks follow instructions [presented below](https://open-and-sustainable.github.io/prismaid/installation-setup.html#use-in-jupyter-notebooks) to run single model reviews.

### Option 4. R Package

**(Supported: Linux AMD64, macOS Arm64)**

Install the `prismaid` R package from [R-universe](https://open-and-sustainable.r-universe.dev/prismaid) using:
```r
install.packages("prismaid", repos = c("https://open-and-sustainable.r-universe.dev", "https://cloud.r-project.org"))
```

All inputs and outputs are file-based. For example, to load and run a review configuration:
```r
library(prismaid)
toml_content <- paste(readLines("proj_test.toml"), collapse = "\n")
RunReview(toml_content)
```

### Option 5. Julia Package

**(Supported: Linux and Windows AMD64, macOS Arm64)**

Install the `PrismAId` package using Julia's package manager and running the following commands in your Julia REPL. This will add the `PrismAId` package directly from the Julia General registry:
```julia
using Pkg
Pkg.add("PrismAId")
```

This Julia package provides an interface that wraps a C shared library, allowing configuration and review processing within Julia workflows and Jupyter notebooks. Once installed, import `PrismAId` and use it to load and execute review projects, as shown in the example below:
```julia
# Load the package
using PrismAId
# Input a review project configuration
toml_test = """
       [project]
       name = "Test of prismAId"
       ...
       """
# Run the review
PrismAId.run_review(toml_test)
```

## Additional Setup Information

### Initialize the Configuration File
prismAId binaries and Go module offer an interactive terminal application to help create draft configuration files. Use the -init flag to start the setup:
```bash
# For Linux on Intel
./prismAId_linux_amd64 -init
```

![Terminal app for drafting project configuration file](https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/terminal.gif)

A web-based initializer is also availeble on the [Review Configurator](review-configurator) page.

### Use in Jupyter Notebooks
When using versions <= 0.6.6 it is not possible to disable the prompt asking the user's confirmatiom to proceed with the review, leading Jupyter notebooks to crash the python engine and to the impossibility to run reviews with single models (in ensemble reviews, on the contrary, confirmation requests are automatically disabled).

To overcome this problem, it is possible to intercept the IO on the terminal as it follows:
```python
import pty
import os
import time
import select

def run_review_with_auto_input(input_str):
    master, slave = pty.openpty()  # Create a pseudo-terminal

    pid = os.fork()
    if pid == 0:  # Child process
        os.dup2(slave, 0)  # Redirect stdin
        os.dup2(slave, 1)  # Redirect stdout
        os.dup2(slave, 2)  # Redirect stderr
        os.close(master)
        import prismaid
        prismaid.RunReviewPython(input_str.encode("utf-8"))
        os._exit(0)

    else:  # Parent process
        os.close(slave)
        try:
            while True:
                rlist, _, _ = select.select([master], [], [], 5)
                if master in rlist:
                    output = os.read(master, 1024).decode("utf-8", errors="ignore")
                    if not output:
                        break  # Process finished

                    print(output, end="")

                    if "Do you want to continue?" in output:
                        print("\n[SENDING INPUT: y]")
                        os.write(master, b"y\n")
                        time.sleep(1)
        finally:
            os.close(master)
            os.waitpid(pid, 0)  # Ensure the child process is cleaned up

# Load your review (TOML) configuration
with open("config.toml", "r") as file:
    input_str = file.read()

# Run the review function
run_review_with_auto_input(input_str)
```


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
