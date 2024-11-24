---
title: 'An Introduction to `prismAId`: Open-Source and Open Science AI for Advancing Information Extraction in Systematic Reviews'
tags:
  - systematic literature review
  - generative AI
  - Go
  - Python
  - R
  - Julia
  - Open Science
authors:
  - name: Riccardo Boero
    orcid: 0000-0002-7468-9096
    affiliation: 1
affiliations:
 - name: NILU, the Climate and Environmental Research Institute
   index: 1
   ror: 00q7d9z06
date: 24 November 2024
bibliography: paper.bib
---

# Summary

`prismAId` is an open-source tool designed to streamline systematic literature reviews by leveraging generative AI models for information extraction. It offers an accessible, efficient, and replicable method for extracting and analyzing data from scientific literature, eliminating the need for coding expertise. Supporting various review protocols, including PRISMA 2020, `prismAId` is distributed across multiple platforms -- Go, Python, Julia, R -- and provides user-friendly binaries compatible with Windows, macOS, and Linux. The tool integrates with leading large language models (LLMs) such as OpenAI's GPT series, Google's Gemini, Cohere's Command, and Anthropic's Claude, ensuring comprehensive and up-to-date literature analysis. `prismAId` facilitates systematic reviews, enabling researchers to conduct thorough, fast, and reproducible analyses, thereby advancing open science initiatives.

# Statement of Need

Systematic literature reviews play a crucial role in synthesizing research findings across various disciplines. However, traditional approaches are typically labor-intensive, time-consuming, and prone to subjectivity, making it challenging to manage large datasets while ensuring consistency and reproducibility. `prismAId` addresses these challenges by automating data extraction and analysis through advanced AI models, significantly reducing the time and effort required. Its compatibility with multiple programming languages and operating systems ensures broad accessibility, while integration with established review protocols like PRISMA 2020 guarantees methodological rigor. `prismAId` provides a user-friendly interface that requires no coding skills, enabling researchers from diverse backgrounds to perform comprehensive literature analyses efficiently.

# Introduction

A systematic literature review (SLR) is a structured method of synthesizing existing research to answer specific questions. Unlike traditional literature reviews, SLRs employ rigorous and transparent procedures to minimize bias and ensure reproducibility. This approach involves clearly defining research questions, systematically searching for relevant studies, critically appraising their quality, and synthesizing findings to provide comprehensive insights into a particular topic.

The importance of SLRs lies in their ability to consolidate vast amounts of information, offering evidence-based conclusions that inform practice, policy, and future research. By evaluating and integrating findings from multiple studies, SLRs help identify consistent patterns, resolve discrepancies, and highlight areas needing further investigation. This process not only advances scientific knowledge but also supports decision-making across various fields.

Many protocols exist for conducting SLRs [@schiavo_prospero_2019], including PRISMA 2020 [@page_prisma_2021], which inspired the name of our tool, as well as preregistration services and templates designed to standardize and support SLR processes.

Many recent contributions discuss the role that AI models can play in supporting SLRs. The integration of AI is predominantly highlighted in the screening phase, where machine learning algorithms are employed to filter and prioritize studies, significantly reducing the burden of manual work. However, the broader application of AI across the entire review process, including data extraction and risk of bias assessment, reveals both potential benefits and significant limitations.

The review by @blaizot_using_2022 provides an in-depth analysis of AI methods in SLRs, particularly in health sciences. The authors identify title and abstract screening as the primary area of application, where 73% of AI tools are concentrated. Machine learning algorithms assist in filtering studies, expediting the initial triage process. Despite the efficiency gains, the authors caution that current AI tools require substantial human input for validation, as fully autonomous screening remains limited. In data extraction, tools like RobotReviewer show promise but exhibit frequent errors, necessitating manual corrections. AI methods also attempt to automate risk of bias assessment, yet significant issues with accuracy and consistency demand human oversight. The authors conclude that while AI can enhance efficiency in certain stages of SLRs, challenges related to data quality, model transparency, and the need for expert intervention persist.

