document.addEventListener("DOMContentLoaded", function () {
    // Setup event listener for the Generate Configuration button if needed
    var generateButton = document.getElementById("generateConfigButton");
    if (generateButton) {
        generateButton.addEventListener("click", generateConfig);
    }

    // Setup event listener for the Download Configuration button
    var downloadButton = document.getElementById("downloadButton");
    if (downloadButton) {
        downloadButton.addEventListener("click", downloadConfiguration);
    }
});

function generateConfig() {
    // Gather data from form fields
    var data = {
        project: {
            name: document.getElementById("name").value,
            author: document.getElementById("author").value,
            version: document.getElementById("version").value,
        },
        configuration: {
            input_directory: document.getElementById("input_directory").value,
            results_file_name:
                document.getElementById("results_file_name").value,
            output_format: document.getElementById("output_format").value,
            log_level: document.getElementById("log_level").value,
            duplication: document.getElementById("duplication").value,
            cot_justification:
                document.getElementById("cot_justification").value,
            summary: document.getElementById("summary").value,
        },
        llm_providers: collectProviderData(),
        prompt: {
            persona: document.getElementById("persona").value,
            task: document.getElementById("task").value,
            expected_result: document.getElementById("expected_result").value,
            definitions: document.getElementById("definitions").value,
            example: document.getElementById("example").value,
            failsafe: document.getElementById("failsafe").value,
        },
        review_items: collectReviewData(),
    };

    // Generate TOML string from data
    var toml = generateTOMLString(data);
    document.getElementById("configOutput").value = toml;
}

function collectProviderData() {
    const providers = document.querySelectorAll(".llm-provider");
    const data = Array.from(providers).map((provider) => ({
        provider: provider.querySelector(".provider-select").value,
        api_key: provider.querySelector(".api-key-input").value,
        model: provider.querySelector(".model-input").value,
        temperature: provider.querySelector(".temperature-input").value,
        tpm_limit: provider.querySelector(".tpm-limit-input").value,
        rpm_limit: provider.querySelector(".rpm-limit-input").value,
        base_url: provider.querySelector(".base-url-input")?.value || "",
        endpoint_type:
            provider.querySelector(".endpoint-type-input")?.value || "",
        region: provider.querySelector(".region-input")?.value || "",
        project_id: provider.querySelector(".project-id-input")?.value || "",
        location: provider.querySelector(".location-input")?.value || "",
        api_version: provider.querySelector(".api-version-input")?.value || "",
    }));
    return data;
}

function collectReviewData() {
    const reviews = document.querySelectorAll(".review-item");
    const data = Array.from(reviews).map((review) => {
        const key = review.querySelector(".review-key").value;
        const valuesInput = review.querySelector(".review-values").value;

        // Check if the values input is empty
        const values = valuesInput
            ? valuesInput.split(",").map((v) => v.trim())
            : [];

        return { key, values };
    });
    return data;
}

function generateTOMLString(data) {
    // Build TOML string from the structured data
    var toml = ["[project]"];
    Object.keys(data.project).forEach(function (key) {
        toml.push(`${key} = "${data.project[key]}"`);
    });

    toml.push("\n[project.configuration]");
    Object.keys(data.configuration).forEach(function (key) {
        let value = data.configuration[key];
        // Check if the value contains backslashes
        if (value.includes("\\")) {
            value = value.replace(/\\/g, "/"); // Replace backslashes with forward slashes
        }
        toml.push(`${key} = "${value}"`);
    });

    toml.push("\n[project.llm]");
    // Append LLM provider configurations to the TOML string
    data.llm_providers.forEach((provider, index) => {
        toml.push(`\n[project.llm.${index + 1}]`);
        toml.push(`provider = "${provider.provider}"`);
        toml.push(`api_key = "${provider.api_key}"`);
        toml.push(`model = "${provider.model}"`);
        toml.push(`temperature = ${provider.temperature}`);
        toml.push(`tpm_limit = ${provider.tpm_limit}`);
        toml.push(`rpm_limit = ${provider.rpm_limit}`);
        if (provider.base_url) toml.push(`base_url = "${provider.base_url}"`);
        if (provider.endpoint_type)
            toml.push(`endpoint_type = "${provider.endpoint_type}"`);
        if (provider.region) toml.push(`region = "${provider.region}"`);
        if (provider.project_id)
            toml.push(`project_id = "${provider.project_id}"`);
        if (provider.location) toml.push(`location = "${provider.location}"`);
        if (provider.api_version)
            toml.push(`api_version = "${provider.api_version}"`);
    });

    toml.push("\n[prompt]");
    Object.keys(data.prompt).forEach(function (key) {
        toml.push(`${key} = "${data.prompt[key]}"`);
    });

    toml.push("\n[review]");
    data.review_items.forEach((review, index) => {
        toml.push(`\n[review.${index + 1}]`);
        toml.push(`key = "${review.key}"`);

        // Properly format `values` as an array of strings
        if (Array.isArray(review.values)) {
            const formattedValues = review.values
                .map((value) => `"${value}"`)
                .join(", ");
            toml.push(`values = [${formattedValues}]`);
        } else {
            toml.push(`values = []`); // Fallback if `values` is not an array
        }
    });

    return toml.join("\n");
}

