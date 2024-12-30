---
title: Using prismAId
layout: default
---

# Using prismAId

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
input_conversion = ""
results_file_name = "/path/to/save/results"
output_format = "json"
log_level = "low"
duplication = "no"
cot_justification = "no"
summary = "no"
```
**`[project.configuration]`** specifies execution settings:
- **`input_directory`**: Location of `.txt` files for review.
- **`input_conversion`**: Non-active if left empty (default) or key removed. Enable with `pdf`, `docx`, `html`, or as a comma-separated list (e.g., `pdf,docx`).
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

### Zotero Section
```toml
[project.zotero]
user = "12345678"
api_key = "fdjkdfnjhfd4556"
group = "My Group/My Collection"
```
- **`[project.zotero]`** contains the parameters needed to integrate Zotero collections or groups into your review process. Omitting this section or leaving its fileds empty (i.e., `""`) will disable Zotero integration. See details also [below](https://open-and-sustainable.github.io/prismaid/using-prismaid.html#zotero-integration).

Parameters:
- **`user`**: Your Zotero user ID, which can be found by visiting [Zotero Settings](https://www.zotero.org/settings). Look for "User ID for use in API calls" under your API keys.
- **`api_key`**: A private API key for accessing the Zotero API. Create one by going to [Zotero Settings](https://www.zotero.org/settings) and selecting "Create new private key". When creating the key, ensure that you enable "Allow library access" and set the permissions to "Read Only" for all groups under "Default Group Permissions".
- **`group`**: The name of the collection or group containing the documents you wish to review. If the collection or group is nested, represent the hierarchy using a forward slash (/), e.g., "Parent Collection/Sub Collection".

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
    - **Cohere**: Options are `command-r-plus`, `command-r`, `command-light`, `command`.
    - **Anthropic**: Includes `claude-3-5-sonnet`, `claude-3-5-haiku`, `claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku`.
- **`temperature`**: Controls response variability (range: 0 to 1 for most models); lower values increase consistency.
- **`tpm_limit`**: Defines maximum tokens per minute. Default is `0` (no delay). Use a non-zero value based on your provider TPM limits (see Rate Limits in [Advanced Features](https://open-and-sustainable.github.io/prismaid/using-prismaid.html#rate-limits) below).
- **`rpm_limits`**: Sets maximum requests per minute. Default is `0` (no limit). See provider’s RPM restrictions in [Advanced Features](https://open-and-sustainable.github.io/prismaid/using-prismaid.html#rate-limits) below.

### Supported Models
Each model has specific limits for input size and costs, as summarized below:

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
            <td style="text-align: left;">Claude 3.5 Haiku</td>
            <td style="text-align: right;">200,000</td>
            <td style="text-align: right;">$1.00</td>
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

## Section 2: 'Prompt' Details

**Section 2** and **3** of the project configuration file define the prompts that guide AI models in extracting targeted information. This section is central to a review project, with prismAId’s robust design enabling the tool’s Open Science benefits.

The **`[prompt]`** section breaks down the prompt structure into essential components to ensure accurate data extraction and minimize potential misinterpretations.

### Rationale
- This section provides explicit instructions and context for the AI model.
- The prompt consists of structured elements: 
<div style="text-align: center;"> 
    <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/prompt_struct.png" alt="Prompt Structure Diagram" style="width: 90%;"> 
</div>

- Each component clarifies the model’s role, task, and expected output, reducing ambiguity.
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

**CoT Justification** generates a .txt file per manuscript, logging the model’s thought process, responses, and relevant passages. Example output:
```md
- **clustering**: "no" - The text does not mention any clustering techniques or grouping of data points based on similarities.
- **copulas**: "yes" - The text explicitly mentions the use of copulas to model the joint distribution of multiple flooding indicators (maximum soil moisture, runoff, and precipitation). "The multidimensional representation of the joint distributions of relevant hydrological climate impacts is based on the concept of statistical copulas [43]."
- **forecasting**: "yes" - The text explicitly mentions the use of models to predict future scenarios of flooding hazards and damage. "Future scenarios use hazard and damage data predicted for the period 2018–2100."
```

### Rate Limits

Model usage limits can be managed with two main parameters set in **[project.llm]** section of the project configuration:

- **`tpm_limit`**: Sets a maximum for tokens processed per minute.
- **`rpm_limit`**: Sets a maximum for requests per minute.

Defaults for both are `0`, meaning no delays are applied. For non-zero values, prismAId enforces delays to meet specified limits.

**Note**: Daily request limits are not automatically enforced, so manual monitoring is required for users with daily limits.


#### OpenAI Rate Limits 
**(August 2024, tier 1 users)**

<table class="table-spacing">
    <thead>
        <tr>
            <th style="text-align: left;">Model</th>
            <th style="text-align: right;">RPM</th>
            <th style="text-align: right;">RPD</th>
            <th style="text-align: right;">TPM</th>
            <th style="text-align: right;">Batch Queue Limit</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td style="text-align: left;">gpt-4o</td>
            <td style="text-align: right;">500</td>
            <td style="text-align: right;">-</td>
            <td style="text-align: right;">30,000</td>
            <td style="text-align: right;">90,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">gpt-4o-mini</td>
            <td style="text-align: right;">500</td>
            <td style="text-align: right;">10,000</td>
            <td style="text-align: right;">200,000</td>
            <td style="text-align: right;">2,000,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">gpt-4-turbo</td>
            <td style="text-align: right;">500</td>
            <td style="text-align: right;">-</td>
            <td style="text-align: right;">30,000</td>
            <td style="text-align: right;">90,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">gpt-3.5-turbo</td>
            <td style="text-align: right;">3,500</td>
            <td style="text-align: right;">10,000</td>
            <td style="text-align: right;">200,000</td>
            <td style="text-align: right;">2,000,000</td>
        </tr>
    </tbody>
</table>


#### GoogleAI Rate Limits 
**(October 2024)**

**Free Tier**:
<table class="table-spacing">
    <thead>
        <tr>
            <th style="text-align: left;">Model</th>
            <th style="text-align: right;">RPM</th>
            <th style="text-align: right;">RPD</th>
            <th style="text-align: right;">TPM</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td style="text-align: left;">Gemini 1.5 Flash</td>
            <td style="text-align: right;">15</td>
            <td style="text-align: right;">1,500</td>
            <td style="text-align: right;">1,000,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">Gemini 1.5 Pro</td>
            <td style="text-align: right;">2</td>
            <td style="text-align: right;">50</td>
            <td style="text-align: right;">32,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">Gemini 1.0 Pro</td>
            <td style="text-align: right;">15</td>
            <td style="text-align: right;">1,500</td>
            <td style="text-align: right;">32,000</td>
        </tr>
    </tbody>
</table>

**Pay-as-you-go**:
<table class="table-spacing">
    <thead>
        <tr>
            <th style="text-align: left;">Model</th>
            <th style="text-align: right;">RPM</th>
            <th style="text-align: right;">RPD</th>
            <th style="text-align: right;">TPM</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td style="text-align: left;">Gemini 1.5 Flash</td>
            <td style="text-align: right;">2000</td>
            <td style="text-align: right;">-</td>
            <td style="text-align: right;">4,000,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">Gemini 1.5 Pro</td>
            <td style="text-align: right;">1000</td>
            <td style="text-align: right;">-</td>
            <td style="text-align: right;">4,000,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">Gemini 1.0 Pro</td>
            <td style="text-align: right;">360</td>
            <td style="text-align: right;">30,000</td>
            <td style="text-align: right;">120,000</td>
        </tr>
    </tbody>
</table>

#### Cohere Rate Limits
Cohere production keys have no limit, but trial keys are limited to 20 API calls per minute. 

#### Anthropic Rate Limits 
**(November 2024, tier 1 users)**
<table class="table-spacing">
    <thead>
        <tr>
            <th style="text-align: left;">Model</th>
            <th style="text-align: right;">RPM</th>
            <th style="text-align: right;">TPM</th>
            <th style="text-align: right;">TPD</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td style="text-align: left;">Claude 3.5 Sonnet</td>
            <td style="text-align: right;">50</td>
            <td style="text-align: right;">40,000</td>
            <td style="text-align: right;">1,000,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">Claude 3.5 Haiku</td>
            <td style="text-align: right;">50</td>
            <td style="text-align: right;">50,000</td>
            <td style="text-align: right;">5,000,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">Claude 3 Opus</td>
            <td style="text-align: right;">50</td>
            <td style="text-align: right;">20,000</td>
            <td style="text-align: right;">1,000,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">Claude 3 Sonnet</td>
            <td style="text-align: right;">50</td>
            <td style="text-align: right;">40,000</td>
            <td style="text-align: right;">1,000,000</td>
        </tr>
        <tr>
            <td style="text-align: left;">Claude 3 Haiku</td>
            <td style="text-align: right;">50</td>
            <td style="text-align: right;">50,000</td>
            <td style="text-align: right;">5,000,000</td>
        </tr>
    </tbody>
</table>

**Note**: To ensure adherence to provider limits, users should manually set the lowest applicable `tpm` and `rpm` values in the configuration, as prismAId does not enforce automatic checks.

### Cost Minimization
In **Section 1** of the project configuration:
- **`model`**: Leaving this field empty (`''`) enables automatic selection of the most cost-efficient model from the chosen provider. This may result in varying models for manuscripts based on length and token limits.

#### How Costs are Computed
- **Token Libraries**: prismAId uses libraries specific to each provider:
  - OpenAI’s cost estimation uses the [Tiktoken library](https://github.com/pkoukk/tiktoken-go).
  - Google’s token estimation uses the [CountTokens API](https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/count-tokens).
  - Cohere uses its [API](https://docs.cohere.com/docs/rate-limits).
  - Anthropic approximates token counts via OpenAI’s tokenizer.

Concise prompts are cost-efficient. Check costs on the provider dashboards: [OpenAI](https://platform.openai.com/usage), [Google AI](https://console.cloud.google.com/billing), [Cohere](https://dashboard.cohere.com/billing). 

**Note**: Cost estimates are indicative and may vary.

### Ensemble Review
Specifying multiple LLMs enables an 'ensemble' review, allowing result validation and uncertainty quantification. You can select multiple models from one or more providers, configuring each with specific parameters.

To set up an ensemble review in the `[project.llm]` section, for instance with models from four different providers, use:

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
```

