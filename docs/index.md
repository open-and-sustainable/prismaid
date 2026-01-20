---
title: Home
layout: default
---

# Open Science AI Tools for Systematic, Protocol-Based Literature Reviews

## Purpose and Benefits
prismAId changes the way researchers conduct systematic reviews using generative AI. Designed for scholars and professionals alike, it provides a comprehensive toolkit that simplifies the screening, extraction, and analysis of data from scientific literature without requiring coding skills. Whether you're exploring established fields or pioneering new research areas, prismAId ensures that your review processes are efficient, accurate, and reproducible.

### Key Advantages

- **Accessibility**: Easy-to-use interfaces ensure that anyone can leverage advanced AI tools for literature reviews, whether working from the command line or a web browser.
- **Flexibility**: Supports a wide range of literature review protocols, including the Prisma 2020, so teams can align the toolkit with existing workflows.
- **Replicability**: Enables seamless saving and sharing of review inputs, ensuring that any researcher can reproduce or continue the analysis effortlessly, even across institutions.
- **Efficiency**: Optimized for handling large datasets with minimal setup, reducing the time from research to results by automating repetitive screening steps.
- **Innovation**: Continuously updated to incorporate the latest AI advancements, keeping your research at the cutting edge with modern language models.
- **Multi-platform**: Available through multiple programming languages (Go, Python, R, Julia) and as standalone binaries, enabling integration into diverse analytical stacks.

## The prismAId Toolkit

prismAId offers a suite of tools to support every stage of your systematic review:

<div style="text-align: left;">
    <img src="https://raw.githubusercontent.com/open-and-sustainable/prismaid/main/figures/tools.png" alt="Tools Overview" style="width: 600px;">
</div>

### Core Tools
1. **Screening** - Filter and tag manuscripts to identify items for exclusion
2. **Download** - Acquire papers from Zotero collections or from URL lists
3. **Convert** - Transform files (PDF, DOCX, HTML) to plain text for analysis
4. **Review** - Process systematic literature reviews based on TOML configurations

### Access Methods
- **Command Line Interface** - For users who prefer terminal-based workflows
- **Web Initializer** - A browser-based setup tool for configuring reviews
- **Programming Libraries** - API access through multiple languages:
  - Go (native implementation)
  - Python package
  - R package
  - Julia package

## Workflow
Our tools support a comprehensive systematic review workflow following the standard sequence: Search → Screen → Download → Convert → Review

<div style="text-align: left;">
    <img src="https://raw.githubusercontent.com/ricboer0/prismaid/main/figures/prismAId_workflow.png" alt="Workflow Diagram" style="width: 600px;">
</div>

## Table of Contents
Explore this website for comprehensive guidance on using the prismAId toolkit:
1. [Installation & Setup](installation/setup-overview): Learn how to install prismAId tools and configure them for different environments.
2. [Screening Tool](tools/screening-tool): Filter manuscripts with multiple screening filters:
   - [Deduplication](filters/deduplication) - Identify and remove duplicate manuscripts
   - [Language Detection](filters/language) - Filter by manuscript language
   - [Article Type](filters/article-type) - Classify publication types
   - [Topic Relevance](filters/topic-relevance) - Score relevance to research topics
3. [Download Tool](tools/download-tool): Discover how to efficiently acquire papers from Zotero collections or URL lists.
4. [Convert Tool](tools/convert-tool): Learn to transform documents from various formats into plain text for analysis.
5. [Review Tool](tools/review-tool): Master the core systematic review functionality for extracting structured information.
6. [Review Support](review/review-workflow): Learn about methodologies and best practices for systematic reviews with prismAId.
7. [Review Configurator](review/review-configurator): Quickly set up your review project with the web initializer tool.
8. [Help & Development](support/help): Find troubleshooting tips and answers to frequently asked questions about prismAId features and results and how you can contribute to its advancement.

## New Releases and Updates
Follow the Matrix [prismAId Announcements Room](https://matrix.to/#/#prismAId-announcements:matrix.org) for the latest updates and release notifications.

## Credits
### Authors
Riccardo Boero - ribo@nilu.no

### Acknowledgments
This project was initiated with the generous support of a SIS internal project from [NILU](https://nilu.com). Their support was crucial in starting this research and development effort. Further, acknowledgment is due for the research credits received from the [OpenAI Researcher Access Program](https://grants.openai.com/prog/openai_researcher_access_program/) and the [Cohere For AI Research Grant Program](https://share.hsforms.com/1aF5ZiZDYQqCOd8JSzhUBJQch5vw?ref=txt.cohere.com), both of which have significantly contributed to the advancement of this work.

## License
GNU AFFERO GENERAL PUBLIC LICENSE, Version 3

![license](https://www.gnu.org/graphics/agplv3-155x51.png)

## Citation
Boero, R. (2024). prismAId - Open Science AI Tools for Systematic, Protocol-Based Literature Reviews. Zenodo. https://doi.org/10.5281/zenodo.11210796

[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.11210796.svg)](https://doi.org/10.5281/zenodo.11210796)

```bibtex
@software{boero2024prismaid,
  author  = {Boero, Riccardo},
  title   = {prismAId - Open Science AI Tools for Systematic, Protocol-Based Literature Reviews},
  year    = {2024},
  doi     = {10.5281/zenodo.11210796},
  url     = {https://doi.org/10.5281/zenodo.11210796}
}
```

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