function addLLMProvider() {
    const container = document.getElementById("llmProviders");
    const index = container.children.length + 1; // This index is now used only to label the sections visually

    const providerDiv = document.createElement("div");
    providerDiv.className = "llm-provider";

    // Define the model options for each provider
    const modelOptions = {
        OpenAI: [
            "gpt-5-nano",
            "gpt-5-mini",
            "gpt-5.2",
            "gpt-5.1",
            "gpt-5",
            "o4-mini",
            "o3-mini",
            "o3",
            "o1-mini",
            "o1",
            "gpt-4.1-nano",
            "gpt-4.1-mini",
            "gpt-4.1",
            "gpt-4o-mini",
            "gpt-4o",
            "gpt-4-turbo",
            "gpt-3.5-turbo",
            "",
        ],
        GoogleAI: [
            "gemini-3-flash-preview",
            "gemini-3-pro-preview",
            "gemini-2.5-flash-lite",
            "gemini-2.5-flash",
            "gemini-2.5-pro",
            "gemini-2.0-flash-lite",
            "gemini-2.0-flash",
            "gemini-1.5-flash",
            "gemini-1.5-pro",
            "",
        ],
        Cohere: [
            "command-a-reasoning-08-2025",
            "command-a-03-2025",
            "command-r-08-2024",
            "command-r7b-12-2024",
            "command-r-plus",
            "command-r",
            "command-light",
            "command",
            "",
        ],
        Anthropic: [
            "claude-4-5-haiku",
            "claude-4-5-sonnet",
            "claude-4-5-opus",
            "claude-4-0-opus",
            "claude-4-0-sonnet",
            "claude-3-7-sonnet",
            "claude-3-5-sonnet",
            "claude-3-5-haiku",
            "claude-3-opus",
            "claude-3-sonnet",
            "claude-3-haiku",
            "",
        ],
        DeepSeek: ["deepseek-chat", "deepseek-reasoner", ""],
        Perplexity: [
            "sonar-deep-research",
            "sonar-reasoning-pro",
            "sonar-pro",
            "sonar",
            "",
        ],
        "AWS Bedrock": [""],
        "Azure AI": [""],
        "Vertex AI": [""],
        SelfHosted: [""],
    };

    // HTML content for the provider
    providerDiv.innerHTML = `
        <h3 class="form-heading">Large Language Model ${index}</h3>
        <label class="form-label">Provider:</label>
        <select class="form-input provider-select">
            <option value="OpenAI">OpenAI</option>
            <option value="GoogleAI">GoogleAI</option>
            <option value="Cohere">Cohere</option>
            <option value="Anthropic">Anthropic</option>
            <option value="DeepSeek">DeepSeek</option>
            <option value="Perplexity">Perplexity</option>
            <option value="AWS Bedrock">AWS Bedrock</option>
            <option value="Azure AI">Azure AI</option>
            <option value="Vertex AI">Vertex AI</option>
            <option value="SelfHosted">Self-Hosted</option>
        </select><br>

        <label class="form-label">API Key:</label>
        <input type="text" class="form-input api-key-input"><br>

        <label class="form-label">Model:</label>
        <select class="form-input model-input"></select><br>

        <label class="form-label">Temperature:</label>
        <input type="number" class="form-input temperature-input" value="0.01" step="0.01"><br>

        <label class="form-label">Tokens Per Minute:</label>
        <input type="number" class="form-input tpm-limit-input" value="0"><br>

        <label class="form-label">Requests Per Minute:</label>
        <input type="number" class="form-input rpm-limit-input" value="0"><br>

        <div class="optional-fields" style="display:none;">
            <label class="form-label">Base URL (Self-Hosted):</label>
            <input type="text" class="form-input base-url-input"><br>
            <input type="hidden" class="endpoint-type-input">

            <label class="form-label region-label" style="display:none;">AWS Region:</label>
            <input type="text" class="form-input region-input" style="display:none;"><br>

            <label class="form-label project-id-label" style="display:none;">Project ID:</label>
            <input type="text" class="form-input project-id-input" style="display:none;"><br>

            <label class="form-label location-label" style="display:none;">Location:</label>
            <input type="text" class="form-input location-input" style="display:none;"><br>

            <label class="form-label api-version-label" style="display:none;">API Version:</label>
            <input type="text" class="form-input api-version-input" style="display:none;"><br>
        </div>
    `;

    // Append the remove button
    const removeButton = document.createElement("button");
    removeButton.textContent = "Remove";
    removeButton.type = "button";
    removeButton.style.backgroundColor = "#ffffff";
    removeButton.style.color = "#FF0000";
    removeButton.onclick = function () {
        providerDiv.remove(); // Directly remove the provider block
    };
    providerDiv.appendChild(removeButton);

    // Append the providerDiv to the container
    container.appendChild(providerDiv);

    // Get the select elements
    const providerSelect = providerDiv.querySelector(".provider-select");
    const modelSelect = providerDiv.querySelector(".model-input");

    // Function to update model options and optional fields based on the selected provider
    function updateModelOptions() {
        // Clear the current options
        modelSelect.innerHTML = "";

        // Get the selected provider and the corresponding models
        const selectedProvider = providerSelect.value;
        const models = modelOptions[selectedProvider] || [];

        // Populate the model select with the new options
        models.forEach((model) => {
            const option = document.createElement("option");
            option.value = model;
            option.textContent = model || "Default"; // Show 'Default' for empty string
            modelSelect.appendChild(option);

            // Set "Default" as the selected value
            if (model === "") {
                option.selected = true; // Mark the "Default" option as selected
            }
        });

        // Show/hide optional fields based on provider
        const optionalFields = providerDiv.querySelector(".optional-fields");
        const baseUrlInput = providerDiv.querySelector(".base-url-input");
        const endpointTypeInput = providerDiv.querySelector(
            ".endpoint-type-input",
        );
        const regionInput = providerDiv.querySelector(".region-input");
        const regionLabel = providerDiv.querySelector(".region-label");
        const projectIdInput = providerDiv.querySelector(".project-id-input");
        const projectIdLabel = providerDiv.querySelector(".project-id-label");
        const locationInput = providerDiv.querySelector(".location-input");
        const locationLabel = providerDiv.querySelector(".location-label");
        const apiVersionInput = providerDiv.querySelector(".api-version-input");
        const apiVersionLabel = providerDiv.querySelector(".api-version-label");

        // Hide all optional fields first
        optionalFields.style.display = "none";
        baseUrlInput.style.display = "none";
        regionInput.style.display = "none";
        regionLabel.style.display = "none";
        projectIdInput.style.display = "none";
        projectIdLabel.style.display = "none";
        locationInput.style.display = "none";
        locationLabel.style.display = "none";
        apiVersionInput.style.display = "none";
        apiVersionLabel.style.display = "none";

        // Reset values
        endpointTypeInput.value = "";
        baseUrlInput.value = "";
        regionInput.value = "";
        projectIdInput.value = "";
        locationInput.value = "";
        apiVersionInput.value = "";

        // Show relevant fields based on provider
        if (selectedProvider === "SelfHosted") {
            optionalFields.style.display = "block";
            baseUrlInput.style.display = "inline-block";
        } else if (selectedProvider === "AWS Bedrock") {
            optionalFields.style.display = "block";
            endpointTypeInput.value = "bedrock";
            regionInput.style.display = "inline-block";
            regionLabel.style.display = "inline";
            regionInput.value = "us-east-1";
        } else if (selectedProvider === "Azure AI") {
            optionalFields.style.display = "block";
            endpointTypeInput.value = "azure";
            baseUrlInput.style.display = "inline-block";
            apiVersionInput.style.display = "inline-block";
            apiVersionLabel.style.display = "inline";
            apiVersionInput.value = "2024-02-15-preview";
        } else if (selectedProvider === "Vertex AI") {
            optionalFields.style.display = "block";
            endpointTypeInput.value = "vertex";
            projectIdInput.style.display = "inline-block";
            projectIdLabel.style.display = "inline";
            locationInput.style.display = "inline-block";
            locationLabel.style.display = "inline";
            locationInput.value = "us-central1";
        }
    }

    // Initialize the model options on creation
    updateModelOptions();

    // Add event listener to update models when the provider changes
    providerSelect.addEventListener("change", updateModelOptions);
}

