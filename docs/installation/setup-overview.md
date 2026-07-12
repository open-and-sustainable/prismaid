---
title: Setup Overview
layout: default
---

# Setup Overview

---

## Supported Systems
prismAId is accessible across multiple platforms, offering flexibility based on user preference and system requirements:

1. **MCP Server**: A main entry point for conversational use — connect an AI agent and let it drive the tools. See the [MCP Server](../mcp-server) page.

2. **Binaries**: Standalone executables for Windows, macOS, and Linux, requiring no coding skills.

3. **Go Package**: Full functionality for Go-based projects.

4. **Python Package** on PyPI: For integration in Python scripts and Jupyter notebooks.

5. **R Package** on R-universe: Compatible with R and RStudio environments.

6. **Julia Package** from the Julia official package registry: For integration in Julia workflows and Jupyter notebooks.

## Toolkit Overview

prismAId offers several specialized tools to support systematic reviews:

1. **Screening Tool**: Filter manuscripts to identify items for exclusion
   - Remove duplicates using exact or fuzzy matching
   - Filter by language detection
   - Classify and filter by article type (research, review, editorial, etc.)

2. **Download Tool**: Acquire papers for your review
   - Download PDFs directly from Zotero collections
   - Download files from URL lists

3. **Convert Tool**: Transform documents into analyzable text
   - Convert PDFs, DOCX, and HTML files to plain text
   - Prepare documents for AI processing

4. **Review Tool**: Process systematic literature reviews based on TOML configurations
   - Configure review criteria, AI model settings, and output formats
   - Extract structured information from scientific papers
   - Generate comprehensive review summaries

5. **RevAIse Documentation Support**: Optionally document workflow stages as [RevAIse](../review/revaise-integration) review records
   - Update the same review record across Zotero download, screening, and review/extraction stages
   - Preserve previous snapshots with automatic backups

## Workflow Overview
1. **AI Model Provider Account and API Key**:
    - Register for an account with [OpenAI](https://www.openai.com/), [GoogleAI](https://aistudio.google.com), [Cohere](https://cohere.com/), or [Anthropic](https://www.anthropic.com/) and obtain an API key.
    - Generate an API key from the provider's dashboard.
2. **Install prismAId**:
    - Follow the installation instructions below based on your preferred system.
3. **Prepare Papers for Review:**
    - Screen the manuscript list using the Screening tool to exclude out-of-scope items
    - Download the retained papers using the Download tool
    - Convert papers to text format using the Convert tool
4. **Define and Run the Review Project:**
    - Set up a configuration file (.toml) for your review project
    - Use the Review tool to process your papers and extract information
5. **Document the Review Record Optionally:**
    - Add `[revaise]` blocks to supported configuration files to update one cumulative review record
    - See the [RevAIse integration guide](../review/revaise-integration)


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
