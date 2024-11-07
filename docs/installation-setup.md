---
title: Installation & Workflow
layout: default
---

# Installation & Workflow

## Supported Systems
prismAId is accessible across multiple platforms, offering flexibility based on user preference and system requirements:

1. **Go Package**: Full functionality for Go-based projects.

2. **Python Package** on PyPI: For integration in Python scripts and Jupyter notebooks.

3. **R Package** on R-universe: Compatible with R and RStudio environments.

4. **Binaries**: Standalone executables for Windows, macOS, and Linux, requiring no coding skills.

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

To add the prismAId Go package to your project:
1. Install with:
```bash
go get "github.com/open-and-sustainable/prismaid"
```

2. Import when needed:
```go
import "github.com/open-and-sustainable/prismaid"
```

Refer to full [documentation on pkg.go.dev](https://pkg.go.dev/github.com/open-and-sustainable/prismaid) for additional details.


### Option 2. Python Package

Install the prismAId package from [PYPI](https://pypi.org/project/prismaid/) with:
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

### Option 3. R Package

Install the prismAId R package from [R-universe](https://open-and-sustainable.r-universe.dev/prismaid) using:
```r
install.packages("prismaid", repos = c("https://open-and-sustainable.r-universe.dev", "https://cloud.r-project.org"))
```

All inputs and outputs are file-based. For example, to load and run a review configuration:
```r
library(prismaid)
toml_content <- paste(readLines("proj_test.toml"), collapse = "\n")
RunReview(toml_content)
```

### Option 4. Binaries

Download the appropriate executable for your OS from our [GitHub Releases](https://github.com/open-and-sustainable/prismaid/releases). No coding is required.

prismAId uses a human-readable `.toml` project configuration file for setup. You can find a template and example in the [GitHub repository](https://github.com/open-and-sustainable/prismaid/tree/main/projects). Once your `.toml` file is ready, execute the project with:
```bash
# For Windows
./prismAId_windows_amd64.exe -project your_project.toml
```

## Addiitonal Setup Information

### Initialize the Configuration File
prismAId binaries and Go module offer an interactive terminal application to help create draft configuration files. Use the -init flag to start the setup: 
```bash
# For Linux on Intel
./prismAId_linux_amd64 -init
```

![Terminal app for drafting project configuration file](https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/terminal.gif)

A web-based initializer is also availeble on the [Review Configurator](review-configurator) page.

### Literature Preparation
Follow documented protocols for literature search and identification, such as [PRISMA 2020](https://doi.org/10.1136/bmj.n71). You may remove non-essential sections, like reference lists, abstracts, and introductions, which typically do not contribute relevant information. Exercise caution when including review articles unless necessary, as they can complicate analysis.

Removing unnecessary content helps reduce costs and resource usage and may improve model performance, as excessive information can [negatively affect](https://arxiv.org/abs/2404.08865) analysis outcomes.

Additionally, the tool supports integration with Zotero, allowing you to incorporate collections and groups of literature manuscripts directly into the review process. For more details on this feature, see the [Zotero Integration](https://open-and-sustainable.github.io/prismaid/using-prismaid.html#zotero-integration) section.

### Cost Estimation at Startup
After loading the project configuration, prismAId provides an estimated cost (in USD) to run the review using the specified OpenAI model. This estimate primarily reflects the input processing cost, which is typically the largest component in review projects.

To proceed, the user must confirm by entering 'y'; otherwise, the process exits without making API calls, ensuring no cost is incurred:
```bash
Total cost (USD - $): 0.0035965
Do you want to continue? (y/n): 
```
**Note**: Cost estimation is only available when a single model is configured; ensemble reviews do not include this feature.


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>