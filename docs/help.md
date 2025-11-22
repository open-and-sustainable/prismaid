---
title: Help
layout: default
---

# Help

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

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
