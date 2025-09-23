# ![logo](https://raw.githubusercontent.com/ricboer0/prismAId/main/figures/prismAId_logo.png) prismAId
# Open Science AI Tools for Systematic, Protocol-Based Literature Reviews

prismAId offers a suite of tools using generative AI models to streamline systematic reviews of scientific literature.

It provides simple-to-use, efficient, and replicable methods for analyzing research papers with no coding skills required.

---

[![GitHub Release](https://img.shields.io/github/v/release/Open-and-Sustainable/prismAId?sort=semver&display_name=tag&style=flat)](https://github.com/Open-and-Sustainable/prismAId/releases)
[![GitHub top language](https://img.shields.io/github/languages/top/Open-and-Sustainable/prismAId?style=flat)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-and-sustainable/prismaid)](https://goreportcard.com/report/github.com/open-and-sustainable/prismaid)
[![Go Reference](https://pkg.go.dev/badge/github.com/open-and-sustainable/prismaid.svg)](https://pkg.go.dev/github.com/open-and-sustainable/prismaid)
[![PyPI - Version](https://img.shields.io/pypi/v/prismaid?logo=pypi)](https://pypi.org/project/prismaid/)
[![R-universe status badge](https://open-and-sustainable.r-universe.dev/badges/prismaid)](https://open-and-sustainable.r-universe.dev/prismaid)

[![GitHub License](https://img.shields.io/github/license/Open-and-Sustainable/prismAId?style=flat)](https://www.gnu.org/licenses/agpl-3.0.en.html#license-text)
[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.11210796.svg)](https://doi.org/10.5281/zenodo.11210796)
[![DOI](https://joss.theoj.org/papers/10.21105/joss.07616/status.svg)](https://doi.org/10.21105/joss.07616)
[![DOI]( https://img.shields.io/badge/user_manual-10.5281/zenodo.15394332-blue)](https://raw.githubusercontent.com/open-and-sustainable/prismaid_manual/main/prismaid_manual.pdf)

---

## Tools Overview
prismAId offers a comprehensive set of tools for systematic literature reviews:

<div style="text-align: center;">
    <img src="https://raw.githubusercontent.com/open-and-sustainable/prismaid/main/figures/tools.png" alt="Tools Overview" style="width: 40%;">
</div>

### Core Tools
1. **Screening** - Filter and tag manuscripts to identify items for exclusion
2. **Download** - Download papers from Zotero collections or from URL lists
3. **Convert** - Convert files (PDF, DOCX, HTML) to plain text for analysis
4. **Review** - Process systematic literature reviews based on TOML configurations

### Access Methods
- **Command Line Interface** - For users who prefer terminal-based workflows
- **Web Initializer** - A browser-based setup tool for configuring reviews
- **Programming Libraries** - API access through multiple languages:
  - Go (native implementation)
  - Python package
  - R package
  - Julia package

---

## Specifications
- **Review protocol**: Supports any literature review protocol with a preference for [Prisma 2020](https://www.prisma-statement.org/prisma-2020), which inspired our project name.
- **Distribution**: Available as:
  - Go [package](https://pkg.go.dev/github.com/open-and-sustainable/prismaid)
  - Python [package](https://pypi.org/project/prismaid/)
  - R [package](https://open-and-sustainable.r-universe.dev/prismaid)
  - Julia [package](https://github.com/JuliaRegistries/General/tree/master/P/PrismAId)
  - 'no-coding' [binaries](https://github.com/open-and-sustainable/prismaid/releases) for Windows, MacOS, and Linux (AMD64/ARM64)
- **Supported LLMs**:
    1. **OpenAI**: GPT-3.5 Turbo, GPT-4 Turbo, GPT-4o, GPT-4o Mini, GPT-4.1, GPT-4.1 Mini, GPT-4.1 Nano, o1, o1 Mini, o3, o3 Mini, and o4 Mini
    2. **GoogleAI**: Gemini 1.0 Pro, Gemini 1.5 Pro, Gemini 1.5 Flash, Gemini 2.0 Flash, and Gemini 2.0 Flash Lite
    3. **Cohere**: Command, Command Light, Command R, Command R+, Command R7B, Command R (August 2024), and Command A
    4. **Anthropic**: Claude 3 Sonnet, Claude 3 Opus, Claude 3 Haiku, Claude 3.5 Haiku, Claude 3.5 Sonnet, Claude 3.7 sonnet, Claude 4.0 Sonnet, and Claude 4.0 Opus
    5. **DeepSeek**: DeepSeek Chat v3, and DeepSeek Reasoner v3
- **Screening capabilities**: Deduplication, language filtering, article type classification, and off-topic detection
- **Output format**: Data in CSV or JSON formats
- **Performance**: Efficiently processes extensive datasets with minimal setup and **no coding** required
- **Programming Language**: Core implementation in Go with bindings for Python, R, and Julia

---

## Documentation
All information on installation, usage, and development is available at [open-and-sustainable.github.io/prismaid](https://open-and-sustainable.github.io/prismaid/) and in the [prismAId User Manual](https://raw.githubusercontent.com/open-and-sustainable/prismaid_manual/main/prismaid_manual.pdf).

---

## Credits
### Authors
Riccardo Boero - ribo@nilu.no

### Acknowledgments
This project was initiated with the generous support of a SIS internal project from [NILU](https://nilu.com). Their support was crucial in starting this research and development effort. Further, acknowledgment is due for the research credits received from the [OpenAI Researcher Access Program](https://grants.openai.com/prog/openai_researcher_access_program/) and the [Cohere For AI Research Grant Program](https://share.hsforms.com/1aF5ZiZDYQqCOd8JSzhUBJQch5vw?ref=txt.cohere.com), both of which have significantly contributed to the advancement of this work.

---

## License
GNU AFFERO GENERAL PUBLIC LICENSE, Version 3

[![license](https://www.gnu.org/graphics/agplv3-155x51.png)](https://www.gnu.org/licenses/agpl-3.0.en.html#license-text)

---

## Contributing
Contributions are welcome! Please follow guidelines at [open-and-sustainable.github.io/prismaid/research-development.html](https://open-and-sustainable.github.io/prismaid/research-development.html#contributing).

---

## Citation
Boero, R. (2025). prismAId - Open Science AI Tools for Systematic, Protocol-Based Literature Reviews. Zenodo. [DOI: 10.5281/zenodo.11210796](https://doi.org/10.5281/zenodo.11210796)
