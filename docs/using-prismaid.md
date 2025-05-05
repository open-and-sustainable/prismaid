---
title: Using prismAId
layout: default
---

# Using prismAId

---

<details>
<summary><strong>Page Contents</strong></summary>

- [**Overview of the prismAId Toolkit**](#overview-of-the-prismaid-toolkit): introduction to the different tools available
- [**Section 1: 'Project' Details**](#section-1-project-details): reference guide to all entries in the first section of the review project configuration
- [**Section 2: 'Prompt' Details**](#section-2-prompt-details): instructions on configuring prompts for information extraction
- [**Section 3: 'Review' Details**](#section-3-review-details): information content to be extracted and reviewed
- [**Advanced Features**](#advanced-features): how to leverage debugging, validation, and other advanced capabilities

</details>

---

## Overview of the prismAId Toolkit

Since version 0.8.0, prismAId has been restructured as a comprehensive toolkit with separate tools for different parts of the systematic review workflow:

1. **Download Tool**: Acquire papers from Zotero collections or URL lists
2. **Convert Tool**: Transform files (PDF, DOCX, HTML) to plain text for analysis
3. **Review Tool**: Process systematic literature reviews based on TOML configurations

This page focuses on configuring and using the **Review Tool**. For information on the Download and Convert tools, see the [Review Support](review-support) page.

## Configuring Review Projects

Prepare a project configuration file in [TOML](https://toml.io/en/), following the three-section structure, explanations, and recommendations provided in the [`template.toml`](https://github.com/open-and-sustainable/prismaid/blob/main/projects/template.toml) and below. Alternatively, you can use the terminal-based initialization option (`-init` in binaries) or the web-based tool on the [Review Configurator](review-configurator) page.

**Section 1**, introduced below, focuses on essential project settings. **Sections 2** and **3** cover **prompt design** and follow in sequence, while **advanced features** in Section 1 are discussed at the end of this page.

## Section 1: 'Project' Details

### Project Information
```toml
[project]
name = "Use of LLM for Systematic Review"
author = "John Doe"
version = "1.0"
```
- The **`[project]`** section contains basic project information:
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
    - `no`: Deafult.
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
- **`provider`**:  Supported providers are `OpenAI`, `GoogleAI`, `Cohere`, and `Anthropic`.
- **`api_key`**: Define project-specific keys here, or leave empty to default to environment variables.
- **`model`**: select model:
    - Leave blank `''` for cost-efficient automatic model selection.
    - **OpenAI**: Models include `gpt-4o-mini`, `gpt-4o`, `gpt-4-turbo`, `gpt-3.5-turbo`.
    - **GoogleAI**: Choose from `gemini-1.5-flash`, `gemini-1.5-pro`, `gemini-1.0-pro`.
    - **Cohere**: Options are `command-r7b-12-2024`, `command-r-plus`, `command-r`, `command-light`, `command`.
    - **Anthropic**: Includes `claude-3-5-sonnet`, `claude-3-5-haiku`, `claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku`.
    - **DeepSeek**: Provides `deepseek-chat`, version 3.
- **`temperature`**: Controls response variability (range: 0 to 1 for most models); lower values increase consistency.
- **`tpm_limit`**: Defines maximum tokens per minute. Default is `0` (no delay). Use a non-zero value based on your provider TPM limits (see Rate Limits in [Advanced Features](https://open-and-sustainable.github.io/prismaid/using-prismaid.html#rate-limits) below).
- **`rpm_limits`**: Sets maximum requests per minute. Default is `0` (no limit). See provider's RPM restrictions in [Advanced Features](https://open-and-sustainable.github.io/prismaid/using-prismaid.html#rate-limits) below.

### Supported Models
For comprehensive information on supported models, input token limits, and associated costs, please refer to the provider's official documentation. Additionally, you can find a detailed comparison of all supported models in the [`alembica` documentation](https://open-and-sustainable.github.io/alembica/supported-models.html), which provides a centralized reference for model capabilities across all providers.

## Section 2: 'Prompt' Details

**Section 2** and **3** of the project configuration file define the prompts that guide AI models in extracting targeted information. This section is central to a review project, with prismAId's robust design enabling the tool's Open Science benefits.

The **`[prompt]`** section breaks down the prompt structure into essential components to ensure accurate data extraction and minimize potential misinterpretations.

### Rationale
- This section provides explicit instructions and context for the AI model.
- The prompt consists of structured elements:
<div style="text-align: center;">
    <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/prompt_struct.png" alt="Prompt Structure Diagram" style="width: 90%;">
</div>

- Each component clarifies the model's role, task, and expected output, reducing ambiguity.
- Definitions and examples enhance clarity, while a failsafe mechanism prevents forced responses if information is absent.

```toml
[prompt]
persona = "You are an experienced scientist working on a systematic review of the literature."
task = "You are asked to map the concepts discussed in a scientific paper attached here."
expected_result = "You should output a JSON object with the following keys and possible values: "
failsafe = "If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value."
definitions = "'Interest rate' is the percentage charged by a lender for borrowing money or earned by an investor on a deposit over a specific period, typically expressed annually."
example = ""
```
This structured approach increases consistency, reduces model hallucinations, and facilitates precise information extraction in line with research objectives.

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

## Section 3: 'Review' Details

The **`[review]`** section specifies the information to be extracted from the text, defining the JSON output structure with keys and their possible values.

### Rationale
- This section serves as a knowledge map to guide the extraction process.
- Each item includes:
  - `key`: A concept or topic of interest.
  - `values`: Possible values for that key.
- This structure ensures consistency and adherence to the schema. You may add as many review items as needed.

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
    - An empty string allows any value.

- **`[review.2]`**:
  - Represents the second item to review.
  - **`key`**: "regression models"
  - **`values`**: `["yes", "no"]`
    - Allows "yes" or "no" as binary options.

- **`[review.3]`**:
  - Represents the third item to review.
  - **`key`**: "geographical scale"
  - **`values`**: `["world", "continent", "river basin"]`
    - Specifies scale options for analysis.

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
- To ensure compliance with provider restrictions, manually configure the lowest applicable `tpm` and `rpm` values in your project, as prismAId relies on these explicit settings rather than performing automatic limit detection

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

Concise prompts are cost-efficient. Check costs on the provider dashboards: [OpenAI](https://platform.openai.com/usage), [Google AI](https://console.cloud.google.com/billing), [Cohere](https://dashboard.cohere.com/billing), [Anthropic](https://console.anthropic.com/dashboard), and [DeepSeek](https://platform.deepseek.com/usage).

**Note**: Cost estimates are approximate and subject to change. Users with strict budgets should verify all costs thoroughly before conducting reviews.

### Ensemble Review
Specifying multiple LLMs enables an 'ensemble' review, allowing result validation and uncertainty quantification. You can select multiple models from one or more providers, configuring each with specific parameters.

To set up an ensemble review in the `[project.llm]` section, for instance with models from five different providers, use:

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
```

## Zotero Integration with the Download Tool

Since version 0.8.0, the Zotero integration has been moved from the review project configuration to a separate Download tool in the prismAId toolkit. This change allows for a more modular approach to the systematic review workflow, separating the literature acquisition step from the analysis step.

### Using the Zotero Download Tool

To download PDFs from your Zotero library, you'll need your Zotero credentials:

1. Go to the [Zotero Settings](https://www.zotero.org/settings) page, navigate to the **Security** tab, and then to the **Applications** section to find your **user ID**:

<div style="text-align: center;">
    <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/zotero_user.png" alt="Zotero User ID" style="width: 600px;">
</div>

2. Generate an API key by clicking "Create new private key". When creating a new API key, **enable** "Allow library access" and set the **permissions** to "Read Only" for all groups under "Default Group Permissions". Provide a name for the key, such as "prismAId":

<div style="text-align: center;">
    <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/zotero_apikey.png" alt="Zotero API Key" style="width: 600px;">
</div>

3. Use the Download tool with your credentials:

```bash
# Using the binary (requires a TOML config file)
# First create a file zotero_config.toml with:
#   user = "your_username"
#   api_key = "your_api_key"
#   group = "Your Collection/Your Sub Collection"
./prismaid -download-zotero zotero_config.toml

# Using Go
import "github.com/open-and-sustainable/prismaid"
prismaid.DownloadZoteroPDFs("username", "apiKey", "Collection/Sub Collection", "./papers")

# Using Python
import prismaid
prismaid.download_zotero_pdfs("username", "api_key", "Collection/Sub Collection", "./papers")

# Using R
library(prismaid)
DownloadZoteroPDFs("username", "api_key", "Collection/Sub Collection", "./papers")

# Using Julia
using PrismAId
PrismAId.download_zotero_pdfs("username", "api_key", "Collection/Sub Collection", "./papers")
```

### Specifying Collections and Groups

The collection parameter uses a filesystem-like representation for your Zotero library structure:

- For a parent collection with a sub-collection: `"Parent Collection/Sub Collection"`
- For a group with a collection: `"Group Name/Collection Name"`

### Integration with Review Workflow

Zotero is ideal for organizing manuscripts during a systematic review:

- **Collections** are private and accessible only to you. For step-by-step instructions, see the [University of Ottawa Library's guide](https://uottawa.libguides.com/how_to_use_zotero/create_collections).

- **Groups** allow collaboration on shared references. Learn how to create a group from the [University of Ottawa Library's guide](https://uottawa.libguides.com/how_to_use_zotero/groups).

After downloading papers with the Zotero Download tool, you'll likely want to convert them to text using the Convert tool before analysis:

```bash
# After downloading PDFs with the Zotero tool
./prismaid -convert-pdf ./papers
```

**<span style="color: red; font-weight: bold;">IMPORTANT:</span>** Due to limitations in the PDF format, conversions might be imperfect. Always manually check converted manuscripts for completeness before processing them with the Review tool.

### Workflow Integration

In a systematic review workflow, the Zotero integration fits perfectly after literature identification:

1. **Design and register** your review protocol
2. **Identify literature** using search engines and selection criteria
3. **Save papers to Zotero** in a dedicated collection
4. **Download papers** using the prismAId Zotero Download tool
5. **Convert papers** using the prismAId Convert tool
6. **Configure and run** your review using the prismAId Review tool

This modular approach gives you more control over each step of the process and allows for iterative refinement at any stage.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
