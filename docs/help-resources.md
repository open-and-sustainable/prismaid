---
title: Help & Resources
layout: default
---

# Help & Resources

## Common Issues
- **Package Crashes**: Most of the time it is because of path problems, like typos or non existent paths for inputs or outputs.
- **Debugging Information**: The level of debugging information and its permanent storage can be easily activated within the project configuration.
- **API Keys**: These may be loaded either through system variables or the project configuration. If multiple keys are provided, the latter will have priority.
- **Partial Results**: Results presenting only the first few entries of a review containing many more manuscripts suggest an incorrect adherence to the Token Per Minute limits set in the project configuration.
- **Software Bugs**: For troubleshooting software issues, submit an [issue on the GitHub repository](https://github.com/open-and-sustainable/prismaid/issues).
- **Feature Requests**: To submit requests for new functionalities, participate in [GitHub Discussions](https://github.com/open-and-sustainable/prismaid/discussions).

## Best Practices
### Learn prismAId
- Carefully read the technical FAQs below to avoid misusing the tool and to access emerging scientific references on issues related to the use of generative AI similar to those you may encounter in prismAId.
- We provide an additional example in the [projects](https://github.com/open-and-sustainable/prismaid/blob/main/projects/test.toml) directory. This includes not only the project configuration but also [input files](https://github.com/open-and-sustainable/prismaid/tree/main/projects/input/test) and [output files](https://github.com/open-and-sustainable/prismaid/tree/main/projects/output/test). The input text is extracted from a study we conducted [doi.org/10.3390/cli10020027](https://doi.org/10.3390/cli10020027).
- Multiple protocols for reporting systematic literature reviews are supported by prismAId [https://doi.org/10.1186/s13643-023-02255-9](https://doi.org/10.1186/s13643-023-02255-9). Users are encouraged to experiment and define their own prismAId methodologies.

### Methodology of prismAId
1. Remove unnecessary sections from the literature to be reviewed.
2. It's better to risk repeating an explanation of the information you are seeking through examples than not defining it clearly enough.
3. If the budget allows, conduct a separate review process for each piece of information you want to extract. This allows for more detailed definitions for each information piece.
4. Try to avoid using open-ended answers and define and explain all possible answers the AI model can provide.
5. First, run a test on a single paper. Once the results are satisfactory, repeat the process with a different batch of papers. If the results are still satisfactory, proceed with the rest of the literature.
6. Focus on primary sources and avoid reviewing reviews unless it is intentional and carefully planned. Do not mix primary and secondary sources in the same review process.
7. Include the project configuration (the .toml file) in the appendix of your paper, it's all about open science.
8. Properly cite prismAId [doi.org/10.5281/zenodo.11210796](https://doi.org/10.5281/zenodo.11210796).

## Technical FAQs

1. **Q: Will I always get the same answer if I set the model temperature to zero?**

**A:** No, setting the model temperature to zero does not guarantee identical answers every time. Generative AI models, including GPT-4, can exhibit some variability even with a temperature setting of zero. While the temperature parameter influences the randomness of the output, setting it to zero aims to make the model more deterministic. However, GPT-4 and similar models are sparse mixture-of-experts models, meaning they may still show some probabilistic behavior at higher levels.<br>
This probabilistic behavior becomes more pronounced when the prompts are near the maximum token limit. In such cases, the content within the model's attention window may change due to space constraints, leading to different outputs. Additionally, there are other mechanisms within the model that can affect its determinism.<br>
Nevertheless, using a lower temperature is a good strategy to minimize probabilistic behavior. During the development phase of your project, repeating prompts can help achieve more consistent and replicable answers, contributing to a robust reviewing process. Robustness of answers could further be tested by modifying the sequence order of prompt building blocks, i.e., the order by which information is presented to the model.<br>
**Further reading:** [https://doi.org/10.48550/arXiv.2308.00951](https://doi.org/10.48550/arXiv.2308.00951)

2. **Q: Does noise (or how much information is hidden in the literature to be reviewed) have an impact?**

**A:** Yes, the presence of noise and the degree to which information is hidden significantly impact the quality of information extraction. The more obscured the information and the higher the noise level in the prompt, the more challenging it becomes to extract accurate and high-quality information. <br>
During the project configuration development phase, thorough testing can help identify the most effective prompt structures and content. This process helps in refining the prompts to minimize noise and enhance clarity, thereby improving the ability to find critical information even when it is well-hidden, akin to finding a "needle in the haystack."<br>
**Further reading:** [https://doi.org/10.48550/arXiv.2404.08865](https://doi.org/10.48550/arXiv.2404.08865)

3. **Q: What happens if the literature to be reviewed says something different from the data used to train the model?**

**A:** This is a challenge that cannot be completely avoided. We do not have full transparency on the exact data used for training the model. If the literature and the training data conflict, information extraction from the literature could be biased, transformed, or augmented by the training data. <br>
While focusing strictly on providing results directly from the prompt can help minimize these risks, they cannot be entirely eliminated. These biases must be carefully considered, especially when reviewing information on topics that lack long-term or widely agreed-upon consensus. These biases are similar to those of human reviewers and are very difficult to control. <br>
However, the prismAId ability to replicate reviews and experiment with different prompts provides an additional tool for checking these biases, offering an advantage over traditional human reviewers.<br>
**Further reading:** [https://doi.org/10.48550/arXiv.2404.08865](https://doi.org/10.48550/arXiv.2404.08865)

4. **Q: Are there reasoning biases I should expect when analyzing literature with generative AI models?**

**A:** Yes, similar to human reasoning biases, AI models trained on human texts can replicate these biases and may lead to false information extraction if the prompts steer them in that direction. This is because the models learn from and mimic the patterns found in the training data, which includes human reasoning biases. <br>
A good strategy to address this in prismAId is to ensure prompts are carefully crafted to be as precise, neutral, and unbiased as possible. prismAId's prompt structure supports the minimization of these problems, but there is still no guarantee against misuse by researchers.<br>
**Further reading:** [https://doi.org/10.1038/s43588-023-00527-x](https://doi.org/10.1038/s43588-023-00527-x)

5. **Q: Is it always better to analyze literature by extracting one piece of information at a time (one piece of information per prismAId project)?**

**A:** Yes, creating a separate prismAId project for each piece of information to be analyzed is a viable and highly effective approach. The main advantage is that it allows you to tailor the prompt structure and content, effectively guiding the AI model to provide accurate answers. Adding multiple information retrieval tasks within a single prismAId project requires writing much longer prompts, which can lead to more complex, potentially noisier, and confused requests for the AI model. <br>
The only drawback is the higher cost incurred from using the model API. Separating two questions into distinct projects approximately doubles the cost of the analysis, as most of the tokens in each project are comprised of the literature text. Therefore, the only constraint to quality is the budget.<br>
**Further reading:** [OpenAI API Prices](https://openai.com/api/pricing/) - [https://doi.org/10.48550/arXiv.2404.08865](https://doi.org/10.48550/arXiv.2404.08865)


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>