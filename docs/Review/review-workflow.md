---
title: Review Workflow
layout: default
---

# Review Workflow

---

## Reasons to Use

- **Objective**: prismAId leverages Large Language Models (LLMs) for systematic scientific literature reviews, making them accessible and efficient without coding.
- **Speed**: Faster than traditional methods, prismAId provides high-speed software for systematic reviews.
- **Replicability**: Addresses the challenge of consistent, unbiased analysis, countering the subjective nature of human review.
- **Cost**: More economical than custom AI solutions, with review costs typically between $0.0025 and $0.10 per paper.
- **Flexibility**: Available through multiple programming languages (Go, Python, R, Julia) and as standalone tools.
- **Audience**: Suitable for scientists conducting literature reviews, meta-analyses, project development, and research proposals.

## Review Process Overview

The prismAId toolkit supports a comprehensive systematic review process:

1. **Design and Register**: Define your review protocol and register it if applicable
2. **Literature Identification**: Design and run queries on repositories
3. **Literature Acquisition**: Download papers using the prismAId Download tool
4. **Format Conversion**: Convert papers to text using the prismAId Convert tool
5. **Screening**: Screen manuscripts for duplications and relevance
6. **Project Configuration**: Set up a prismAId review project using the configurator
7. **Analysis**: Execute the review and extract structured information
8. **Results Processing**: Examine and synthesize the extracted data

