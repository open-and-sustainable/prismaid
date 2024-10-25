---
title: For Developers
layout: default
---

# For Developers

## Contributing
We value community contributions that help improve prismAId. Whether you're fixing bugs, adding features, or improving documentation, your input is welcome:
- **Branching Strategy**: Please create a new branch for each set of related changes and submit a pull request through GitHub.
- **Code Reviews**: All submissions undergo a thorough review process to maintain code quality and consistency.
- **Community Engagement**: Engage with us through GitHub [issues](https://github.com/open-and-sustainable/prismaid/issues) and [discussions](https://github.com/open-and-sustainable/prismaid/discussions) for feature requests, suggestions, or any queries related to project development.

For detailed guidelines on contributing, please refer to our [`CONTRIBUTING.md`](CONTRIBUTING.md) and [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

## Software Stack and Approach
prismAId is built using the Go programming language, known for its simplicity and efficiency in handling concurrent operations. We prioritize staying up-to-date with the latest stable releases of Go to leverage the newest features and improvements.

### Development Environment Setup
To facilitate the development and testing of prismAId, templates for configuring VSCodium (or Visual Studio Code) are provided. These templates include predefined settings and extensions that enhance the development experience, ensuring consistency across different setups.
- **Accessing Templates**: You can find the configuration templates in the [`cmd` directory](https://github.com/open-and-sustainable/prismaid/tree/main/cmd) of our source repository. 

#### Using the Templates
1. **Clone the Repository**: Start by cloning the prismAId repository to your local machine.
2. **Open with VSCodium/VSCode**: Open the directory within VSCodium or VSCode.
3. **Copy the .json Files**: Copy them in a newly created `vscode`directory on the root of the project.
4. **Remove the .template extension**: Change the file names and follow the instructions within the files.
5. **Ignore the Files in GIT**: Add the files to your local .gitignore to avoid sharing secrets and other private information.

### Architecture
Our architecture is designed to be robust yet simple, ensuring that the tool remains accessible to both technical and non-technical users:
- **Self-Contained Binaries**: prismAId is distributed as self-contained binaries, which means all necessary libraries and dependencies are packaged together. This approach eliminates the need for external installations and simplifies the setup process.
- **Cross-Platform Compatibility**: Compatible with major operating systems such as Windows, MacOS, and Linux, ensuring that prismAId can be used in diverse environments.

### Development Philosophy
- **Open Source**: We embrace an open-source model, encouraging community contributions and transparency in development.
- **Continuous Integration/Continuous Deployment (CI/CD)**: We utilize CI/CD pipelines to maintain high standards of quality and reliability, automatically testing and deploying new versions as they are developed.


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>