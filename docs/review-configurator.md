---
title: Config Generator
layout: default
---

<link rel="stylesheet" href="assets/css/styles.css">

# Generate Your Review Configuration File

<form id="configForm">
    <h2 id="project-information">Project Information</h2>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Enter the name of your project.</p>
        <label for="name" class="form-label">Project Name:</label>
        <input type="text" id="name" name="name" value="Review Project Title" class="form-input"><br>
    </div>
 
    <div class="form-group">
        <p class="description" style="font-style: italic;">Enter the author of the project.</p>
        <label for="author" class="form-label">Project Author:</label>
        <input type="text" id="author" name="author" value="First Lastname" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Specify the version of the project configuration.</p>
        <label for="version" class="form-label">Version:</label>
        <input type="text" id="version" name="version" value="0.1" class="form-input"><br>
    </div>

    <h2 id="project-configuration">Project Configuration</h2>
    <div class="form-group">
        <p class="description" style="font-style: italic;">Specify the directory where the input text files are located (if not a Zotero based review project).</p>
        <label for="input_directory" class="form-label">Input Directory:</label>
        <input type="text" id="input_directory" name="input_directory" value="/path/to/txt/files" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Choose the input conversion format(s), if necessary (and not a Zotero based review project). You can select multiple formats for conversion.</p>
        <label for="input_conversion" class="form-label">Input Conversion:</label>
        <select id="input_conversion" name="input_conversion" class="form-input">
            <option value="" selected>None</option>
            <option value="pdf">PDF</option>
            <option value="docx">DOCX</option>
            <option value="html">HTML</option>
            <option value="pdf,docx">PDF+DOCX</option>
            <option value="pdf,html">PDF+HTML</option>
            <option value="docx,html">DOCX+HTML</option>
            <option value="pdf,docx,html">PDF+DOCX+HTML</option>
        </select><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Specify the file name and path where the results will be saved. Ensure the path exists.</p>
        <label for="results_file_name" class="form-label">Results File Name:</label>
        <input type="text" id="results_file_name" name="results_file_name" value="/path/to/save/results_file" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Select the format for the output results. Available options are CSV and JSON.</p>
        <label for="output_format" class="form-label">Output Format:</label>
        <select id="output_format" name="output_format" class="form-input">
            <option value="csv" selected>CSV</option>
            <option value="json">JSON</option>
        </select><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Set the log level for the process. Low logs minimal information, medium logs more, and high logs all details.</p>
        <label for="log_level" class="form-label">Log Level:</label>
        <select id="log_level" name="log_level" class="form-input">
            <option value="low" selected>Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
        </select><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Choose whether to enable duplication for debugging purposes. This will run model queries twice.</p>
        <label for="duplication" class="form-label">Duplication:</label>
        <select id="duplication" name="duplication" class="form-input">
            <option value="no" selected>No</option>
            <option value="yes">Yes</option>
        </select><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Enable Chain of Thought (CoT) justification to request and save model reasoning for answers provided.</p>
        <label for="cot_justification" class="form-label">CoT Justification:</label>
        <select id="cot_justification" name="cot_justification" class="form-input">
            <option value="no" selected>No</option>
            <option value="yes">Yes</option>
        </select><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Choose whether to generate and save summaries of the manuscript.</p>
        <label for="summary" class="form-label">Summary:</label>
        <select id="summary" name="summarthe name of the collection or group containing the document to reviewy" class="form-input">
            <option value="no" selected>No</option>
            <option value="yes">Yes</option>
        </select><br>
    </div>

    <h2 id="project-zotero-integration">Project Zotero Integration</h2>
    <div class="form-group">
        <p class="description" style="font-style: italic;">Specify the user ID accessible at https://www.zotero.org/settings/security.</p>
        <label for="user" class="form-label">User ID:</label>
        <input type="text" id="user" name="user" value="" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Specify the private key created at https://www.zotero.org/settings/security.</p>
        <label for="api_key" class="form-label">API key:</label>
        <input type="text" id="api_key" name="api_key" value="" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Specify the name of the collection or group containing the document to review.</p>
        <label for="group" class="form-label">Group:</label>
        <input type="text" id="group" name="group" value="" class="form-input"><br>
    </div>

    <h2 id="llm-configuration">LLM Configuration</h2>
    <div id="llmProviders">
        <!-- LLM providers will be added dynamically here -->
    </div>
    <button type="button" onclick="addLLMProvider()" style="background-color: #ffffff;">Add LLModel</button>
    <br><br>

    <h2 id="prompt-components">Prompt Components</h2>
    <div class="form-group">
        <p class="description" style="font-style: italic;">Describe the persona or role the model should adopt when performing the task.</p>
        <label for="persona" class="form-label">Persona:</label>
        <input type="text" id="persona" name="persona" class="form-long-input" value="You are an experienced scientist working on a systematic review of the literature."><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Specify the task the model should perform.</p>
        <label for="task" class="form-label">Task:</label>
        <input type="text" id="task" name="task" class="form-long-input" value="You are asked to map the concepts discussed in a scientific paper attached here."><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Outline the expected result in detail, in particular stating you are looking for a JSON response object.</p>
        <label for="expected_result" class="form-label">Expected Result:</label>
        <input type="text" id="expected_result" name="expected_result" class="form-long-input" value="You should output a JSON object with the following keys and possible values: "><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Provide any specific definitions to avoid misunderstandings.</p>
        <label for="definitions" class="form-label">Definitions:</label>
        <input type="text" id="definitions" name="definitions" class="form-long-input" value=""><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Include examples to clarify the request.</p>
        <label for="example" class="form-label">Examples:</label>
        <input type="text" id="example" name="example" class="form-long-input" value=""><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Specify a failsafe response if the concepts are unclear or missing.</p>
        <label for="failsafe" class="form-label">Failsafe:</label>
        <input type="text" id="failsafe" name="failsafe" class="form-long-input" value="If the concepts neither are clearly discussed in the document nor they can be deduced from the text, respond with an empty '' value."><br>
    </div>


    <h2 id="review-items">Review Items</h2>
    <div id="reviews">
        <!-- Review items will be added dynamically here -->
    </div>
    <button type="button" onclick="addReviewBlock()" style="background-color: #ffffff;">Add Review Item</button>
    <br><br>

</form>

## Generate Configuration
<button type="button" id="generateConfigButton" style="background-color: #0056b3; color: #ffffff;">Generate TOML</button>

<textarea id="configOutput" class="wide-textarea"></textarea>

## Download Configuration
<button type="button" id="downloadButton" style="background-color: #0056b3; color: #ffffff;">Download File</button>

<script src="assets/js/tomlGenerator.js"></script>


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>

