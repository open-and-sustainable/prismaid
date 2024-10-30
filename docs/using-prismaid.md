---
title: Using prismAId
layout: default
---

# Using prismAId

Prepare a project configuration file in [TOML](https://toml.io/en/), following the three-section structure, explanations, and recommendations provided in the [`template.toml`](https://github.com/open-and-sustainable/prismaid/blob/main/projects/template.toml) and below. Alternatively, you can use the terminal-based initialization option (`-init` in binaries) or the web-based tool on the [Review Configurator](review-configurator) page.

**Section 1**, introduced below, focuses on essential project settings. **Sections 2** and **3** cover **prompt design** and follow in sequence, while **advanced features** in Section 1 are discussed at the end of this page.

## Section 1: 'Project' Details

### Project Information:
```toml
[project]
name = "Use of LLM for Systematic Review"
author = "John Doe"
version = "1.0"
```
- The `[project]` section contains basic project information:
  - `name`: Project title.
  - `author`: Project author.
  - `version`: Configuration version.

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
- `[project.configuration]` specifies execution settings:
  - `input_directory`: Location of `.txt` files for review.
  - `input_conversion` : Non-active if left empty (default) or key removed. Enable with `pdf`, `docx`, `html`, or as a comma-separated list (e.g., `pdf,docx`).
  - `results_file_name`: Path to save results.
  - `output_format`: `csv` or `json`.
  - `log_level`: Sets log detail:
    - `low`: Minimal logging, essential output only (default).
    - `medium`: Logs details sent to stdout.
    - `high`: Logs are saved in a file.
  - `duplication`:  Controls review duplication for debugging:
    - `no`: Default.
    - `yes`: Files in the input directory are duplicated, reviewed, and removed before the program concludes.
  - `cot_justification`: Adds justification logs:
    - `no`: Default.
    - `yes`: Logs justification per manuscript, saved in the same directory.
  - `summary`: Enables summary logging:
    - `no`: Deafult.
    - `yes`: A summary is generated for each manuscript and saved in the same directory.

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
- `[project.llm]` specifies model configurations for review execution. At least one model is required. When multiple models are configured, results will represent an 'ensemble' analysis.

The `[project.llm.#]` fields manage LLM usage:
- `provider`:  Supported providers are `OpenAI`, `GoogleAI`, `Cohere`, and `Anthropic`.
- `api_key`: Define project-specific keys here, or leave empty to default to environment variables.
- `model`: select model:
    - Leave blank `''` for cost-efficient automatic model selection.
    - **OpenAI**: Models include `gpt-4o-mini`, `gpt-4o`, `gpt-4-turbo`, `gpt-3.5-turbo`.
    - **GoogleAI**: Choose from `gemini-1.5-flash`, `gemini-1.5-pro`, `gemini-1.0-pro`.
    - **Cohere**: Options are `command-r-plus`, `command-r`, `command-light`, `command`.
    - **Anthropic**: Includes `claude-3-5-sonnet`, `claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku`.
- `temperature`: Controls response variability (range: 0 to 1 for most models); lower values increase consistency.
- `tpm_limit`: Defines maximum tokens per minute. Default is `0` (no delay). Use a non-zero value based on your provider TPM limits (see Rate Limits in Advanced Features below).
- `rpm_limits`: Sets maximum requests per minute. Default is `0` (no limit). See provider’s RPM restrictions in Advanced Features below.

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
**Section 2 and 3** of the project configuration file define the prompts used to run the generative AI models to extract the information researchers are looking for. This is the key of a review project andthe prismAId robust approach to this part enables the many [Open Science](open-science) advantages provided by the tool.


The `[prompt]` section is aimed at defining the building blocks of the prompt, ensuring high accuracy in information extraction and minimizing hallucinations and misinterpretations.

#### Logic of the Prompt Section
- The prompt section allows the user providing clear instructions and context to the AI model.
- The prompt structure is made of these blocks:
<div style="text-align: center;">
    <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/prompt_struct.png" alt="Prompt Structure Diagram" style="width: 90%;">
</div>
- It ensures that the model understands the role it needs to play, the task it needs to perform, and the format of the expected output.
- By providing definitions and examples, it minimizes the risk of misinterpretation and improves the accuracy of the information extracted.
- A failsafe mechanism is included to prevent the model from forcing answers when information is not available.

```toml
[prompt]
persona = "You are an experienced scientist working on a systematic review of the literature."
task = "You are asked to map the concepts discussed in a scientific paper attached here."
expected_result = "You should output a JSON object with the following keys and possible values: "
failsafe = "If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value."
definitions = "'Interest rate' is the percentage charged by a lender for borrowing money or earned by an investor on a deposit over a specific period, typically expressed annually."
example = ""
```

#### Examples and Explanation of Entries
- `persona`:
  - "You are an experienced scientist working on a systematic review of the literature."
  - Personas help in setting the expectation on the model's role, providing context for the responses.
- `task`:
  - "You are asked to map the concepts discussed in a scientific paper attached here."
  - This entry defines the specific task the model needs to accomplish.
- `expected_result`:
  - "You should output a JSON object with the following keys and possible values: "
  - This introduces the expected output format, specifying that the result should be a JSON object with particular keys and values.
- `failsafe`:
  - "If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value."
  - This entry provides a fail-safe mechanism to avoid forcing answers when the required information is not present, ensuring accuracy and avoiding misinterpretation.
- `definitions`:
  - "'Interest rate' is the percentage charged by a lender for borrowing money or earned by an investor on a deposit over a specific period, typically expressed annually."
  - This allows for defining specific concepts to avoid misconceptions, helping the model understand precisely what is being asked.
- `example`:
  - ""
  - This is an opportunity to provide an example of the desired output, further reducing the risk of misinterpretation and guiding the model towards the correct response.

## Section 3: 'Review' Details
The `[review]` section is focused on defining the information to be extracted from the text. It outlines the structure of the JSON file to be returned by the LLM, specifying the keys and possible values for the extracted information.

#### Logic of the Review Section
- The review section defines the knowledge map that the model needs to fill in, guiding the extraction process.
- Each review item specifies a key, representing a concept or topic of interest, and possible values that the model can assign to that key.
- This structured approach ensures that the extracted information is consistent and adheres to the predefined schema.
- There can be as many review items as needed.

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
#### Examples and Explanation of Entries
- `[review]`:
  - This section header indicates the beginning of the review items configuration, which defines the structure of the knowledge map.
- `[review.1]`:
  - Defines the first item to be reviewed.
  - `key`: "interest rate"
    - The concept or topic to be extracted.
  - `values`: [""]
    - Possible values for this key. An empty string indicates that any value can be assigned.
- `[review.2]`:
  - Defines the second item to be reviewed.
  - `key`: "regression models"
    - The concept or topic to be extracted.
  - `values`: ["yes", "no"]
    - The key "regression models" can take either "yes" or "no" as its value, providing a clear binary choice.
- `[review.3]`:
  - Defines the third item to be reviewed.
  - `key`: "geographical scale"
    - The concept or topic to be extracted.
  - `values`: ["world", "continent", "river basin"]
    - The key "geographical scale" can take one of these specific values, indicating the scale of the geographical analysis.


## Advanced Features

### Debugging & Validation
In **Section 1** of the project configuration, there are three parameters supporting the devleopment of projects and testing of prompt configurations.
They are:
  - `log_level`: [`low`], `medium`, or `high`.
  - `duplication`: [`no`], or `yes`.
  - `cot_justification`:  [`no`], or `yes`.
First, if debuggin level is higher than low all API responses can be inspected in details. This means that besides output files, users will be able to access, either on terminal (stdout - `log_level`: `medium`) or in a log file (`log_level`: `high`), the complete reponses and eventual errors from the API and the prismAId execution.

Second, duplication makes possible to test whether a prompt definition is clear enough. In fact, if running twice the same prompt generates different ouptut it is very likely that the prompt is not deifning the model reviewing task clearly enough. Setting `duplication`: `yes` and then checking if answers differ in the two analyses of the same manuscript is a good way to assess whether the prompt is clear enough to be used for the review project. 

Duplicating manuscripts increases the cost of the project run, but the total cost presented at the beginning of the analysis is updated accordingly to let researchers assess the cost to be incurred. Hence, for instance, with Google AI as provider and Gemini 1.5 Flash model, without duplication:
```bash
Unless you are using a free tier with Google AI, the total cost (USD - $) to run this review is at least: 0.0005352
This value is an estimate of the total cost of input tokens only.
Do you want to continue? (y/n): y
Processing file #1/1: lit_test
```
With duplication active:
```bash
Unless you are using a free tier with Google AI, the total cost (USD - $) to run this review is at least: 0.0010704
This value is an estimate of the total cost of input tokens only.
Do you want to continue? (y/n): y
Processing file #1/2: lit_test
Waiting... 30 seconds remaining
Waiting... 25 seconds remaining
Waiting... 20 seconds remaining
Waiting... 15 seconds remaining
Waiting... 10 seconds remaining
Waiting... 5 seconds remaining
Wait completed.
Processing file #2/2: lit_test_duplicate
```

Third, in order to assess if the prompt definition are not only clear but also effective in extracting the information the researcher is looking for, it is is possible to use `cot_justification`: `yes`. This will create an output `.txt` for each manuscript containing the chain of thought (CoT) justification for the answer provided. Technically, the justification is provided by the model in the same chat as the answer, and right after it.

The ouput in the justification output reports the information requested, the answer provided, the modle CoT, and eventually the relevant sentences in the manuscript reviewd, like in:
```md
- **clustering**: "no" - The text does not mention any clustering techniques or grouping of data points based on similarities.
- **copulas**: "yes" - The text explicitly mentions the use of copulas to model the joint distribution of multiple flooding indicators (maximum soil moisture, runoff, and precipitation). "The multidimensional representation of the joint distributions of relevant hydrological climate impacts is based on the concept of statistical copulas [43]."
- **forecasting**: "yes" - The text explicitly mentions the use of models to predict future scenarios of flooding hazards and damage. "Future scenarios use hazard and damage data predicted for the period 2018–2100."

```

### Rate Limits
We enforce usage limits for models through two primary parameters specified in **Section 1** of the project configuration:

- **`tpm_limit`**: Defines the maximum number of tokens that the model can process per minute.
- **`rpm_limit`**: Specifies the maximum number of requests that the model can handle per minute.

For both parameters, a value of `0` is the default and is used if the parameter is not specified in the configuration file. The default value has a special meaning: no delay will be applied. However, if positive numbers are provided, the algorithm will compute delays and wait times between requests to the API accordingly.

Please note that we **do not support automatic enforcement of daily request limits**. If your usage tier includes a maximum number of requests per day, you will need to monitor and manage this limit manually.

On [OpenAI](https://platform.openai.com/docs/guides/rate-limits/usage-tiers?context=tier-one), for example, as of August 2024 users in tier 1 are subject to the following rate limits:

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


On [GoogleAI](https://ai.google.dev/pricing), as of October 2024 **free of charge** users are subject to the limits:

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

while **pay-as-you-go** users are subject to:

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

In September 2024 Cohere does not impose rate limits on production keys but trial keys are limited to 20 API calls per minute (refer to the official [documentation](https://docs.cohere.com/docs/rate-limits)).

Anthropic Tier 1 users have the following rate limits:
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


**PLEASE NOTE**: If you choose the cost minimization approach described below you must report in the configuration file the smallest tpm and rpm limits of the models by the provider you selected. This is the only way to ensure respecting limits since there is no authomatic check on them by prismAId and the selected model varies because of number of tokens in requests and model use prices.

### Cost Minimization
In **Section 1** of the project configuration:
 - `model`: Determines the model to use. Options are:
    - Leave empty `''`
This feature allows to always automatically select the cheapest model for the job provided by the provider selected. Please note that this may mean that different manuscripts are analyzed by different models depending on the manuscript length.

#### How costs are computed
- The cost of using OpenAI models is calculated based on [tokens](https://help.openai.com/en/articles/4936856-what-are-tokens-and-how-to-count-them).
- prismAId utilizes a [library](https://github.com/pkoukk/tiktoken-go) to compute the input tokens for each single-shot prompt before actually executing the call using another [library](https://github.com/sashabaranov/go-openai). Based on the information provided by OpenAI, the cost of each input token for the different models is used to compute the total cost of the inputs to be used in the review. This estimated cost is presented to the user, allowing them to decide whether to proceed with the analysis and incur the associated cost.
- prismAId calls the Google CountTokens [API](https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/count-tokens) to compute the input tokens for each single-shot prompt before actually executing the call using a [library](https://github.com/google/generative-ai-go). Based on the information provided by Google AI, the cost of each input token for the different models is used to compute the total cost of the inputs to be used in the review.
- prismAId calls the Cohere API to compute the input tokens for each single-shot prompt before actually executing the call using a [library](https://github.com/cohere-ai/cohere-go/). Please note that different Cohere models are trained with different tokenizers. This means also that the same prompt may be transformed into different number of input tokens depending on the model used. Based on the information provided by Cohere, the cost of each input token for the different models is used to compute the total cost of the inputs to be used in the review.
- Anthropic does not release the tokenizer nor an API free endpoint for counting input tokens. Following suggestions from their own Anthropic models, prismAId estimate the number of input tokens using the OpenAI tokenizer.
- Concise but complete prompts are both cost-effective and efficient in information extraction. Unnecessary text increases costs and may introduce noise, negatively affecting the performance of AI models. While additional explanations and definitions in the prompt engineering part may seem superfluous, they are generally limited in size and do not significantly impact costs.
- By using a project API key, it is possible to track the cost of each project on the OpenAI [dashboard](https://platform.openai.com/usage), the Google AI [dashboard](https://console.cloud.google.com/billing/), or the Cohere [dashboard](https://dashboard.cohere.com/billing).
- **The cost assessment function is indicative.**
  - We strive to maintain up-to-date data for cost estimation, though our estimations currently pertain only to the input aspect of AI model usage. As such, we cannot guarantee precise assessments.
  - Tests should be conducted first, and costs should be estimated more precisely by analyzing the data from the OpenAI [dashboard](https://platform.openai.com/usage) or the Google AI [dashboard](https://console.cloud.google.com/billing/).

### Ensemble Review
By specifying more than one LLM you obtain multiple results and hence an 'ensemble' review in which you can validate results and quantify uncertainties. Multiple LLMs can be selected within the same provider or across providers and for each of them users can specify specific parameters. 

Ensemble reviewing is configured in the `[project.llm]` section of the project configuration file. For instance, to get results from 4 models, each one from a different provider, users can specify as in the `template.toml` configuration provided with prismAId:
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

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>