@van_de_schoot_open_2021 propose AI models, especially those using active learning, as tools to streamline the screening phase. These models prioritize relevant records based on human feedback, reducing the number of studies needing manual evaluation. However, AI models often struggle with complex and evolving inclusion criteria, especially when relevance depends on nuanced context. The selection phase also benefits from AI, as machine learning tools help predict the eligibility of full-text articles. However, reliance on simple text features may result in missed insights, making AI effective for initial triage but insufficient for detailed assessments.

In the context of updating systematic reviews, @van_de_schoot_open_2021 highlight AI’s potential to manage large volumes of new publications. AI models can flag relevant new studies based on learned patterns from previous decisions. However, models trained on outdated criteria may miss key studies, and there is a risk of reinforcing biases from historical data. Approaches like active learning depend heavily on the quality of initial labeled data, and methods such as TF-IDF or BERT embeddings, while effective, often lack domain-specific nuance, impacting reliability.

@dijk_artificial_2023 discuss the use of AI tools like ASReview, emphasizing their efficiency in the screening phase. Machine learning models rank articles based on predicted relevance, significantly reducing manual workload. However, the authors note that these gains are heavily dependent on the quality of training data and may be compromised if the input data is biased. Issues such as deduplication and reviewer influence on AI model learning are highlighted as potential sources of error. The authors call for safeguards like double screening and inter-reviewer checks to ensure reliability. In updating reviews, AI helps manage growing datasets, but there is a risk of missing important studies if the model does not adapt to changes in the research landscape. The authors argue for standardized guidelines from organizations like Cochrane and PRISMA to formalize the use of AI tools in SLR protocols, stressing the need for transparency and further evaluation.

@de_la_torre-lopez_artificial_2023 provide a comprehensive overview of AI in automating SLRs. The authors highlight that the conducting phase sees the most significant integration of AI, particularly in selecting primary studies through machine learning models like support vector machines and neural networks. These techniques can reduce the screening burden by up to 60%, yet their effectiveness hinges on the quality of training data. In the planning phase, AI is less commonly applied but has potential for optimizing search strategies through clustering and process mining. For the reporting phase, NLP and ontology-based systems show promise in data extraction and summarization, although current methods often lack transparency and interpretability, limiting their acceptance.

The article by @fabiano_how_nodate critically examines AI's potential to optimize the SLR process, identifying key contributions in screening and data extraction. Tools like Covidence and Rayyan.ai streamline study selection by ranking articles based on relevance, but they still require substantial human input. AI’s performance is limited by biases in training data, and the tools do not replicate the rigor of manual reviews. For data extraction, AI tools like RobotReviewer can generate PICO-based summaries, aiding synthesis, but automated summaries risk misinterpreting complex study details. @fabiano_how_nodate also discuss AI's role in risk of bias assessment and report writing, noting that AI methods lack the nuance needed for accurate evaluations, necessitating human validation. The authors emphasize concerns about transparency, reproducibility, and the risk of bias amplification, advocating for clearer guidelines on AI integration into review protocols.

`prismAId` approach is different and addresses two key issues in SLRs: the need to support open science practices and the potential of leveraging generative AI models for natural language processing. This leads to a focus on the analysis phase of SLRs, particularly in enhancing data extraction and developing a standard that ensures clear communication of research findings. By doing so, `prismAId` aims to improve the transparency, inspection, and understanding of the review process and results, as well as to boost the reproducibility and extension of scientific knowledge -- not just in terms of review content, but also the review methodology itself. Additionally, to support these goals, `prismAId` offers an accessible, code-free interface, lowering the barriers for users with limited technical skills.

# Features and Capabilities

`prismAId` is compatible with any systematic review protocol [@schiavo_prospero_2019], including the widely adopted PRISMA 2020 guidelines [@page_prisma_2021]. It focuses specifically on the analysis phase, which is a key concluding component of all systematic review protocols, alongside the reporting stage.