### Zotero Integration
The tool can automatically download and process literature from your specified Zotero collections or groups. 

#### Configuration
To enable this, you must configure access credentials and group structure in the `[project.zotero]` section, for example:
```toml
[project.zotero]
user = "12345678"
api_key = "fdjkdfnjhfd4556"
group = "My Group/My Sub Collection"
```

To get your credentials, go to the [Zotero Settings](https://www.zotero.org/settings) page, navigate to the **Security** tab, and then to the **Applications** section. You will find your **user ID** and the button to generate an **API key**, as shown below:

<div style="text-align: center;">
    <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/zotero_user.png" alt="Zotero User ID" style="width: 600px;">
</div>

When creating a new API key, you must **enable** "Allow library access" and set the **permissions** to "Read Only" for all groups under "Default Group Permissions". You must also provide a name for the key, such as "test" or "prismaid".

<div style="text-align: center;">
    <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/zotero_apikey.png" alt="Zotero API Key" style="width: 600px;">
</div>

Once you have added your Zotero API credentials to your project configuration in the `[project.zotero]` section (fields `user` and `api_key`), you must specify the group or collection to review in the `group` field. This field uses a filesystem-like representation for the group and collection structure of your Zotero library. 

For instance, if you have a parent collection called "My Collection" and a nested sub-collection called "My Sub Collection" inside that parent collection, you should specify `"My Collection/My Sub Collection"` for the `group` field. Similarly, if you have a group called "My Group" and within that a collection called "My Sub Collection", you should specify `"My Group/My Sub Collection"` for the `group` field.

All PDFs in the selected collection or group will be copied into a `zotero` subdirectory within the directory you specified in the `[project.configuration]` section to store the `results_file_name`. Then, **prismAId** will convert them into text files and run the review process. 

The manuscript files are stored locally and are available for inspection and further cleaning and analysis without the need to connect to the Zotero API again.

#### Review Workflow Integration
Zotero is a powerful and open-source reference management system designed to help you store, organize, and share your literature. You can structure your manuscripts and references using either **collections** or **groups**.

- **Collections** are private and accessible only to the user who creates them. For step-by-step instructions on creating collections, refer to the [University of Ottawa Library's guide](https://uottawa.libguides.com/how_to_use_zotero/create_collections).

- **Groups** allow multiple users to access and collaborate on shared references, making them ideal for teamwork and collaborative research. To learn how to create a group, follow the [University of Ottawa Library's guide](https://uottawa.libguides.com/how_to_use_zotero/groups).

In the workflow of a systematic literature review, following any protocol, a Zotero collection or group is the perfect place to store the downloaded manuscripts after identifying them through literature search engines and a carefully defined selection query.

The integration of Zotero with **prismAId** supports the next step in the workflow: manuscripts are automatically converted and then passed to LLMs for analysis and information extraction.

Once the literature to be reviewed is defined, the Zotero integration only needs to be activated once, as all manuscripts are downloaded and stored in the `zotero` subdirectory. Subsequent analyses and refinements can be performed on the downloaded texts without requiring further connections to the Zotero API. 

To disable the Zotero integration, simply leave its fields empty in the `[project.zotero]` section of the project configuration.

#### Warning
**<span class="blink">ATTENTION</span>**: The Zotero integration automatically converts PDFs into text using the same methods as those activated by the `input_conversion` field of `[project.configuration]`. However, due to the inherent limitations of the PDF format, these conversions might be imperfect. **Therefore, both for `input_conversion` of PDF documents and Zotero integration, please manually check any converted manuscripts for completeness before further processing.**


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>