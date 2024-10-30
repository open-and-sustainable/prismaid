---
title: Research & Development
layout: default
---

# Research & Development


### Scope
- prismAId is a software tool designed to leverage the capabilities of Large Language Models (LLMs) or AI Foundation models in understanding text content for conducting systematic reviews of scientific literature.
- It aims to make the systematic review process easy, requiring no coding.
- prismAId is designed to be much faster than traditional human-based approaches, offering also a high-speed software implementation.
- It ensures full replicability. Unlike traditional methods, which rely on subjective interpretation and classification of scientific concepts, prismAId addresses the primary issue of replicability in systematic reviews.
- Though running reviews with prismAId incurs costs associated with using AI models, these costs are limited and lower than alternative approaches such as fine-tuning models or developing ad hoc on-premises models, which also complicate replicability. Indicatively, the cost of extracting information from a paper, as of today, can vary between a quarter of a cent to 10 cents (USD or EUR).
- Beneficiaries: Any scientist conducting a literature review or meta-analysis for developing projects, proposals, or manuscripts.

### Description of Underlying Mechanism
- How LLMs work:
  - LLMs (Large Language Models) are AI models trained on vast amounts of text data to understand and generate human-like text.
  - These models can perform a variety of language tasks such as text completion, summarization, translation, and more.  
- Data flow and processing steps:
  - Contemporary state-of-the-art LLMs offer subscription-based API access.
  - While foundation models can be used in various ways, prismAId focuses solely on prompt engineering or prompting.
  - Prompt engineering involves crafting precise prompts to extract specific information from the AI model via the API.
  - prismAId simplifies the creation of rigorous and replicable prompts to extract information through the AI model API.
  - The data flow of prismAId is embedded in protocol-based approaches to reviews:
    - Initially, there is a selection of literature to be analyzed through detailed steps. These are defined by protocols and are easily replicable. 
    - Next, the content of these papers is classified, which is where prismAId comes into play.
  - prismAId allows for parsing the selected literature and extracting the information the researcher is interested in. AI models do not know fatigue and are much faster than humans.
  - prismAId users define the information extraction tasks using the prompt engineering template provided by prismAId.
  - prismAId utilizes multiple single-shot prompt API calls to individually parse each scientific paper.
  - prismAId processes the JSON files returned by the AI model, converting the extracted information into the user-specified format.
  - To facilitate cost management, prismAId tokenizes each single-shot prompt and estimates the execution cost, allowing users to understand the total review cost before proceeding.

  


## Contributing
We value community contributions that help improve prismAId. Whether you're fixing bugs, adding features, or improving documentation, your input is welcome:
- **Branching Strategy**: Please create a new branch for each set of related changes and submit a pull request through GitHub.
- **Code Reviews**: All submissions undergo a thorough review process to maintain code quality and consistency.
- **Community Engagement**: Engage with us through GitHub [issues](https://github.com/open-and-sustainable/prismaid/issues) and [discussions](https://github.com/open-and-sustainable/prismaid/discussions) for feature requests, suggestions, or any queries related to project development.

For detailed guidelines on contributing, please refer to our [`CONTRIBUTING.md`](CONTRIBUTING.md) and [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

## Software Stack and Approach
prismAId is built using the Go programming language, known for its simplicity and efficiency in handling concurrent operations. We prioritize staying up-to-date with the latest stable releases of Go to leverage the newest features and improvements.

### Development Environment Setup
To facilitate the development and testing of prismAId, templates for configuring VSCodium (or Visual Studio Code) are provided. These templates include predefined settings and extensions that enhance the development experience, ensuring consistency across different setups.
- **Accessing Templates**: You can find the configuration templates in the [`cmd` directory](https://github.com/open-and-sustainable/prismaid/tree/main/cmd) of our source repository. 

#### Using the Templates
1. **Clone the Repository**: Start by cloning the prismAId repository to your local machine.
2. **Open with VSCodium/VSCode**: Open the directory within VSCodium or VSCode.
3. **Copy the .json Files**: Copy them in a newly created `vscode`directory on the root of the project.
4. **Remove the .template extension**: Change the file names and follow the instructions within the files.
5. **Ignore the Files in GIT**: Add the files to your local .gitignore to avoid sharing secrets and other private information.

### Architecture
Our architecture is designed to be robust yet simple, ensuring that the tool remains accessible to both technical and non-technical users:
- **Self-Contained Binaries**: prismAId is distributed as self-contained binaries, which means all necessary libraries and dependencies are packaged together. This approach eliminates the need for external installations and simplifies the setup process.
- **Cross-Platform Compatibility**: Compatible with major operating systems such as Windows, MacOS, and Linux, ensuring that prismAId can be used in diverse environments.

### Development Philosophy
- **Open Source**: We embrace an open-source model, encouraging community contributions and transparency in development.
- **Continuous Integration/Continuous Deployment (CI/CD)**: We utilize CI/CD pipelines to maintain high standards of quality and reliability, automatically testing and deploying new versions as they are developed.


## Contribution to Open Science
prismAId supports Open Science in many aspects:

### Transparency and Reproducibility
   - prismAId ensures transparency in the analysis process, making it easier for other researchers to understand, replicate, and validate the findings.
   - prismAId removes the subjectivity of individual interpretations, making systematic literature reviews 100% reproducible.
   - As a software tool, prismAId helps maintain detailed logs and records of the analysis process, enhancing reproducibility.

### Accessibility and Collaboration
   - prismAId facilitates collaboration among researchers by providing an open tool that makes it possible to share data, analysis methods, and results.
   - prismAId is open source and openly licensed. Making analysis tools openly available promotes wider participation and contribution from the scientific community.
   - prismAId releases and their source code are archived on [Zenodo](https://zenodo.org/doi/10.5281/zenodo.11210796), ensuring long-term accessibility and referencability. This helps address legacy issues for analyses conducted using prismAId, making the tool and its results open, replicable, and understandable over the long run.

### Efficiency and Scalability
   - prismAId can handle large volumes of data efficiently, making the analysis phase quicker and more scalable compared to traditional methods.
   - This efficiency supports open science by allowing more comprehensive and timely reviews, reducing the time society needs to properly 'digest' scientific innovations.

### Quality and Accuracy
   - By explicitly defining each piece of reviewed information through prompt configurations, prismAId enhances the quality and accuracy of data extraction and analysis, leading to more reliable systematic reviews.
   - Publishing prismAId project configuration files ensures that approaches, biases, and methods are visible and verifiable by the broader research community. By doing so, they are also reusable and extendable.

### Ethical Considerations and Bias Reduction
   - Using prismAId means explicitly addressing biases and incorporating ethical considerations in its design and implementation to minimize biases in data analysis.
   - prismAId enables open science approaches, ensuring full transparency on ethical standards, with community oversight and input helping to identify and mitigate potential biases.

### Scientific Innovation
   - prismAId promotes scientific innovation by fully formalizing the analysis process, creating standardized procedures that ensure consistency and accuracy in systematic reviews.
   - This formalization makes methods and procedures reusable and extendable, allowing researchers to build upon previous analyses and adapt methods to new contexts.
   - By facilitating incremental discoveries, prismAId supports the cumulative advancement of science, where each study contributes to a larger body of knowledge.
   - prismAId's commitment to open science principles ensures that all tools, methods, and data are openly accessible, fostering collaboration and rapid dissemination of innovations.



<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>