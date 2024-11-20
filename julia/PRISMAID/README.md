# ![PRISMAID Logo](https://raw.githubusercontent.com/ricboer0/prismAId/main/figures/prismAId_logo.png) PRISMAID

PRISMAID is a Julia package designed to facilitate access to the [prismAId](https://github.com/open-and-sustainable/prismaid) tools directly from Julia code and workflows.

## Installation
To install PRISMAID using Julia's package manager, run the following commands in your Julia REPL. This will add the PRISMAID package directly from its subdirectory in the GitHub repository:
```julia
using Pkg
Pkg.add(url="https://github.com/open-and-sustainable/prismaid.git", subdir="julia/PRISMAID")
```

## Usage

To utilize PRISMAID in your Julia environment, you need to load the package and execute the `run_review` function, which requires a TOML-formatted review project configuration.

### Quick Start Example

1. Start by loading the PRISMAID package:
   ```julia
   using PRISMAID   
   ```
   
2. Prepare your review project configuration in TOML format. You can use the template provided in the `proj_test.toml` file located in the `projects` folder of our [GitHub repository](https://github.com/open-and-sustainable/prismaid/tree/main/projects). Here’s a simplified example of what the TOML content might look like:
```julia
toml_test = """
       [project]
       name = "Test of prismAId"
       ...
       """
```
3. Run the review process by passing the TOML configuration string to the `run_review` function:
   ```julia
   PRISMAID.run_review(toml_test)
   ```

### Expected Output
When you run the review project, the following output will be displayed in the terminal:
```bash
Processing file 1/1 lit_test with model gpt-4o-mini
The total cost (USD - $) to run this review is at least: 0.00107895
This value is an estimate of the total cost of input tokens only.
Eventual requests for CoT justifications and summaries increase the cost and are not included here.
Do you want to continue? (y/n):
```
At this prompt, you can decide whether to continue processing the review project. If you proceed, the results of the review process will be saved in the output folder specified in your project configuration.

**ATTENTION**: Interaction with prismAId functionalities is mediated through a C shared library, which can make debugging challenging. It is recommended to set the `log_level` to "high" in your project configuration to ensure comprehensive logging of any issues encountered during the review process, with logs stored in the specified output directory.

## Documentation

Comprehensive documentation for PRISMAID, including detailed descriptions of its functionalities, installation guide, usage examples, and configuration settings, is available online. You can access the complete documentation by visiting the following URL:

[prismAId Documentation](https://open-and-sustainable.github.io/prismaid)

## License
PRISMAID is made available under the GNU Affero General Public License v3 (AGPL v3). For detailed terms and conditions, please see the [LICENSE](https://open-and-sustainable.github.io/prismaid/LICENSE) file located at the root of the repository.