![Traditional approach to literature analysis vs. `prismAId`-supported review workflow.\label{fig:workflow}](workflow.png){width="50%"}

As illustrated in \autoref{fig:workflow}, `prismAId` plays a crucial role in the data extraction activities. The tool leverages generative AI models to analyze scientific documents, extracting information in a structured and quantifiable manner. Users only need to configure a review project by specifying the necessary parameters and providing the documents to be analyzed. `prismAId` then automates the entire extraction process, outputting a comprehensive database of results.

The approach taken by `prismAId` enhances the accuracy of information extraction by leveraging AI for natural language processing tasks that are traditionally performed by humans. By replacing manual efforts with computational methods, `prismAId` reduces subjectivity and the issues that arise from vague or incomplete definitions of concepts. Additionally, AI models maintain consistent performance without the effects of distraction or fatigue, further improving the reliability of the results.

Beyond accuracy, the computational nature of `prismAId` enables extremely fast SLRs and ensures full replicability of the analysis. The modular and project-based design allows for easy integration, expansion, and cumulative value, supporting the ongoing buildup of scientific knowledge through consistent and extendable SLRs.

Building on the tool's aim, `prismAId` includes the capability to preprocess and share manuscripts directly with AI models. To achieve this, `prismAId` provides features for converting files from PDF, DOCX, and HTML formats into plain text (TXT). This conversion is essential for effective text tokenization, which involves breaking down the text into smaller units (tokens) such as words or phrases. Tokenization is a fundamental step in natural language processing, enabling the AI models to parse and understand the document content.

Additionally, `prismAId` integrates with Zotero, a popular open source reference management software, allowing users to access shared collections and groups of scientific literature seamlessly. This integration simplifies document retrieval and enhances collaboration. The connection to Zotero is automatic once the API access is configured, as detailed in the `prismAId` documentation. This feature streamlines the workflow, making it easier for users to manage and analyze large sets of scientific manuscripts.

`prismAId` integrates with various large language models (LLMs) from the four main AI providers: OpenAI, Google Cloud AI, Anthropic, and Cohere. This includes access to state-of-the-art models such as many variants of GPT, Gemini, Claude, and Command. The integration is handled through native APIs, allowing for seamless access to advanced generative AI capabilities. Depending on project needs and available opportunities, support for cloud-based deployments of these models will be expanded. This variety ensures that users can leverage the most up-to-date technology while benefiting from different pricing options, making `prismAId` adaptable to a range of project budgets and requirements.

The main results dataset generated by `prismAId` can be exported in two formats: a tabular CSV and a nested JSON structure. These options ensure compatibility with a wide range of analysis tools and software. The CSV format is ideal for spreadsheet-based analysis and integration with statistical packages, while the JSON format supports more complex data structures and is suitable for use in web applications or advanced data processing pipelines. Together, these formats cover all potential use cases and enable seamless data import and interoperability.

`prismAId` includes advanced functions designed to support debugging and validation throughout the review process. One key feature is input duplication, which is particularly useful when testing on individual or small sets of manuscripts. This feature facilitates sensitivity analysis, allowing users to experiment with prompt design and other parameters, and evaluate their impact on the results. Another useful capability is the option to generate summaries of manuscripts, enabling manual, ad hoc verification of the extracted outcomes against the main messages of the original papers. Additionally, `prismAId` offers the option to request and store Chain of Thought (COT) justifications. Here, the AI models are prompted to explain their reasoning process and identify the specific parts of the manuscript that support their conclusions, providing greater transparency and accountability.

Moreover, `prismAId` supports ensemble reviews, a powerful feature that allows the same review to be run across different models, both within a single provider and across multiple providers. This approach enables the calculation of uncertainty estimates for the extracted answers, helping to assess the robustness of the results and providing a more comprehensive analysis.

