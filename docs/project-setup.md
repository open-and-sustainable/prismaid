---
title: Project Setup
layout: default
---

# Project Setup

## Walkthrough
1. **AI Model Provider Account and API Key:**
    - Create an account at [OpenAI](https://www.openai.com/) or [GoogleAI](https://aistudio.google.com) or [Cohere](https://cohere.com/).
    - Generate an API key from the the provider dashboard.
2. **Install the Software:**
    - Follow instructions in [Getting Started](getting-started).
3. **Prepare Papers for Review:**
    - Ensure that all papers to be reviewed are in .txt format.
    - prismAId offers the ability to automatically convert files in PDF, DOCX, and HTML formats into plain text. To enable this conversion, use the `input_conversion` flag in the `[project.configuration]` section of the review project configuration. Once activated, files are converted to `.txt` format, allowing for further processing and inspection. Since the conversion results are saved as `.txt` files, there is no need to rerun the conversion multiple times.

## Literature Review Requirements
- Follow protocols for literature search and identification, for instance as outlined in [PRISMA 2020](https://doi.org/10.1136/bmj.n71).
- Remove unnecessary elements from the articles. For example, the list of references usually does not provide relevant information. Similarly, the abstract and introductory parts often may (or should) be removed. Reviewing a review should be done with particular care and only if necessary.
- Unnecessary parts of text waste resources and increase analysis costs without any additional value. Actually, uneccessary parts could [negatively affect](https://arxiv.org/abs/2404.08865) model performance.

## How to Configure a Review Project
Prepare a project configuration file in [TOML](https://toml.io/en/) following the 3-sections structure, explanations, and suggestions presented in the [template.toml](https://github.com/open-and-sustainable/prismaid/blob/main/projects/template.toml) and here.

**Section 1** is introduced below focusing on the basic settings to configure a project. **Section 2 and 3** of the configuraiton file are discussed in [Prompt Design](prompt-design).

Additional parameters in **Section 1** can be used to activate the most advanced features of prismAId. these are discussed in [Advanced Features](advanced-features).

## Section 1: 'Project' Details

### Project Information:
```toml
[project]
name = "Use of LLM for Systematic Review"
author = "John Doe"
version = "1.0"
```
- The first section `[project]` contains basic information about the project. This includes:
  - `name`: The name of the project.
  - `author`: The author of the project.
  - `version`: The version of the project configuration.

### Configuration Details:
```toml
[project.configuration]
input_directory = "/path/to/txt/files"
input_conversion = ""
results_file_name = "/path/to/save/results"
output_format = "json"
log_level = "low"
duplication = "no"
cot_justification = "no"
summary = "no"
```
- The subsection `[project.configuration]` contains settings related to the project's execution environment:
  - `input_directory`: The directory where the .txt files to be reviewed are located.
  - `input_conversion` : Could be non active by specifyin the "" value [default] or by removing the key. Can be activated with `pdf`, `docx`, `html`, or any comma separated combination of these formats, as in `pdf,docx`.
  - `results_file_name`: The path where the results will be saved. Ensure the path exists in the filesystem.
  - `output_format`: The format of the output file, either `csv` or `json`.
  - `log_level`: The level of logging:
    - `low`: No logging, only essential text reported on stdout. The default value.
    - `medium`: Logs entries sent to stdout.
    - `high`: Saves logs to a file in the same directory as 'project_name.toml', named 'project_name.log'.
  - `duplication`: Duplication of reviews for debugging purposes:
    - `no`: The default value.
    - `yes`: Manuscripts in the input directory will be duplicated and reviewed. Duplications will be removed before the end of the program execution.
  - `cot_justification`: Printing of justifications for debugging and validation purposes:
    - `no`: The default value.
    - `yes`: Justifications will be asked to the model after results and saved manuscript by manuscript in the same directory.
  - `summary`: Printing of summary for debugging and validation purposes:
    - `no`: The default value.
    - `yes`: A summary will be asked to the model after results and saved manuscript by manuscript in the same directory.

### LLM Configuration:
```toml
[project.llm]
[project.llm.1]
provider = "OpenAI"
api_key = ""
model = ""
temperature = 0.2
tpm_limit = 0
rpm_limit = 0
```
The `[project.llm]` section includes the specifications of the models used in the review. At least one model must be selected and configured. If more than one models are configured, the project will generate multiple 'ensemble' results.

The cost estimation of the review project is conducted and presented to the user only in the case of single model reviews.

The `[project.llm.#]` section includes parameters for managing the use of the LLM:
- `provider`: Currently supported providers are `OpenAI`, `GoogleAI`, `Cohere`, and `Anthropic`.
- `api_key`: The API key can be specified here for tracking project-specific keys. If not provided, the software will look for the key in environment variables.
- `model`: Determines the model to use. Options are:
    - Leave empty - or mit the key - `''` for cost optimization (automatically selects the cheapest model based on token limits depending on provider).
    - **OpenAI** as provider: `gpt-4o-mini`, `gpt-4o`, `gpt-4-turbo`, or `gpt-3.5-turbo` for specific model selection.
    - **GoogleAI**: `gemini-1.5-flash`, `gemini-1.5-pro`, or `gemini-1.0-pro`.
    - **Cohere**: `command-r-plus`, `command-r`, `command-light`, or `command`.
    - **Anthropic**: `claude-3-5-sonnet`, `claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku`
- `temperature`: A value between 0 and 1 to control randomness. A lower value ensures replicability and accurate responses.
- `tpm_limit`: Specifies the maximum number of tokens per minute that can be processed. The default value is `0`, which indicates that there is no delay in processing prompts by prismAId. If set to a non-zero value, this parameter should reflect the minimum tokens per minute allowed by the OpenAI API for your specific model(s) and user tier. To determine the appropriate TPM limit for your use case, consult the TPM limits webpage for your provider or the summary tables on Rate Limits in [Advanced Features](advanced-features).
- `rpm_limits`: Maximum number of API requests per minute. The default value is `0`, which indicates that there is no limit. Check the summary tables on Rate Limits in [Advanced Features](advanced-features).

### Supported Models
Each model has different limits on the size of inputs and a different cost:

<table class="table-spacing">
    <thead>
        <tr>
            <th style="text-align: left;">Model</th>
            <th style="text-align: right;">Maximum Input Tokens</th>
            <th style="text-align: right;">Cost of 1M Input Tokens</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td style="text-align: left; font-style: italic;">OpenAI</td>
            <td></td>
            <td></td>
        </tr>
        <tr>
            <td style="text-align: left;">GPT-4o Mini</td>
            <td style="text-align: right;">128,000</td>
            <td style="text-align: right;">$0.15</td>
        </tr>
        <tr>
            <td style="text-align: left;">GPT-4o</td>
            <td style="text-align: right;">128,000</td>
            <td style="text-align: right;">$5.00</td>
        </tr>
        <tr>
            <td style="text-align: left;">GPT-4 Turbo</td>
            <td style="text-align: right;">128,000</td>
            <td style="text-align: right;">$10.00</td>
        </tr>
        <tr>
            <td style="text-align: left;">GPT-3.5 Turbo</td>
            <td style="text-align: right;">16,385</td>
            <td style="text-align: right;">$0.50</td>
        </tr>
        <tr>
            <td></td>
            <td></td>
            <td></td>
        </tr>
        <tr>
            <td style="text-align: left; font-style: italic;">GoogleAI</td>
            <td></td>
            <td></td>
        </tr>
        <tr>
            <td style="text-align: left;">Gemini 1.5 Flash</td>
            <td style="text-align: right;">1,048,576</td>
            <td style="text-align: right;">$0.15</td>
        </tr>
        <tr>
            <td style="text-align: left;">Gemini 1.5 Pro</td>
            <td style="text-align: right;">2,097,152</td>
            <td style="text-align: right;">$2.50</td>
        </tr>
        <tr>
            <td style="text-align: left;">Gemini 1.0 Pro</td>
            <td style="text-align: right;">32,760</td>
            <td style="text-align: right;">$0.50</td>
        </tr>
        <tr>
            <td></td>
            <td></td>
            <td></td>
        </tr>
        <tr>
            <td style="text-align: left; font-style: italic;">Cohere</td>
            <td></td>
            <td></td>
        </tr>
        <tr>
            <td style="text-align: left;">Command R+</td>
            <td style="text-align: right;">128,000</td>
            <td style="text-align: right;">$2.50</td>
        </tr>
        <tr>
            <td style="text-align: left;">Command R</td>
            <td style="text-align: right;">128,000</td>
            <td style="text-align: right;">$0.15</td>
        </tr>
        <tr>
            <td style="text-align: left;">Command Light</td>
            <td style="text-align: right;">4,096</td>
            <td style="text-align: right;">$0.30</td>
        </tr>
        <tr>
            <td style="text-align: left;">Command</td>
            <td style="text-align: right;">4,096</td>
            <td style="text-align: right;">$1.00</td>
        </tr>
        <tr>
            <td></td>
            <td></td>
            <td></td>
        </tr>
        <tr>
            <td style="text-align: left; font-style: italic;">Anthropic</td>
            <td></td>
            <td></td>
        </tr>
        <tr>
            <td style="text-align: left;">Claude 3.5 Sonnet</td>
            <td style="text-align: right;">200,000</td>
            <td style="text-align: right;">$3.00</td>
        </tr>
        <tr>
            <td style="text-align: left;">Claude 3 Sonnet</td>
            <td style="text-align: right;">200,000</td>
            <td style="text-align: right;">$3.00</td>
        </tr>
        <tr>
            <td style="text-align: left;">Claude 3 Opus</td>
            <td style="text-align: right;">200,000</td>
            <td style="text-align: right;">$15.00</td>
        </tr>
        <tr>
            <td style="text-align: left;">Claude 3 Haiku</td>
            <td style="text-align: right;">200,000</td>
            <td style="text-align: right;">$0.25</td>
        </tr>
    </tbody>
</table>

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
