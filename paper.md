---
title: 'An Introduction to prismAId: Open-Source and Open Science AI for Advancing Information Extraction in Systematic Reviews'
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

Systematic literature reviews (SLRs) play a crucial role in synthesizing research findings across various disciplines. However, traditional approaches are typically labor-intensive, time-consuming, and prone to subjectivity, making it challenging to manage large datasets while ensuring consistency and reproducibility. `prismAId` addresses these challenges by automating data extraction and analysis through advanced AI models, significantly reducing the time and effort required. Its compatibility with multiple programming languages and operating systems ensures broad accessibility, while integration with established review protocols like PRISMA 2020 guarantees methodological rigor. `prismAId` provides a user-friendly interface that requires no coding skills, enabling researchers from diverse backgrounds to perform comprehensive literature analyses efficiently.

# Features and Capabilities
`prismAId` is compatible with any systematic review protocol [@schiavo_prospero_2019], including the widely adopted PRISMA 2020 guidelines [@page_prisma_2021]. It focuses specifically on the analysis phase, which is a key concluding component of all systematic review protocols, alongside the reporting stage.

![Traditional approach to literature analysis vs. `prismAId`-supported review workflow.\label{fig:workflow}](workflow.png){width="50%"}

As illustrated in \autoref{fig:workflow}, `prismAId` plays a crucial role in the data extraction activities. The tool leverages generative AI models to analyze scientific documents, extracting information in a structured and quantifiable manner. Users only need to configure a review project by specifying the necessary parameters and providing the documents to be analyzed. `prismAId` then automates the entire extraction process, outputting a comprehensive database of results.

The approach taken by `prismAId` enhances the accuracy of information extraction by leveraging AI for natural language processing tasks that are traditionally performed by humans. By replacing manual efforts with computational methods, `prismAId` reduces subjectivity and the issues that arise from vague or incomplete definitions of concepts. Additionally, AI models maintain consistent performance without the effects of distraction or fatigue, further improving the reliability of the results.

Beyond accuracy, the computational nature of `prismAId` enables extremely fast SLRs and ensures full replicability of the analysis. The modular and project-based design allows for easy integration, expansion, and cumulative value, supporting the ongoing buildup of scientific knowledge through consistent and extendable SLRs.

Building on the tool's aim, `prismAId` includes the capability to preprocess and share manuscripts directly with AI models. To achieve this, `prismAId` provides features for converting files from PDF, DOCX, and HTML formats into plain text (TXT). This conversion is essential for effective text tokenization, which involves breaking down the text into smaller units (tokens) such as words or phrases. Tokenization is a fundamental step in natural language processing, enabling the AI models to parse and understand the document content.

Additionally, `prismAId` integrates with Zotero, a popular open source reference management software, allowing users to access shared collections and groups of scientific literature seamlessly. This integration simplifies document retrieval and enhances collaboration. The connection to Zotero is automatic once the API access is configured, as detailed in the `prismAId` documentation. This feature streamlines the workflow, making it easier for users to manage and analyze large sets of scientific manuscripts.

`prismAId` integrates with various large language models (LLMs) from the four main AI providers: OpenAI, Google Cloud AI, Anthropic, and Cohere. This includes access to state-of-the-art models such as many variants of GPT, Gemini, Claude, and Command. The integration is handled through native APIs, allowing for seamless access to advanced generative AI capabilities. Depending on project needs and available opportunities, support for cloud-based deployments of these models will be expanded. The main results dataset generated by `prismAId` can be exported in two formats: a tabular CSV and a nested JSON structure. These options ensure compatibility with a wide range of analysis tools and software.

`prismAId` includes advanced functions designed to support debugging and validation throughout the review process. One key feature is input duplication, which is particularly useful when testing on individual or small sets of manuscripts. This feature facilitates sensitivity analysis, allowing users to experiment with prompt design and other parameters, and evaluate their impact on the results. Another useful capability is the option to generate summaries of manuscripts, enabling manual, ad hoc verification of the extracted outcomes against the main messages of the original papers. Additionally, `prismAId` offers the option to request and store Chain of Thought (COT) justifications. Here, the AI models are prompted to explain their reasoning process and identify the specific parts of the manuscript that support their conclusions, providing greater transparency and accountability.

Moreover, `prismAId` supports ensemble reviews, a powerful feature that allows the same review to be run across different models, both within a single provider and across multiple providers. This approach enables the calculation of uncertainty estimates for the extracted answers, helping to assess the robustness of the results and providing a more comprehensive analysis.

