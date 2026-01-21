---
title: Review Tool
layout: default
---

# Review Tool

---

## Purpose and Capabilities

The prismAId Review tool is the core component of the prismAId toolkit, designed to transform how researchers conduct systematic literature reviews. It uses Large Language Models (LLMs) to extract structured information from scientific papers based on user-defined protocols.

Key capabilities include:

1. **Protocol-Based Extraction**: Implements structured, replicable protocols for consistent reviews
2. **Configurable Analysis**: Allows precise definition of the information to be extracted
3. **Multi-Provider Support**: Works with multiple AI providers (OpenAI, GoogleAI, Cohere, Anthropic, DeepSeek, Perplexity)
4. **Ensemble Reviews**: Enables validation through multiple models for enhanced reliability
5. **Structured Output**: Generates organized CSV or JSON results for further analysis
6. **Chain-of-Thought Tracking**: Optional justification logs to track the AI's reasoning process
7. **Cost Management**: Features for minimizing and tracking API usage costs

The Review tool bridges the gap between traditional manual reviews and custom AI solutions, offering a powerful yet accessible approach to systematic reviews without requiring coding skills. It is accompanied by the [Download Tool](download-tool) and the [Convert Tool](convert-tool) to streamline workflows and assist users throughout the review process.

<div style="text-align: left;">
    <img src="https://raw.githubusercontent.com/open-and-sustainable/prismaid/main/figures/info_extract_tools.png" alt="Tools Overview" style="width: 600px;">
</div>


## Usage Methods

The Review tool can be accessed through multiple interfaces to accommodate different workflows:

### Binary (Command Line)

```bash
# Run a systematic review with a TOML configuration file
./prismaid -project your_project.toml

# Initialize a new project configuration interactively
./prismaid -init
```

### Go Package

```go
import "github.com/open-and-sustainable/prismaid"

// Run a systematic review with a TOML configuration string
tomlConfig := "..." // Your TOML configuration as a string
err := prismaid.Review(tomlConfig)
```

### Python Package

```python
import prismaid

# Run a systematic review with a TOML configuration file
with open("project.toml", "r") as file:
    toml_config = file.read()
prismaid.review(toml_config)
```

### R Package

```r
library(prismaid)

# Run a systematic review with a TOML configuration file
toml_content <- paste(readLines("project.toml"), collapse = "\n")
RunReview(toml_content)
```

### Julia Package

```julia
using PrismAId

# Run a systematic review with a TOML configuration file
toml_config = read("project.toml", String)
PrismAId.run_review(toml_config)
```

## Configuration File Structure

The Review tool is driven by a TOML configuration file that defines all aspects of your systematic review. You can create this file manually, use the `-init` flag with the binary for an interactive setup, or use the [Review Configurator](../review/review-configurator) web tool.

The configuration file consists of three main sections:

1. **Project Section**: Basic project information, execution settings, and LLM configurations
2. **Prompt Section**: Structured components that guide the AI in extracting information
3. **Review Section**: Definition of the specific information to be extracted

## Section 1: Project Details

### Project Information
```toml
[project]
name = "Use of LLM for Systematic Review"
author = "John Doe"
version = "1.0"
```
The **`[project]`** section contains basic project information:
- **`name`**: Project title.
- **`author`**: Project author.
- **`version`**: Configuration version.

### Configuration Details
```toml
[project.configuration]
input_directory = "/path/to/txt/files"
results_file_name = "/path/to/save/results"
output_format = "json"
log_level = "low"
duplication = "no"
cot_justification = "no"
summary = "no"
```
**`[project.configuration]`** specifies execution settings:
- **`input_directory`**: Location of `.txt` files for review.
- **`results_file_name`**: Path to save results.
- **`output_format`**: `csv` or `json`.
- **`log_level`**: Sets log detail:
    - `low`: Minimal logging, essential output only (default).
    - `medium`: Logs details sent to stdout.
    - `high`: Logs are saved in a file.
- **`duplication`**:  Controls review duplication for debugging:
    - `no`: Default.
    - `yes`: Files in the input directory are duplicated, reviewed, and removed before the program concludes.