This workflow integrates with established protocols like [PRISMA 2020](https://doi.org/10.1136/bmj.n71) while using AI to automate the most time-consuming aspects.

## Using prismAId Tools

### 1. Screening Tool

The Screening tool filters manuscripts after initial search but before full-text download, saving time and resources:

```bash
# Using the binary
./prismaid --screening screening_config.toml

# Using Go
import "github.com/open-and-sustainable/prismaid"
prismaid.Screening(tomlConfigString)

# Using Python
import prismaid
with open("screening.toml", "r") as file:
    toml_config = file.read()
prismaid.screening(toml_config)

# Using R
library(prismaid)
toml_content <- paste(readLines("screening.toml"), collapse = "\n")
Screening(toml_content)

# Using Julia
using PrismAId
toml_config = read("screening.toml", String)
PrismAId.screening(toml_config)
```

The Screening tool applies multiple filters:
- **Deduplication**: Removes duplicate manuscripts
- **Language**: Filters by accepted languages
- **Article Type**: Excludes reviews, editorials, letters, etc.
- **Topic Relevance**: Scores manuscripts for relevance to your research topic

### 2. Download Tool

The Download tool acquires full-text papers for manuscripts that passed screening:

#### From URL List:
```bash
# Using the binary
./prismaid -download-URL list_of_urls.txt

# Using Go
import "github.com/open-and-sustainable/prismaid"
prismaid.DownloadURLList("list_of_urls.txt")

# Using Python
import prismaid
prismaid.download_url_list("list_of_urls.txt")

# Using R
library(prismaid)
DownloadURLList("list_of_urls.txt")

# Using Julia
using PrismAId
PrismAId.download_url_list("list_of_urls.txt")
```

#### From Zotero:
For Zotero integration, you'll need your username, API key, and collection name:

```bash
# Using the binary (requires a TOML config file)
# First create a file zotero_config.toml with:
#   user = "your_username"
#   api_key = "your_api_key"
#   group = "Your Collection"
./prismaid -download-zotero zotero_config.toml

# Using Go
import "github.com/open-and-sustainable/prismaid"
prismaid.DownloadZoteroPDFs("username", "apiKey", "collectionName", "./papers")

# Using Python
import prismaid
prismaid.download_zotero_pdfs("username", "api_key", "collection_name", "./papers")

# Using R
library(prismaid)
DownloadZoteroPDFs("username", "api_key", "collection_name", "./papers")

# Using Julia
using PrismAId
PrismAId.download_zotero_pdfs("username", "api_key", "collection_name", "./papers")
```

### 3. Convert Tool

The Convert tool transforms downloaded documents into analyzable text:

```bash
# Using the binary (separate commands for each format)
./prismaid -convert-pdf ./papers
./prismaid -convert-docx ./papers
./prismaid -convert-html ./papers

# Using Go
import "github.com/open-and-sustainable/prismaid"
prismaid.Convert("./papers", "pdf,docx,html")

# Using Python
import prismaid
prismaid.convert("./papers", "pdf,docx,html")

# Using R
library(prismaid)
Convert("./papers", "pdf,docx,html")

# Using Julia
using PrismAId
PrismAId.convert("./papers", "pdf,docx,html")
```

**<span style="color: red; font-weight: bold;">IMPORTANT:</span>** Due to limitations in the PDF format, conversions might be imperfect. Always manually check converted manuscripts for completeness before further processing.

### 4. Review Tool

The Review tool processes systematic literature reviews based on TOML configurations:

```bash
# Using the binary
./prismaid -project your_project.toml

# Using Go
import "github.com/open-and-sustainable/prismaid"
prismaid.Review(tomlConfigString)

# Using Python
import prismaid
with open("project.toml", "r") as file:
    toml_config = file.read()
prismaid.review(toml_config)

# Using R
library(prismaid)
toml_content <- paste(readLines("project.toml"), collapse = "\n")
RunReview(toml_content)

# Using Julia
using PrismAId
toml_config = read("project.toml", String)
PrismAId.run_review(toml_config)
```

You can use the [Review Configurator](../review/review-configurator) web tool to easily create TOML configurations or use the `-init` flag with the binary.

## Information Extraction Mechanism

### LLMs Basics
- **How LLMs Work**:
  - Large Language Models (LLMs) are AI trained on extensive text data to understand and generate human-like text.
  - These models handle various language tasks like text completion, summarization, translation, and more.

- **Data Flow and Processing**:
  - Modern LLMs offer subscription-based API access, with prismAId focusing on prompt engineering to extract targeted information.
  - prismAId enables structured, replicable prompt creation for systematic reviews, simplifying rigorous data extraction.

### Data Flow
prismAId's workflow embeds protocol-based approaches:
  - **Literature Selection**: Based on defined protocols, ensuring replicability.
  - **Content Classification**: prismAId handles paper classification, parsing selected literature to extract user-defined information.
  - **API Calls & Cost Management**: prismAId sends single-shot prompts for each paper, processes AI-generated JSON files, and provides token-based cost estimates for informed decision-making.

## Best Practices

Follow these methodologies for effective systematic reviews:

1. **Prepare Literature Carefully**:
   - Remove unnecessary sections (references, abstracts, etc.) that don't contribute relevant information
   - Focus on primary sources; avoid including review articles unless necessary
   - Reducing content helps lower costs and may improve model performance

2. **Configure Reviews Effectively**:
   - Clearly define the information you're seeking, using examples where helpful
   - Avoid open-ended answers; define all possible answers the AI model can provide
   - For better results, run separate reviews for each piece of information you want to extract

3. **Test and Iterate**:
   - Start with a single paper test
   - When satisfied, test with a small batch
   - Only then proceed with the full literature set

4. **Documentation and Transparency**:
   - Include the project configuration (TOML file) in your paper's appendix
   - Properly cite prismAId [doi.org/10.5281/zenodo.11210796](https://doi.org/10.5281/zenodo.11210796)
   - Document your methodology thoroughly

5. **Learn From Examples**:
   - Review the [JOSS paper](https://doi.org/10.21105/joss.07616) and [extended introduction](https://doi.org/10.31222/osf.io/wh8qn)
   - Examine the [example configurations](https://github.com/open-and-sustainable/prismaid/tree/main/projects/test/configs) with sample [inputs](https://github.com/open-and-sustainable/prismaid/tree/main/projects/test/inputs) and [outputs](https://github.com/open-and-sustainable/prismaid/tree/main/projects/test/outputs)
   - Review [configuration templates](https://github.com/open-and-sustainable/prismaid/tree/main/projects/templates) for review, screening, and Zotero tools

## Technical FAQs

1. **Q: Will I always get the same answer if I set the model temperature to zero?**<br>
   **A:** No, setting the temperature to zero doesn't guarantee identical answers. While it reduces randomness, these models are sparse mixture-of-experts systems that may still show probabilistic behavior. This is especially true when prompts approach the token limit or when content shifts within the model's attention window. Using lower temperature settings is still recommended to minimize variability. For critical reviews, consider testing robustness by modifying prompt structure and running multiple trials.<br>
   **Further reading:** [https://doi.org/10.48550/arXiv.2308.00951](https://doi.org/10.48550/arXiv.2308.00951)

2. **Q: Does noise (or how much information is hidden in the literature to be reviewed) have an impact?**<br>
   **A:** Yes, noise and hidden information significantly impact extraction quality. Higher noise levels make accurate information extraction more challenging. During configuration development, thorough testing helps identify effective prompt structures that minimize noise and enhance clarity, improving the ability to find critical information even when obscured.<br>
   **Further reading:** [https://doi.org/10.48550/arXiv.2404.08865](https://doi.org/10.48550/arXiv.2404.08865)

3. **Q: What happens if the literature to be reviewed says something different from the data used to train the model?**<br>
   **A:** This presents an unavoidable challenge, as we lack full transparency about model training data. When conflicts exist, extraction may be biased or transformed by the training data. While focusing on direct results from the prompt helps minimize these risks, they cannot be eliminated entirely. These biases require careful consideration, especially for topics without long-term consensus. However, prismAId's replicability and experimental capabilities provide tools for checking these biases—an advantage over traditional human reviews.<br>
   **Further reading:** [https://doi.org/10.48550/arXiv.2404.08865](https://doi.org/10.48550/arXiv.2404.08865)

4. **Q: Are there reasoning biases I should expect when analyzing literature with generative AI models?**<br>
   **A:** Yes, AI models trained on human texts can replicate human reasoning biases, potentially leading to false information extraction if prompts steer them that way. The best strategy is crafting precise, neutral, and unbiased prompts. prismAId's structured approach helps minimize these issues, but careful prompt design remains essential.<br>
   **Further reading:** [https://doi.org/10.1038/s43588-023-00527-x](https://doi.org/10.1038/s43588-023-00527-x)

5. **Q: Is it always better to analyze literature by extracting one piece of information at a time (one piece of information per prismAId project)?**<br>
   **A:** Yes, separate projects for each information piece is highly effective. This approach allows tailored prompts for more accurate answers. Combining multiple information retrieval tasks requires longer prompts that can confuse the AI model. The only drawback is increased cost, as separating questions approximately doubles the API expense. Therefore, quality is primarily constrained by budget.<br>
   **Further reading:** [OpenAI API Prices](https://openai.com/api/pricing/) - [https://doi.org/10.48550/arXiv.2404.08865](https://doi.org/10.48550/arXiv.2404.08865)

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
