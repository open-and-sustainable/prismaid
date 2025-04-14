---
title: Research & Development
layout: default
---

# Research & Development

## Scope
- **Objective**: prismAId leverages Large Language Models (LLMs) for systematic scientific literature reviews, making them accessible and efficient without coding.
- **Speed**: Faster than traditional methods, prismAId provides high-speed software for systematic reviews.
- **Replicability**: Addresses the challenge of consistent, unbiased analysis, countering the subjective nature of human review.
- **Cost**: More economical than custom AI solutions, with review costs typically between $0.0025 and $0.10 per paper.
- **Audience**: Suitable for scientists conducting literature reviews, meta-analyses, project development, and research proposals.

## Mechanism

### LLM Basics
- **How LLMs Work**:
  - Large Language Models (LLMs) are AI trained on extensive text data to understand and generate human-like text.
  - These models handle various language tasks like text completion, summarization, translation, and more.
- **Data Flow and Processing**:
  - Modern LLMs offer subscription-based API access, with prismAId focusing on prompt engineering to extract targeted information.
  - prismAId enables structured, replicable prompt creation for systematic reviews, simplifying rigorous data extraction.

### Data Flow
- prismAId’s workflow embeds protocol-based approaches:
  - **Literature Selection**: Based on defined protocols, ensuring replicability.
  - **Content Classification**: prismAId handles paper classification, parsing selected literature to extract user-defined information.
  - **API Calls & Cost Management**: prismAId sends single-shot prompts for each paper, processes AI-generated JSON files, and provides token-based cost estimates for informed decision-making.

## Contributing

### How to Contribute
We welcome contributions to improve prismAId, whether you're fixing bugs, adding features, or enhancing documentation:
- **Branching Strategy**: Create a new branch for each set of related changes and submit a pull request via GitHub.
- **Code Reviews**: All submissions undergo thorough review to maintain code quality.
- **Community Engagement**: Connect with us through GitHub [issues](https://github.com/open-and-sustainable/prismaid/issues) and [discussions](https://github.com/open-and-sustainable/prismaid/discussions) for feature requests, suggestions, or questions. - Discuss in the Matrix [prismAId Support Room](https://matrix.to/#/#prismAId-support:matrix.org) or follow the [prismAId Announcements Room](https://matrix.to/#/#prismAId-announcements:matrix.org) for the latest updates and release notifications.

### Guidelines
For detailed contribution guidelines, see our [`CONTRIBUTING.md`](CONTRIBUTING.md) and [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

## Software Stack

prismAId is developed in Go, selected for its simplicity and efficiency with concurrent operations. We prioritize the latest stable Go releases to incorporate improvements.

### Technical Foundation
prismAId leverages the [`alembica`](https://github.com/open-and-sustainable/alembica) pure Go package to manage interactions with Large Language Models. This foundation allows us to concentrate on developing robust protocol-based information extraction tools while `alembica` handles the standardized communication with various LLMs through consistent JSON data schemas, ensuring reliability and interoperability across different AI services.

### Architecture
- **Go Module**: Core logic and API access are implemented in Go.
- **Python Package**: Python wrapper around a C shared library compiled from the Go code.
- **R Package**: Contains a C shared library with an intermediate C wrapper, enabling R interaction with the shared library.
- **Self-Contained Binaries**: Simplifies setup by packaging all dependencies within the binaries.
- **Cross-Platform Compatibility**: Fully operational across Windows, macOS, and Linux.

### Development Philosophy
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

4. **Quality and Accuracy**
   - Explicit prompts define information clearly, ensuring consistent, reliable reviews.

5. **Ethics and Bias Reduction**
   - Transparent design minimizes biases, with community oversight supporting ethical standards.

6. **Scientific Innovation**
   - Standardized, reusable methods facilitate innovation, cumulative knowledge, and rapid knowledge dissemination.



<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