- **`cot_justification`**: Adds justification logs:
    - `no`: Default.
    - `yes`: Logs justification per manuscript, saved in the same directory.
- **`summary`**: Enables summary logging:
    - `no`: Default.
    - `yes`: A summary is generated for each manuscript and saved in the same directory.

### LLM Configuration
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
- **`[project.llm]`** specifies model configurations for review execution. At least one model is required. When multiple models are configured, results will represent an 'ensemble' analysis.

The **`[project.llm.#]`** fields manage LLM usage:
- **`provider`**:  Supported providers are `OpenAI`, `GoogleAI`, `Cohere`, `Anthropic`, `DeepSeek`, `Perplexity`, `AWS Bedrock`, `Azure AI`, `Vertex AI`, and `SelfHosted` (for OpenAI-compatible endpoints).
- **`api_key`**: Define project-specific keys here, or leave empty to default to environment variables.
- **`model`**: select model:
    - Leave blank `''` for cost-efficient automatic model selection.
    - **OpenAI**: Models include `gpt-5-nano`, `gpt-5-mini`, `gpt-5.2`, `gpt-5.1`, `gpt-5`, `o4-mini`, `o3-mini`, `o3`, `o1-mini`, `o1`, `gpt-4.1-nano`, `gpt-4.1-mini`, `gpt-4.1`, `gpt-4o-mini`, `gpt-4o`, `gpt-4-turbo`, `gpt-3.5-turbo`.
    - **GoogleAI**: Choose from `gemini-3-flash-preview`, `gemini-3-pro-preview`, `gemini-2.5-flash-lite`, `gemini-2.5-flash`, `gemini-2.5-pro`, `gemini-2.0-flash-lite`, `gemini-2.0-flash`, `gemini-1.5-flash`, `gemini-1.5-pro`.
    - **Cohere**: Options are `command-a-reasoning-08-2025`, `command-a-03-2025`, `command-r-08-2024`, `command-r7b-12-2024`, `command-r-plus`, `command-r`, `command-light`, `command`.
    - **Anthropic**: Includes `claude-4-5-haiku`, `claude-4-5-sonnet`, `claude-4-5-opus`, `claude-4-0-opus`, `claude-4-0-sonnet`, `claude-3-7-sonnet`, `claude-3-5-sonnet`, `claude-3-5-haiku`, `claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku`.
    - **DeepSeek**: Provides `deepseek-chat`, and `deepseek-reasoner` version 3.
    - **Perplexity**: Supports `sonar-deep-research`, `sonar-reasoning-pro`, `sonar-pro`, `sonar`.
- **`temperature`**: Controls response variability (range: 0 to 1 for most models); lower values increase consistency.
- **`tpm_limit`**: Defines maximum tokens per minute. Default is `0` (no delay).
- **`rpm_limit`**: Sets maximum requests per minute. Default is `0` (no limit).

**Optional fields for cloud providers and self-hosted endpoints:**
- **`base_url`**: Base URL for self-hosted OpenAI-compatible endpoints (e.g., `http://localhost:8000/v1`). Use with `provider = "SelfHosted"`.
- **`endpoint_type`**: Cloud endpoint type. Options: `bedrock`, `azure`, `vertex`. Required for cloud providers.
- **`region`**: AWS region for Bedrock (e.g., `us-east-1`). Required when using AWS Bedrock.
- **`project_id`**: Google Cloud project ID for Vertex AI. Required when using Vertex AI.
- **`location`**: Google Cloud location for Vertex AI (e.g., `us-central1`). Required when using Vertex AI.
- **`api_version`**: API version for Azure OpenAI (e.g., `2024-02-15-preview`). Required when using Azure AI.

### Supported Models
For comprehensive information on supported models, input token limits, and associated costs, please refer to the provider's official documentation. Additionally, you can find a detailed comparison of all supported models in the [`alembica` documentation](https://open-and-sustainable.github.io/alembica/supported-models.html).

## Section 2: Prompt Details