function removeLLMProvider(element) {
    if (element) {
        element.parentNode.removeChild(element);
    }
}

function addReviewBlock() {
    const container = document.getElementById("reviews");

    // Create the review block div
    const reviewDiv = document.createElement("div");
    reviewDiv.className = "review-item";

    // Set up the innerHTML for reviewDiv using classes instead of IDs
    reviewDiv.innerHTML = `
        <h3 class="form-heading">Review Block</h3>
        <label class="form-label">Key:</label>
        <input type="text" class="form-input review-key"><br>

        <label class="form-label">Values:</label>
        <input type="text" class="form-input review-values" placeholder="Enter comma-separated values"><br>
    `;

    // Create and configure the remove button
    const removeButton = document.createElement("button");
    removeButton.textContent = "Remove";
    removeButton.type = "button";
    removeButton.style.backgroundColor = "#ffffff";
    removeButton.style.color = "#FF0000";
    removeButton.onclick = function () {
        removeReviewBlock(reviewDiv);
    };
    reviewDiv.appendChild(removeButton);

    // Append the review block to the container
    container.appendChild(reviewDiv);
}

function removeReviewBlock(element) {
    if (element) {
        element.parentNode.removeChild(element);
    }
}

function downloadConfiguration() {
    var text = document.getElementById("configOutput").value; // Get the content from textarea
    var filename = "configuration.toml"; // Define a filename

    var blob = new Blob([text], { type: "text/plain" });

    var downloadLink = document.createElement("a");
    downloadLink.href = window.URL.createObjectURL(blob);
    downloadLink.download = filename;

    // Append the link to the document, click it, and then remove it
    document.body.appendChild(downloadLink);
    downloadLink.click();
    document.body.removeChild(downloadLink);
}
