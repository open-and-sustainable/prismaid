---
title: Help & Development
layout: default
---

# Help & Development

If you need assistance, you can:
- **Open an Issue** on our [GitHub repository](https://github.com/open-and-sustainable/prismaid/issues).
- **Discuss in the Matrix Support Room**: [prismAId Support Room](https://matrix.to/#/#prismAId-support:matrix.org) ![Matrix](https://img.shields.io/matrix/prismAId-support%3Amatrix.org?server_fqdn=matrix.org&logo=appveyor).
- **Stay Updated with New Releases**: Follow the [prismAId Announcements Room](https://matrix.to/#/#prismAId-announcements:matrix.org) for the latest updates and release notifications.

## Common Issues
- **Package Crashes**: Most of the time it is because of path problems, like typos or non existent paths for inputs or outputs.
- **Debugging Information**: The level of debugging information and its permanent storage can be easily activated within the project configuration.
- **API Keys**: These may be loaded either through system variables or the project configuration. If multiple keys are provided, the latter will have priority.
- **Partial Results**: Results presenting only the first few entries of a review containing many more manuscripts suggest an incorrect adherence to the Token Per Minute limits set in the project configuration.
- **Software Bugs**: For troubleshooting software issues, submit an [issue on the GitHub repository](https://github.com/open-and-sustainable/prismaid/issues).
- **Feature Requests**: To submit requests for new functionalities, participate in [GitHub Discussions](https://github.com/open-and-sustainable/prismaid/discussions).


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