The **`[prompt]`** section breaks down the prompt structure into essential components to ensure accurate data extraction and minimize potential misinterpretations.

### Prompt Structure
<div style="text-align: center;">
    <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/prompt_struct.png" alt="Prompt Structure Diagram" style="width: 90%;">
</div>

Each component clarifies the model's role, task, and expected output, reducing ambiguity. Definitions and examples enhance clarity, while a failsafe mechanism prevents forced responses if information is absent.

```toml
[prompt]
persona = "You are an experienced scientist working on a systematic review of the literature."
task = "You are asked to map the concepts discussed in a scientific paper attached here."
expected_result = "You should output a JSON object with the following keys and possible values: "
failsafe = "If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value."
definitions = "'Interest rate' is the percentage charged by a lender for borrowing money or earned by an investor on a deposit over a specific period, typically expressed annually."
example = ""
```

### Entry Details

- **`persona`**:
  - Example: "You are an experienced scientist working on a systematic review of the literature."
  - Purpose: Sets the model's role, providing context to guide responses appropriately.

- **`task`**:
  - Example: "You are asked to map the concepts discussed in a scientific paper attached here."
  - Purpose: Defines the model's specific task, clarifying its objectives.

- **`expected_result`**:
  - Example: "You should output a JSON object with the following keys and possible values."
  - Purpose: Specifies the output format, ensuring structured responses.

- **`failsafe`**:
  - Example: "If the concepts neither are clearly discussed in the document nor deducible, respond with an empty '' value."
  - Purpose: Prevents the model from generating forced responses when information is missing, enhancing accuracy.

- **`definitions`**:
  - Example: "'Interest rate' is the percentage charged by a lender for borrowing money."
  - Purpose: Provides precise definitions to reduce misinterpretations.

- **`example`**:
  - Example: "For example, given the text 'A recent global analysis based on ARIMA models suggests that wind energy products return is 4.3% annually.' the output JSON object could be: {"interest rate": 4.3, "regression models": "yes", "geographical scale": "world"}"
  - Purpose: Offers a sample output to further clarify expectations, guiding the model toward accurate responses.

## Section 3: Review Details

The **`[review]`** section specifies the information to be extracted from the text, defining the JSON output structure with keys and their possible values.

```toml
[review]
[review.1]
key = "interest rate"
values = [""]
[review.2]
key = "regression models"
values = ["yes", "no"]
[review.3]
key = "geographical scale"
values = ["world", "continent", "river basin"]
```

### Entry Details

- **`[review]`**:
  - Header indicating the start of review items, defining the structure of the knowledge map.

- **`[review.1]`**:
  - Represents the first item to review.
  - **`key`**: "interest rate"
    - The concept or topic to be extracted.
  - **`values`**: `[""]`
    - An empty string allows any value (typically for numerical or open text fields).

- **`[review.2]`**:
  - Represents the second item to review.
  - **`key`**: "regression models"
  - **`values`**: `["yes", "no"]`
    - Restricts responses to "yes" or "no" as binary options.

- **`[review.3]`**:
  - Represents the third item to review.
  - **`key`**: "geographical scale"
  - **`values`**: `["world", "continent", "river basin"]`
    - Limits responses to the specific scales listed.

## Advanced Features

### Debugging & Validation
In **Section 1** of the project configuration, three parameters support project development and prompt testing:
  - **`log_level`**: Controls logging detail with options: `low` (default), `medium`, and `high`.
  - **`duplication`**: Enables prompt duplication for consistency testing (`no`/`yes`).
  - **`cot_justification`**: Activates Chain-of-Thought justifications (`no`/`yes`).

Increasing `log_level` beyond `low` provides detailed API response insights, visible on the terminal (`medium`) or saved to a log file (`high`).

**Duplication** helps validate prompt clarity by duplicating reviews. Inconsistent outputs across duplicates indicate unclear prompts. Costs are displayed based on duplication settings.

