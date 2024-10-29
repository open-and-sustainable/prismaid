---
title: Home
layout: default
---

# Open Science AI Tools for Systematic, Protocol-Based Literature Reviews.

## Purpose
prismAId uses generative AI models to extract data from scientific literature. It offers simple-to-use, efficient, and replicable methods for analyzing literature when conducting systematic reviews. No coding skills are required to use prismAId.

## Specifications
- **Review protocol**: Designed to support any literature review protocol with a preference for [Prisma 2020](https://www.prisma-statement.org/prisma-2020), which inspired our project name.
- **Distribution**: Go [package](https://pkg.go.dev/github.com/open-and-sustainable/prismaid), Python [package](https://pypi.org/project/prismaid/), R [package](https://open-and-sustainable.r-universe.dev/prismaid), and 'no-coding' [binaries](https://github.com/open-and-sustainable/prismaid/releases) compatible with Windows, MacOS, and Linux operating systems on AMD64 and ARM64 platforms.
- **Supported LLMs**: 
    1. **OpenAI**: GPT-3.5 Turbo, GPT-4 Turbo, GPT-4o, and GPT-4o Mini.
    2. **GoogleAI**: Gemini 1.0 Pro, Gemini 1.5 Pro, and Gemini 1.5 Flash.
    3. **Cohere**: Command, Command Light, Command R, and Command R+.
    4. **Anthropic**: Claude 3 Sonnet, Claude 3 Opus, Claude 3 Haiku, Claude 3.5 Sonnet
- **Output format**: Outputs data in CSV or JSON formats.
- **Performance**: Designed to process extensive datasets efficiently with minimal user setup and **no coding** required.
- **Programming Language**: Developed in Go.

## Table of Contents
Visit this [documentation website](https://open-and-sustainable.github.io/prismaid/) to find: 
1. Instructions to install prismAId and start using it: [Getting Started](getting-started).
2. A walkthrough of the process of setting up your systematic literature review project: [Project Setup](project-setup).
3. A detailed guide on designing robust prompts to exploit generative AI models for information extraction from the literature: [Prompt Design](prompt-design).
4. A presentation of the most advanced features provided by the tool: [Advanced Features](advanced-features).
5. Troubleshooting directions and frequently asked questions on prismAId and generative AI features and results: [FAQs and Troubleshooting](faqs).
6. A description of the software implementation of prismAId along with guidelines on how to contribute to its future development: [For Developers](for-developers).


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

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>