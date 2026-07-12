---
title: Review Configurator
layout: default
---

<link rel="stylesheet" href="/assets/css/styles.css">

# Review Configurator

---

This configurator helps you create a TOML configuration file for the prismAId [Review tool](../tools/review-tool), which processes systematic literature reviews using AI models. For other tools in the prismAId toolkit, see the [Download](../tools/download-tool) and [Convert](../tools/convert-tool) tool pages.

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
        <p class="description" style="font-style: italic;">Specify the directory where the input text files are located for review.</p>
        <label for="input_directory" class="form-label">Input Directory:</label>
        <input type="text" id="input_directory" name="input_directory" value="/path/to/txt/files" class="form-input"><br>
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
        <p class="description" style="font-style: italic;">Set the log level for the process. Low logs minimal information, medium logs more and prints out on screen, and high logs all details also on file.</p>
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
        <select id="summary" name="summary" class="form-input">
            <option value="no" selected>No</option>
            <option value="yes">Yes</option>
        </select><br>
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

    <h2 id="revaise-documentation">RevAIse Documentation (Optional)</h2>
    <div class="form-group">
        <p class="description" style="font-style: italic;">Optionally document this review as a data-extraction stage in a shared <a href="https://revaise-model.readthedocs.io/stable/">RevAIse</a> review record. Disabled by default; normal review outputs are unchanged when enabled.</p>
        <label for="revaise_enabled" class="form-label">Enable RevAIse:</label>
        <select id="revaise_enabled" name="revaise_enabled" class="form-input">
            <option value="no" selected>No</option>
            <option value="yes">Yes</option>
        </select><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Path to the RevAIse record to create or update. Reuse the same file across stages to build one cumulative record.</p>
        <label for="revaise_record_file" class="form-label">Record File:</label>
        <input type="text" id="revaise_record_file" name="revaise_record_file" value="review.revaise.json" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Record format. If omitted, it is detected from the record file (JSON by default).</p>
        <label for="revaise_format" class="form-label">Format:</label>
        <select id="revaise_format" name="revaise_format" class="form-input">
            <option value="json" selected>JSON</option>
            <option value="yaml">YAML</option>
        </select><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">RevAIse schema version to record.</p>
        <label for="revaise_schema_version" class="form-label">Schema Version:</label>
        <input type="text" id="revaise_schema_version" name="revaise_schema_version" value="0.7.1" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">How much a human reviews the AI output. Defaults to None, since prismAId itself performs no human review; raise it to reflect the review you actually carry out.</p>
        <label for="revaise_human_oversight" class="form-label">Human Oversight Level:</label>
        <select id="revaise_human_oversight" name="revaise_human_oversight" class="form-input">
            <option value="NONE" selected>None</option>
            <option value="MINIMAL">Minimal</option>
            <option value="EXCEPTION_ONLY">Exception only</option>
            <option value="CONFIDENCE_BASED">Confidence based</option>
            <option value="SAMPLE_REVIEW">Sample review</option>
            <option value="FULL_REVIEW">Full review</option>
        </select><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Human-readable label for this extraction stage.</p>
        <label for="revaise_stage_label" class="form-label">Stage Label:</label>
        <input type="text" id="revaise_stage_label" name="revaise_stage_label" value="AI-assisted extraction" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Run identifier. Reuse the same id to update the same run; change it to record a new one (e.g. a pilot vs. a full extraction).</p>
        <label for="revaise_run_id" class="form-label">Run ID:</label>
        <input type="text" id="revaise_run_id" name="revaise_run_id" value="full_extraction_001" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Human-readable label for this extraction run.</p>
        <label for="revaise_run_label" class="form-label">Run Label:</label>
        <input type="text" id="revaise_run_label" name="revaise_run_label" value="Full extraction on included studies" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Extraction form identifier, name, and version.</p>
        <label for="revaise_form_id" class="form-label">Form ID:</label>
        <input type="text" id="revaise_form_id" name="revaise_form_id" value="extraction_form_v1" class="form-input"><br>
        <label for="revaise_form_name" class="form-label">Form Name:</label>
        <input type="text" id="revaise_form_name" name="revaise_form_name" value="Extraction form" class="form-input"><br>
        <label for="revaise_form_version" class="form-label">Form Version:</label>
        <input type="text" id="revaise_form_version" name="revaise_form_version" value="1" class="form-input"><br>
    </div>

    <div class="form-group">
        <p class="description" style="font-style: italic;">Identifier of the extractor performing this run.</p>
        <label for="revaise_extractor_id" class="form-label">Extractor ID:</label>
        <input type="text" id="revaise_extractor_id" name="revaise_extractor_id" value="prismaid" class="form-input"><br>
    </div>

</form>

## Generate Configuration
<button type="button" id="generateConfigButton" style="background-color: #0056b3; color: #ffffff;">Generate TOML</button>

<textarea id="configOutput" class="wide-textarea"></textarea>

## Download Configuration
<button type="button" id="downloadButton" style="background-color: #0056b3; color: #ffffff;">Download File</button>

<script src="/assets/js/tomlGenerator.js"></script>


<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