**CoT Justification** generates a .txt file per manuscript, logging the model's thought process, responses, and relevant passages. Example output:
```md
- **clustering**: "no" - The text does not mention any clustering techniques or grouping of data points based on similarities.
- **copulas**: "yes" - The text explicitly mentions the use of copulas to model the joint distribution of multiple flooding indicators (maximum soil moisture, runoff, and precipitation). "The multidimensional representation of the joint distributions of relevant hydrological climate impacts is based on the concept of statistical copulas [43]."
- **forecasting**: "yes" - The text explicitly mentions the use of models to predict future scenarios of flooding hazards and damage. "Future scenarios use hazard and damage data predicted for the period 2018–2100."
```

### Rate Limits

The prismAId toolkit allows you to manage model usage limits through two key parameters in the **[project.llm]** section of your configuration:

- **`tpm_limit`**: Defines the maximum tokens processed per minute
- **`rpm_limit`**: Sets the maximum requests per minute

By default, both parameters are set to `0`, which applies no rate limiting. When configured with non-zero values, prismAId automatically enforces appropriate delays to ensure your usage remains within specified limits.

For comprehensive information on provider-specific rate limits, we recommend consulting each provider's official documentation. The [`alembica` documentation](https://open-and-sustainable.github.io/alembica/rate-limits.html) also offers a centralized reference comparing rate limits across all supported models and providers.

**Important considerations:**
- Daily request limits are not automatically managed by prismAId and require manual monitoring
- To ensure compliance with provider restrictions, manually configure the lowest applicable `tpm` and `rpm` values in your project

### Cost Minimization
In **Section 1** of the project configuration:
- **`model`**: Leaving this field empty (`''`) enables automatic selection of the most cost-efficient model from the chosen provider. This may result in varying models for manuscripts based on length and token limits.

#### How Costs are Computed
Cost minimization considers both the cost of using the model for each unit (token) of input and the total number of input tokens, because more economical models may have stricter limits on how much data they can handle.

