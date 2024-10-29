---
title: Config Generator
layout: default
---

# Generate Your Review Configuration File

<form id="configForm">
    <h2>Project Information</h2>
    <label for="name">Project Name:</label>
    <input type="text" id="name" name="name" value="Use of LLM for systematic review"><br>

    <label for="author">Project Author:</label>
    <input type="text" id="author" name="author" value="John Doe"><br>

    <label for="version">Version:</label>
    <input type="text" id="version" name="version" value="1.0"><br>

    <h2>Project Configuration</h2>
    <label for="input_directory">Input Directory:</label>
    <input type="text" id="input_directory" name="input_directory" value="/path/to/txt/files"><br>

    <label for="input_conversion">Input Conversion:</label>
    <input type="text" id="input_conversion" name="input_conversion" value=""><br>

    <label for="results_file_name">Results File Name:</label>
    <input type="text" id="results_file_name" name="results_file_name" value="/path/to/save/results"><br>

    <label for="output_format">Output Format:</label>
    <select id="output_format" name="output_format">
        <option value="csv">CSV</option>
        <option value="json" selected>JSON</option>
    </select><br>

    <label for="log_level">Log Level:</label>
    <select id="log_level" name="log_level">
        <option value="low" selected>Low</option>
        <option value="medium">Medium</option>
        <option value="high">High</option>
    </select><br>

    <label for="duplication">Duplication:</label>
    <select id="duplication" name="duplication">
        <option value="no" selected>No</option>
        <option value="yes">Yes</option>
    </select><br>

    <label for="cot_justification">CoT Justification:</label>
    <select id="cot_justification" name="cot_justification">
        <option value="no" selected>No</option>
        <option value="yes">Yes</option>
    </select><br>

    <label for="summary">Summary:</label>
    <select id="summary" name="summary">
        <option value="no" selected>No</option>
        <option value="yes">Yes</option>
    </select><br>

    <button type="button" onclick="generateConfig()">Generate Configuration</button>
</form>

<textarea id="configOutput" rows="20" cols="70"></textarea>


<script src="assets/js/tomlGenerator.js"></script>


