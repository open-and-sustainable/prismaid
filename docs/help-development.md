---
title: Help & Development
layout: default
---

# Help & Development

 ---

<details>
<summary><strong>Page Contents</strong></summary>
<ul>
  <li><a href="#getting-help"><strong>Getting Help</strong></a>: where to find assistance</li>
  <li><a href="#common-issues"><strong>Common Issues</strong></a>: solutions to frequent problems</li>
  <li><a href="#contributing"><strong>Contributing</strong></a>: how to improve prismAId</li>
  <li><a href="#software-stack"><strong>Software Stack</strong></a>: technical architecture</li>
  <li><a href="#open-science-support"><strong>Open Science Support</strong></a>: how prismAId supports open science principles</li>
</ul>
</details>

 ---

## Getting Help

If you need assistance with any prismAId tool, you can:
- **Open an Issue** on our [GitHub repository](https://github.com/open-and-sustainable/prismaid/issues).
- **Discuss in the Matrix Support Room**: [prismAId Support Room](https://matrix.to/#/#prismAId-support:matrix.org) ![Matrix](https://img.shields.io/matrix/prismAId-support%3Amatrix.org?server_fqdn=matrix.org&logo=appveyor).
- **Stay Updated with New Releases**: Follow the [prismAId Announcements Room](https://matrix.to/#/#prismAId-announcements:matrix.org) for the latest updates and release notifications.

## Common Issues

### General Issues
- **Path Problems**: Most crashes occur due to incorrect paths in configurations, such as typos or non-existent directories.
- **API Keys**: These may be loaded either through system variables or directly in configurations. When both are provided, the configuration values take priority.
- **Software Bugs**: For troubleshooting software issues, submit an [issue on the GitHub repository](https://github.com/open-and-sustainable/prismaid/issues).
- **Feature Requests**: To submit requests for new functionalities, participate in [GitHub Discussions](https://github.com/open-and-sustainable/prismaid/discussions).

### Screening Tool Issues
- **False Positives in Deduplication**: Adjust the similarity threshold (increase from 0.85 to 0.95 for stricter matching).
- **Language Detection Errors**: Enable AI-based detection for mixed-language documents or check text encoding.
- **Article Type Misclassification**: Review classification rules or use AI-based classification for ambiguous cases.

### Download Tool Issues
- **Zotero Authentication Errors**: Verify your user ID and API key, ensuring the API key has appropriate permissions.
- **Collection Not Found**: Check that the collection/group path uses the correct format and exists in your Zotero library.
- **Download Failures**: If some PDFs fail to download, check if they are actually available/accessible in your Zotero library.

### Convert Tool Issues
- **Conversion Quality**: PDF conversion may result in imperfect text extraction due to limitations of the PDF format. Always check converted files.
- **Unsupported Formats**: Ensure you're using supported file formats (PDF, DOCX, HTML).
- **Character Encoding**: Some converted texts may display encoding issues with special characters.

### Review Tool Issues
- **Debugging Information**: Control the level of debugging information via the `log_level` parameter in the project configuration.
- **Partial Results**: If only the first few entries of a review are processed, check the Token Per Minute (TPM) limits in your configuration.
- **Response Format Errors**: Ensure your prompt and review sections are well-structured to guide the AI to produce correctly formatted outputs.
- **Token Limits**: Very large documents may exceed model token limits; consider splitting or summarizing them.

## Contributing

### How to Contribute
We welcome contributions to improve any aspect of the prismAId toolkit, whether you're fixing bugs, adding features, or enhancing documentation:
- **Branching Strategy**: Create a new branch for each set of related changes and submit a pull request via GitHub.
- **Code Reviews**: All submissions undergo thorough review to maintain code quality.
- **Community Engagement**: Connect with us through GitHub [issues](https://github.com/open-and-sustainable/prismaid/issues) and [discussions](https://github.com/open-and-sustainable/prismaid/discussions) for feature requests, suggestions, or questions. - Discuss in the Matrix [prismAId Support Room](https://matrix.to/#/#prismAId-support:matrix.org) or follow the [prismAId Announcements Room](https://matrix.to/#/#prismAId-announcements:matrix.org) for the latest updates and release notifications.

### Guidelines
For detailed contribution guidelines, see our [`CONTRIBUTING.md`](CONTRIBUTING.md) and [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

## Software Stack

prismAId is developed in Go, selected for its simplicity and efficiency with concurrent operations. We prioritize the latest stable Go releases to incorporate improvements.

### Technical Foundation
prismAId leverages the [`alembica`](https://github.com/open-and-sustainable/alembica) pure Go package to manage interactions with Large Language Models. This foundation allows us to concentrate on developing robust protocol-based information extraction tools while `alembica` handles the standardized communication with various LLMs through consistent JSON data schemas, ensuring reliability and interoperability across different AI services.

### Toolkit Architecture
The prismAId toolkit is structured as a set of modular tools (Screening, Download, Convert, Review) that can be used together or independently:

- **Go Module**: Core logic and API access for all tools are implemented in Go.
- **Cross-Language Support**: Each tool is accessible through:
  - **Python Package**: Python wrapper around a C shared library compiled from the Go code.
  - **R Package**: Contains a C shared library with an intermediate C wrapper, enabling R interaction.
  - **Julia Package**: Contains a C shared library with Julia bindings for direct integration.
- **Self-Contained Binaries**: Simplifies setup by packaging all dependencies within the binaries.
- **Cross-Platform Compatibility**: Fully operational across Windows, macOS, and Linux.

### Development Philosophy
- **Modularity**: Tools that work together but can be used independently following the workflow: Search → Screen → Download → Convert → Review.
- **Open Source**: We value community contributions and transparency.
- **CI/CD Pipelines**: Automated testing and deployment maintain quality and reliability.

## Open Science Support
prismAId actively supports Open Science principles through:

1. **Transparency and Reproducibility**
   - prismAId enhances transparency, making analyses understandable and reproducible, with consistent results across systematic reviews.
   - Detailed logs and records improve reproducibility.

2. **Accessibility and Collaboration**
   - An open-source, openly licensed tool fostering collaboration and participation.
   - Long-term accessibility through [Zenodo](https://zenodo.org/doi/10.5281/zenodo.11210796).

3. **Efficiency and Scalability**
   - Efficient data handling enables timely, comprehensive reviews.
   - Modular tools allow flexible workflows adapted to different research needs.

4. **Quality and Accuracy**
   - Explicit prompts define information clearly, ensuring consistent, reliable reviews.
   - Separate tools for each workflow step improve focus and quality.

5. **Ethics and Bias Reduction**
   - Transparent design minimizes biases, with community oversight supporting ethical standards.

6. **Scientific Innovation**
   - Standardized, reusable methods facilitate innovation, cumulative knowledge, and rapid knowledge dissemination.


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