In conclusion, `prismAId` offers multi-language support, including Go, Python, Julia, and R, along with cross-platform compatibility. This flexibility allows `prismAId` features to be integrated into larger workflows or newly developed pipelines across these programming environments. Additionally, the tool is designed to run smoothly on different operating systems and computing platforms (including Intel, AMD, and ARM 64-bit architectures), ensuring that users are not limited by their hardware or software setup. 


# Technical Architecture

`prismAId` is an open-source software project released under the AGPL-3.0 license, with the codebase hosted on GitHub [@prismaid] and preserved on Zenodo [@Boero_prismAId_-_Open]. Support is provided through GitHub issues, discussion threads, and Matrix chat rooms, ensuring a collaborative environment for troubleshooting and feature requests.

The underlying architecture of `prismAId` is built as a Go module, utilizing a pure Go implementation. This approach leverages the efficiency and concurrency capabilities of Go, while also integrating libraries that enable API access to various AI models. The use of native Go components enhances performance and portability, making the tool robust and compatible across different environments without relying on external dependencies.

![`prismAId` architecture and codebase
organization.\label{fig:codebase-architecture}](codebase-architecture.png){width="60%"}

The software architecture of prismAId is clearly reflected in the modular structure of its codebase, as shown in \autoref{fig:codebase-architecture}. The design is modular, with specific components handling different aspects of the tool's functionality. There is a section dedicated to input conversion and integration, which processes scientific literature in various formats, such as PDF, DOCX, and HTML, preparing it for further analysis. Another key part of the codebase focuses on the connection and configuration of external AI models, enabling API integration with a variety of large language models.

User interaction is supported by multiple components that include access to comprehensive documentation, making it easier for users to configure and navigate the tool. The architecture also incorporates robust support for multi-platform language and OS compatibility, ensuring smooth deployment across various environments, including different operating systems and hardware architectures.

![Core logical flow of `prismAId`.\label{fig:flowchart}](flowchart.png){width="50%"}

At the heart of the architecture lies the core module, where the logic of the workflows is integrated. This central part manages the main review rationale, coordinating tasks related to data extraction, analysis, and the generation of outputs, as represented in \autoref{fig:flowchart}.

# User Experience

The primary access to `prismAId` is through standalone binaries, compiled for Windows, macOS, and Linux, with support for both AMD64 (x86-64) and ARM64 architectures. These self-contained binaries allow users to run and manage review projects, as well as initialize project configuration files directly from the terminal. They do not require the installation of any additional software beyond terminal access, making them simple and convenient to use. The binaries can be easily downloaded from the release page on the project GitHub repository [@prismaid].

For users with more technical expertise, there are additional programmatic options for accessing `prismAId`. One option is to integrate the tool directly by using the Go module available on the GitHub repo. Alternatively, `prismAId` can be installed as a Python package via PyPI, as a Julia package from the General package registry, or as an R package through R-universe. These options offer more flexibility and are suitable for integrating `prismAId` into custom workflows and advanced scripting environments.

![`prismAId` use case diagram.\label{fig:use-case}](use-case.png){width="60%"}

The tool requires a single input: the project configuration file, which defines all the necessary parameters for conducting a systematic review. This file specifies key details, including the location of the manuscripts to be reviewed and the desired file name and directory for the results. The main steps for creating a project configuration file are illustrated in \autoref{fig:use-case}. Users need to define the input and output paths, select the AI models to be used, configure any advanced features, and specify both the prompt design and the information model for data extraction. This structured approach ensures that `prismAId` can efficiently manage the review process based on clear user-defined parameters.

The tool includes a template.toml file, which outlines and explains all the configuration options that can be used in a project configuration file. To assist users in creating a complete configuration, `prismAId` offers two interactive initializers. The first initializer is terminal-based and can be launched by running the binaries with the `-init` flag. This option provides a step-by-step guide directly in the terminal, helping users generate a draft configuration file.

The second initializer is web-based, accessible via the `prismAId` documentation website [@prismaid-doc]. This version offers a dynamic interface where users can specify all configuration components interactively. It also provides a real-time preview of the configuration file in the web browser, with the option to download the finalized file for further editing or direct use with `prismAId`.

