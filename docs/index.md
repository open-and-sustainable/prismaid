---
title: Home
layout: default
---

# prismAId

### Open Science AI Tools for Systematic, Protocol-Based Literature Reviews

**Run systematic literature reviews with generative AI: transparent, reproducible, and verifiable against reporting standards — with no coding required.**

prismAId screens, acquires, converts, and extracts data from scientific literature, and lets you check the result against reporting protocols such as PRISMA 2020. It is **Open Science first**: every step is built to be shared, reproduced, and machine-checked against public standards.

[Get started](installation/setup-overview) · [Use it from an AI agent](mcp-server) · [Source on GitHub](https://github.com/open-and-sustainable/prismaid)

[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.11210796.svg)](https://doi.org/10.5281/zenodo.11210796) · Free and open source under AGPL-3.0

## Why prismAId

prismAId is built for Open Science, and every feature serves that goal:

- **Open by design** — free and open source under AGPL-3.0, archived on [Zenodo](https://doi.org/10.5281/zenodo.11210796), and usable with no coding skills, so anyone can run and reproduce a review.
- **Reproducible** — each review is defined by shareable configuration files and cumulative [RevAIse](review/revaise-integration) records, so any researcher can rerun or continue an analysis, even across institutions.
- **Verifiable against standards** — conformance to reporting protocols such as PRISMA 2020 is decided by machine-checked [SHACL shapes](conformance), not asserted by the model. It is a reproducible claim, not a promise — and you can see a protocol's [full requirement checklist](guidance) before you start.
- **Accessible everywhere** — drive the whole toolkit from an [AI agent](mcp-server), the command line, a web form, or Go/Python/R/Julia.
- **Current** — supports the latest models from OpenAI, Google, Anthropic, Cohere, DeepSeek and others, plus cloud and self-hosted endpoints.

## Quickstart

**Fastest — with an AI agent.** Connect an assistant to the [MCP server](mcp-server) and describe your review; it generates and validates the configuration, runs the tools, and checks conformance for you.

**Or install and run directly:**

```bash
pip install prismaid   # also available for Go, R, and Julia, or as a no-coding binary
```

Create a configuration with the [web configurator](review/review-configurator) or the `-init` command, then run your review. Full instructions: [Installation & Setup](installation/setup-overview).

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
5. **RevAIse documentation support** - Optionally document review stages as [RevAIse](https://revaise-model.readthedocs.io/stable/) review records

### Access Methods

prismAId can be used two complementary ways:

- **Through an AI agent (MCP server)** — a main entry point for anyone exploring prismAId. Connect an AI assistant to the [prismAId MCP server](mcp-server) and work in conversation: it helps you generate and validate configurations, run the tools, and check and plan for protocol conformance, exposing every capability below through one interface.
- **Directly, on your platform of choice** — granular, multi-platform access to the same tools:
  - **Command Line Interface** — for terminal-based workflows
  - **Web Initializer** — a browser-based setup tool for configuring reviews
  - **Programming Libraries** — Go (native implementation), Python, R, and Julia packages

## Workflow
Our tools support a comprehensive systematic review workflow following the standard sequence: Search → Screen → Download → Convert → Review. Optional RevAIse support can document Zotero download, screening, and review/extraction stages in one cumulative review record.

<div style="text-align: left;">
    <img src="https://raw.githubusercontent.com/open-and-sustainable/prismaid/main/figures/prismAId_workflow.png" alt="Workflow Diagram" style="width: 600px;">
</div>

## Documentation

**Get started**
- [Installation & Setup](installation/setup-overview) — install prismAId and configure it for your environment
- [Recipes](recipes) — short, task-oriented guides for common workflows
- [MCP Server](mcp-server) — drive the whole toolkit from an AI agent

**Tools**
- [Screening](tools/screening-tool) — filter manuscripts: [deduplication](filters/deduplication), [language](filters/language), [article type](filters/article-type), [topic relevance](filters/topic-relevance)
- [Download](tools/download-tool) — acquire papers from Zotero collections or URL lists
- [Convert](tools/convert-tool) — transform PDF, DOCX, and HTML into plain text
- [Review](tools/review-tool) — extract structured information from the literature

**Reviews & records**
- [Review Workflow](review/review-workflow) — methodology and best practices
- [Review Configurator](review/review-configurator) — build a configuration in the browser
- [RevAIse Integration](review/revaise-integration) — maintain a cumulative review record across stages

**Protocol conformance**
- [Protocol Conformance](conformance) — check a review record against a protocol such as PRISMA 2020
- [Protocol Guidance](guidance) — a protocol's full requirement checklist, up front

**Help & contributing**
- [Help](support/help) — troubleshooting and FAQ
- [Development](support/development) — how prismAId is built and how to contribute

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