- **Tokenization Libraries**: prismAId uses libraries specific to each provider:
  - OpenAI's cost minimization uses the [Tiktoken library](https://github.com/pkoukk/tiktoken-go).
  - Google's token minimization uses the [CountTokens API](https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/count-tokens).
  - Cohere uses its [API](https://docs.cohere.com/docs/rate-limits).
  - Anthropic approximates token counts via OpenAI's tokenizer.
  - DeepSeek approximates token counts via OpenAI's tokenizer.

Concise prompts are cost-efficient. Check costs on the provider dashboards: [OpenAI](https://platform.openai.com/usage), [Google AI](https://console.cloud.google.com/billing), [Cohere](https://dashboard.cohere.com/billing), [Anthropic](https://console.anthropic.com/dashboard), [DeepSeek](https://platform.deepseek.com/usage), and [Perplexity](https://www.perplexity.ai/settings/api).

**Note**: Cost estimates are approximate and subject to change. Users with strict budgets should verify all costs thoroughly before conducting reviews.

### Ensemble Review
Specifying multiple LLMs enables an 'ensemble' review, allowing result validation and uncertainty quantification. You can select multiple models from one or more providers, configuring each with specific parameters.

Example ensemble configuration with models from five different providers:

```toml
[project.llm]
[project.llm.1]
provider = "OpenAI"
api_key = ""
model = "gpt-4o-mini"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

[project.llm.2]
provider = "GoogleAI"
api_key = ""
model = "gemini-1.5-flash"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

[project.llm.3]
provider = "Cohere"
api_key = ""
model = "command-r"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

[project.llm.4]
provider = "Anthropic"
api_key = ""
model = "claude-3-haiku"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

[project.llm.5]
provider = "DeepSeek"
api_key = ""
model = "deepseek-chat"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

[project.llm.6]
provider = "Perplexity"
api_key = ""
model = "sonar-pro"
temperature = 0.01
tpm_limit = 0
rpm_limit = 0

# Example: AWS Bedrock
[project.llm.7]
provider = "AWS Bedrock"
api_key = ""  # Uses AWS_ACCESS_KEY_ID from environment
model = "anthropic.claude-3-sonnet-20240229-v1:0"
temperature = 0.01
endpoint_type = "bedrock"
region = "us-east-1"
tpm_limit = 0
rpm_limit = 0

# Example: Azure OpenAI
[project.llm.8]
provider = "Azure AI"
api_key = ""  # Uses AZURE_OPENAI_API_KEY from environment
model = "gpt-4o"
temperature = 0.01
endpoint_type = "azure"
base_url = "https://your-resource.openai.azure.com"
api_version = "2024-02-15-preview"
tpm_limit = 0
rpm_limit = 0

# Example: Vertex AI
[project.llm.9]
provider = "Vertex AI"
api_key = ""  # Uses GOOGLE_APPLICATION_CREDENTIALS from environment
model = "gemini-1.5-pro"
temperature = 0.01
endpoint_type = "vertex"
project_id = "your-gcp-project-id"
location = "us-central1"
tpm_limit = 0
rpm_limit = 0

# Example: Self-hosted OpenAI-compatible endpoint
[project.llm.10]
provider = "SelfHosted"
api_key = "your-api-key"
model = "llama-3-70b"
temperature = 0.01
base_url = "http://localhost:8000/v1"
tpm_limit = 0
rpm_limit = 0
```

## Best Practices

### Project Configuration Best Practices

1. **Clear Project Organization**:
   - Use descriptive project names and versioning
   - Store configuration files in a version control system
   - Document any modifications to configuration files

2. **Path Management**:
   - Use absolute paths to avoid confusion
   - Ensure all directories exist before running the review
   - Keep input and output directories separate and well-organized

3. **Testing and Validation**:
   - Start with a small sample of papers to validate configuration
   - Use the duplication feature to check response consistency
   - Gradually scale up to full dataset after validation

### Prompt Design Best Practices

1. **Clarity and Specificity**:
   - Define the persona clearly to set appropriate context
   - Make tasks unambiguous and precisely defined
   - Specify output formats explicitly

2. **Example-Driven Design**:
   - Include clear examples whenever possible
   - Show the exact format of expected outputs
   - Demonstrate edge cases if applicable

3. **Failsafe Mechanisms**:
   - Always include clear failsafe instructions
   - Define how missing information should be handled
   - Specify alternatives for ambiguous cases

### Review Structure Best Practices

1. **Focused Information Extraction**:
   - Limit each review project to a coherent set of related information
   - Consider creating separate review projects for distinct types of information
   - Structure review items in a logical sequence

2. **Value Constraints**:
   - Use empty value lists (`[""]`) only when necessary for numerical or free-text fields
   - For categorical data, always provide exhaustive value lists
   - Include catch-all options like "other" if appropriate

3. **Manageable Complexity**:
   - Keep the number of review items reasonable (generally under 10 per review)
   - For complex reviews, consider breaking into multiple focused projects
   - Structure complex taxonomies hierarchically

## Workflow Integration

The Review tool is the culmination of the systematic review workflow:

1. **Literature Search**:
   - Search databases and identify potentially relevant papers
   - Export search results to CSV or reference manager

2. **Screening** ([Screening Tool](screening-tool)):
   - Filter out duplicates, wrong languages, and irrelevant article types
   - Create a refined list of papers to acquire

3. **Literature Acquisition** ([Download Tool](download-tool)):
   - Download only the screened papers from Zotero collections or URL lists

4. **Format Conversion** ([Convert Tool](convert-tool)):
   - Convert downloaded papers to text format for analysis

5. **Review Configuration**:
   - Set up your review project configuration using the [Review Configurator](../review/review-configurator) or the `-init` flag
   - Define your information extraction protocol through prompt and review sections

6. **Systematic Review** (Review Tool):
   - Process the converted text files to extract structured information
   - Analyze results for patterns, trends, and insights

7. **Results Analysis**:
   - Use the structured CSV or JSON outputs for further analysis
   - Integrate with other tools like R, Python, or spreadsheet applications

By using the Review tool as part of this integrated workflow, researchers can conduct comprehensive, protocol-based systematic reviews with unprecedented efficiency and consistency.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