# User Experience
The primary access to `prismAId` is through standalone binaries, compiled for Windows, macOS, and Linux, with support for both AMD64 (x86-64) and ARM64 architectures. These self-contained binaries allow users to run and manage review projects, as well as initialize project configuration files directly from the terminal. They do not require the installation of any additional software beyond terminal access, making them simple and convenient to use. The binaries can be easily downloaded from the release page on the project GitHub repository.

For users with more technical expertise, there are additional programmatic options for accessing `prismAId`. One option is to integrate the tool directly by using the Go module available on the GitHub repo. Alternatively, `prismAId` can be installed as a Python package via PyPI, as a Julia package from the General package registry, or as an R package through R-universe. These options offer more flexibility and are suitable for integrating `prismAId` into custom workflows and advanced scripting environments.

![`prismAId` use case diagram.\label{fig:use-case}](use-case.png){width="60%"}

The tool requires a single input: the project configuration file, which defines all the necessary parameters for conducting a systematic review. This file specifies key details, including the location of the manuscripts to be reviewed and the desired file name and directory for the results. The main steps for creating a project configuration file are illustrated in \autoref{fig:use-case}. Users need to define the input and output paths, select the AI models to be used, configure any advanced features, and specify both the prompt design and the information model for data extraction. This structured approach ensures that `prismAId` can efficiently manage the review process based on clear user-defined parameters.

The tool includes a template.toml file, which outlines and explains all the configuration options that can be used in a project configuration file. To assist users in creating a complete configuration, `prismAId` offers two interactive initializers. The first initializer is terminal-based and can be launched by running the binaries with the `-init` flag. This option provides a step-by-step guide directly in the terminal, helping users generate a draft configuration file.

The second initializer is web-based, accessible via the `prismAId` documentation website [@prismaid-doc]. This version offers a dynamic interface where users can specify all configuration components interactively. It also provides a real-time preview of the configuration file in the web browser, with the option to download the finalized file for further editing or direct use with `prismAId`.

![Prompt blocks in `prismAId`.\label{fig:prompt}](prompt.png){width="60%"}

A key component in ensuring the replicability of SLRs conducted with `prismAId` is the project configuration file itself. When this file is published alongside the review report, it enables others to replicate the original analysis, extend the review to cover new bodies of knowledge, or update the existing review with additional literature. Beyond replicability, `prismAId` also emphasizes quality support, especially through the standardization of prompt design, which guides users in crafting robust and accurate prompts, as illustrated in \autoref{fig:prompt}.

The standard review prompt in `prismAId` may consist of six main blocks, which are combined with the text of the manuscript being reviewed. Each model call is isolated for individual manuscripts, creating separate interactions to prevent cross-manuscript influence and bias.

The six blocks include: the persona, which defines the model’s role to provide context for its responses; the task, specifying the exact information extraction task; and the expected result, typically a structured JSON object formatted according to the configuration file. The failsafe block offers guidance for the model on how to handle uncertain situations (e.g., leave the response empty if unsure). The definitions block includes detailed explanations for the keys and possible values of the JSON object, minimizing the risk of misinterpretation. Finally, the examples block includes sample manuscripts and their expected outputs, offering concrete instances that help the model understand the desired format and content of the extracted information, thus reducing the risk of misinterpretation further.

# Conclusions

`prismAId`, for which @boero_extended_2024 provides details on the software architecture and a comparison with existing AI approaches to SLRs, has the potential to make a significant impact on the research community. By leveraging advanced AI capabilities, the tool offers a faster, more efficient, and scalable alternative to traditional methods to SLRs. The focus on open science principles enhances transparency and reproducibility, making systematic reviews more accessible and easier to update. This shift is especially important in fields where the volume of literature is growing rapidly, helping researchers keep up with the latest developments while reducing the manual workload.

Anecdotal evidence suggests that `prismAId` outperforms traditional SLR methods in both accuracy and efficiency. Preliminary applications, such as the one described by @boero_ai-enhanced_2024, demonstrate significant time savings, with reviews completed in a fraction of the time required for manual analysis. The cost of running these smaller-scale reviews was minimal and has been continuously and significantly decreasing, reflecting the rapid advancements in the availability and affordability of AI model services.

# Acknowledgements

`prismAId` was initiated with the generous support of a SIS internal project from NILU. Their support was crucial in starting this research and development effort. Further, acknowledgment is due for the research credits received from the OpenAI Researcher Access Program and the Cohere For AI Research Grant Program, both of which have significantly contributed to the advancement of this work.

# References
