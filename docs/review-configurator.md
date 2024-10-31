---
title: Config Generator
layout: default
---

# Generate Your Review Configuration File

<form id="configForm">
    <h2>Project Information</h2>
    <label for="name">Project Name:</label>
    <input type="text" id="name" name="name" value="Review Project Title"><br>

    <label for="author">Project Author:</label>
    <input type="text" id="author" name="author" value="First Lastname"><br>

    <label for="version">Version:</label>
    <input type="text" id="version" name="version" value="0.1"><br>

    <h2>Project Configuration</h2>
    <label for="input_directory">Input Directory:</label>
    <input type="text" id="input_directory" name="input_directory" value="/path/to/txt/files"><br>

    <label for="input_conversion">Input Conversion:</label>
     <select id="input_conversion" name="input_conversion">
        <option value="" selected>None</option>
        <option value="pdf">PDF</option>
        <option value="docx">DOCX</option>
        <option value="html">HTML</option>
        <option value="pdf,docx">PDF+DOCX</option>
        <option value="pdf,html">PDF+HTML</option>
        <option value="docx,html">DOCX+HTML</option>
        <option value="pdf,docx,html">PDF+DOCX+HTML</option>
    </select><br>

    <label for="results_file_name">Results File Name:</label>
    <input type="text" id="results_file_name" name="results_file_name" value="/path/to/save/results_file"><br>

    <label for="output_format">Output Format:</label>
    <select id="output_format" name="output_format">
        <option value="csv" selected>CSV</option>
        <option value="json">JSON</option>
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

    <h2>LLM Configuration</h2>
    <div id="llmProviders">
        <!-- LLM providers will be added dynamically here -->
    </div>
    <button type="button" onclick="addLLMProvider()" style="background-color: #ffffff;">Add LLModel</button>
    <br><br>

    <h2>Prompt Components</h2>
    <label for="persona">Persona:</label>
        <input type="text" id="persona" name="persona" style="width: 70%;" value="You are an experienced scientist working on a systematic review of the literature."><br>

    <label for="task">Task:</label>
        <input type="text" id="task" name="task" style="width: 70%;" value="You are asked to map the concepts discussed in a scientific paper attached here."><br>

    <label for="expected_result">Expected Result:</label>
        <input type="text" id="expected_result" name="expected_result" style="width: 70%;" value="You should output a JSON object with the following keys and possible values: "><br>

    <label for="definitions">Definitions:</label>
            <input type="text" id="definitions" name="definitions" style="width: 70%;" value=""><br>

    <label for="example">Examples:</label>
        <input type="text" id="example" name="example" style="width: 70%;" value=""><br>

    <label for="failsafe">Failsafe:</label>
        <input type="text" id="failsafe" name="failsafe" style="width: 70%;" value="If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value."><br>


    <h2>Review Items</h2>
    <div id="reviews">
        <!-- Review items will be added dynamically here -->
    </div>
    <button type="button" onclick="addReviewBlock()" style="background-color: #ffffff;">Add Review Item</button>
    <br><br>


    <button type="button" id="generateConfigButton" style="background-color: #0056b3; color: #ffffff;">Generate Configuration</button>
</form>

<textarea id="configOutput" rows="20" cols="70"></textarea>

<script src="assets/js/tomlGenerator.js"></script>


