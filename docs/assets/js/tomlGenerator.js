document.addEventListener('DOMContentLoaded', function() {
    var button = document.getElementById('generateConfigButton');
    if (button) {
        button.addEventListener('click', generateConfig);
    }
});

function generateConfig() {
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
        }
    };

    var toml = generateTOMLString(data);
    document.getElementById('configOutput').value = toml;
}

function generateTOMLString(data) {
    var toml = [];
    toml.push("[project]");
    Object.keys(data.project).forEach(function(key) {
        toml.push(key + ' = "' + data.project[key] + '"');
    });
    toml.push("\n[project.configuration]");
    Object.keys(data.configuration).forEach(function(key) {
        toml.push(key + ' = "' + data.configuration[key] + '"');
    });
    return toml.join("\n");
}