![Prompt blocks in `prismAId`.\label{fig:prompt}](prompt.png){width="60%"}

A key component in ensuring the replicability of SLRs conducted with `prismAId` is the project configuration file itself. When this file is published alongside the review report, it enables others to replicate the original analysis, extend the review to cover new bodies of knowledge, or update the existing review with additional literature. Beyond replicability, `prismAId` also emphasizes quality support, especially through the standardization of prompt design, which guides users in crafting robust and accurate prompts, as illustrated in \autoref{fig:prompt}.

The standard review prompt in `prismAId` may consist of six main blocks, which are combined with the text of the manuscript being reviewed. Each model call is isolated for individual manuscripts, creating separate interactions to prevent cross-manuscript influence and bias.

The six blocks include: the persona, which defines the model’s role to provide context for its responses; the task, specifying the exact information extraction task; and the expected result, typically a structured JSON object formatted according to the configuration file. The failsafe block offers guidance for the model on how to handle uncertain situations (e.g., leave the response empty if unsure). The definitions block includes detailed explanations for the keys and possible values of the JSON object, minimizing the risk of misinterpretation. Finally, the examples block includes sample manuscripts and their expected outputs, offering concrete instances that help the model understand the desired format and content of the extracted information, thus reducing the risk of misinterpretation further.

# Discussion and Conclusions

`prismAId` has the potential to make a significant impact on the research community, particularly in the field of SLRs. By leveraging advanced AI capabilities, the tool streamlines the review process, offering a faster, more efficient, and scalable alternative to traditional methods. The focus on open science principles enhances transparency and reproducibility, making systematic reviews more accessible and easier to update. This shift is especially important in fields where the volume of literature is growing rapidly, helping researchers keep up with the latest developments while reducing the manual workload.

Despite its strengths, there are potential limitations that need to be addressed. The reliance on generative AI models means that the quality of extracted information depends on the model's understanding and performance, which can vary across different contexts and disciplines. Ensuring that prompts are well-designed and tailored to specific use cases is crucial, as poorly constructed prompts could lead to inaccuracies or misinterpretations. Additionally, while `prismAId` integrates with leading AI models and provides robust features, further testing and feedback are needed to optimize its performance across diverse research areas. Planned future developments include expanded cloud support, integration with new AI models, and enhancements to the user interface to streamline configuration and improve usability. 

An important area for future development in the coming years will be the creation of standardized approaches to assess potential biases, both in the literature being reviewed and within the review configuration itself. By establishing clear guidelines and robust evaluation methods, `prismAId` aims to address the risks of bias that may arise from the selection of manuscripts, the framing of prompts, and the interpretation of AI-generated outputs. These advancements will contribute to enhancing the reliability and credibility of SLRs, further aligning with the principles of open science and evidence-based research.

Anecdotal evidence suggests that `prismAId` outperforms traditional SLR methods in both accuracy and efficiency. Preliminary applications, such as the one described by @boero_ai-enhanced_2024, demonstrate significant time savings, with reviews completed in a fraction of the time required for manual analysis. The cost of running these smaller-scale reviews was minimal and has been continuously and significantly decreasing, reflecting the rapid advancements in the availability and affordability of AI model services.

In summary, `prismAId` represents a major step forward in the field of SLRs, advancing the principles of open science by enabling faster, more transparent, and replicable analyses. Its key contributions lie in the automation of data extraction, the standardization of prompts, and the provision of accessible, multi-platform tools for researchers. Its open-source nature is fundamental to `prismAId`, as it enables thorough code inspection, facilitates the replication of results, and invites the broader research community to actively contribute to its ongoing development and improvement.

# Acknowledgements

`prismAId` was initiated with the generous support of a SIS internal project from NILU. Their support was crucial in starting this research and development effort. Further, acknowledgment is due for the research credits received from the OpenAI Researcher Access Program and the Cohere For AI Research Grant Program, both of which have significantly contributed to the advancement of this work.

# References
