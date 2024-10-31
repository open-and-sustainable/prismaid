document.addEventListener('DOMContentLoaded', function() {
    // Ensure the button exists and is correctly targeted
    var button = document.getElementById('generateConfigButton');
    if (button) {
        button.addEventListener('click', generateConfig);
    } else {
        console.error('Generate Config Button not found!');
    }
});

function generateConfig() {
    // Gather data from form fields
    var data = {
        project: {
            name: document.getElementById('name').value,
            author: document.getElementById('author').value,
            version: document.getElementById('version').value,
        },
        configuration: {
            input_directory: document.getElementById('input_directory').value,
            input_conversion: document.getElementById('input_conversion').value,
            results_file_name: document.getElementById('results_file_name').value,
            output_format: document.getElementById('output_format').value,
            log_level: document.getElementById('log_level').value,
            duplication: document.getElementById('duplication').value,
            cot_justification: document.getElementById('cot_justification').value,
            summary: document.getElementById('summary').value,
        },
        llm_providers: [],
        prompt: {
            persona: document.getElementById('persona').value,
            task: document.getElementById('task').value,
            expected_result: document.getElementById('expected_result').value,
            definitions: document.getElementById('definitions').value,
            example: document.getElementById('example').value,
            failsafe: document.getElementById('failsafe').value,
        },
        review_items: []
    };

    // Collect data from dynamically added LLM providers
    const providers = document.querySelectorAll('.llm-provider');
    providers.forEach((provider, index) => {
        const providerData = {
            provider: provider.querySelector(`#provider${index + 1}`).value,
            api_key: provider.querySelector(`#api_key${index + 1}`).value,
            model: provider.querySelector(`#model${index + 1}`).value,
            temperature: provider.querySelector(`#temperature${index + 1}`).value,
            tpm_limit: provider.querySelector(`#tpm_limit${index + 1}`).value,
            rpm_limit: provider.querySelector(`#rpm_limit${index + 1}`).value
        };
        data.llm_providers.push(providerData);
    });

    // Collect data from dynamically added review items
    const reviews = document.querySelectorAll('.review');
    reviews.forEach((review, index) => {
        const reviewData = {
            key: review.querySelector(`#key${index + 1}`).value,
            values: review.querySelector(`#values${index + 1}`).value
        };
        data.review_items.push(reviewData);
    });

    // Generate TOML string from data
    var toml = generateTOMLString(data);
    document.getElementById('configOutput').value = toml;
}

function generateTOMLString(data) {
    // Build TOML string from the structured data
    var toml = ["[project]"];
    Object.keys(data.project).forEach(function(key) {
        toml.push(`${key} = "${data.project[key]}"`);
    });

    toml.push("\n[project.configuration]");
    Object.keys(data.configuration).forEach(function(key) {
        toml.push(`${key} = "${data.configuration[key]}"`);
    });

    toml.push("\n[project.llm]");
    // Append LLM provider configurations to the TOML string
    data.llm_providers.forEach((provider, index) => {
        toml.push(`\n[project.llm.${index + 1}]`);
        Object.keys(provider).forEach(key => {
            toml.push(`${key} = "${provider[key]}"`);
        });
    });

    toml.push("\n[prompt]");
    Object.keys(data.prompt).forEach(function(key) {
        toml.push(`${key} = "${data.prompt[key]}"`);
    });

    toml.push("\n[review]");
    // Append review items to the TOML string
    data.review_items.forEach((review, index) => {
        toml.push(`\n[review.${index + 1}]`);
        Object.keys(review).forEach(key => {
            toml.push(`${key} = "${review[key]}"`);
        });
    });

    return toml.join("\n");
}

function addLLMProvider() {
    const container = document.getElementById('llmProviders');
    const index = container.children.length + 1;

    const providerDiv = document.createElement('div');
    providerDiv.className = 'llm-provider';
    providerDiv.id = `llmProvider${index}`;

    providerDiv.innerHTML = `
        <h3>Large Language Model ${index}</h3>
        <label for="provider${index}">Provider:</label>
        <select id="provider${index}" name="provider${index}">
            <option value="OpenAI">OpenAI</option>
            <option value="GoogleAI">GoogleAI</option>
            <option value="Cohere">Cohere</option>
            <option value="Anthropic">Anthropic</option>
        </select><br>
        <label for="api_key${index}">API Key:</label>
        <input type="text" id="api_key${index}" name="api_key${index}"><br>
        <label for="model${index}">Model:</label>
        <input type="text" id="model${index}" name="model${index}"><br>
        <label for="temperature${index}">Temperature:</label>
        <input type="number" id="temperature${index}" value="0.01" name="temperature${index}" step="0.01"><br>
        <label for="tpm_limit${index}">Tokens Per Minute:</label>
        <input type="number" id="tpm_limit${index}" value="0" name="tpm_limit${index}"><br>
        <label for="rpm_limit${index}">Requests Per Minute:</label>
        <input type="number" id="rpm_limit${index}" value="0" name="rpm_limit${index}"><br>
    `;
    
    const removeButton = document.createElement('button');
    removeButton.textContent = 'Remove';
    removeButton.type = 'button';
    removeButton.onclick = function() { removeLLMProvider(index); };
    providerDiv.appendChild(removeButton);

    container.appendChild(providerDiv);
}

function removeLLMProvider(index) {
    const element = document.getElementById('llmProvider' + index);
    if (element) {
        element.parentNode.removeChild(element);
    }
}

function addReviewBlock() {
    const container = document.getElementById('review_items');
    const index = container.children.length + 1;

    const reviewDiv = document.createElement('div');
    reviewDiv.className = 'review';
    reviewDiv.id = `review${index}`;

    reviewDiv.innerHTML = `
        <h3>Review Block ${index}</h3>
        <label for="key${index}">Key:</label>
        <input type="text" id="key${index}" name="key${index}"><br>
        <label for="values${index}">Values:</label>
        <input type="text" id="values${index}" name="values${index}" placeholder="Enter comma-separated values"><br>
    `;

    const removeButton = document.createElement('button');
    removeButton.textContent = 'Remove';
    removeButton.type = 'button';
    removeButton.onclick = function() { removeReviewBlock(index); };
    reviewDiv.appendChild(removeButton);

    container.appendChild(reviewDiv);
}

function removeReviewBlock(index) {
    const element = document.getElementById('review' + index);
    if (element) {
        element.parentNode.removeChild(element);
    }
}

