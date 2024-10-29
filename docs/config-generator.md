---
title: Config Generator
layout: default
---

# Generate Your Review Configuration File

<form id="configForm">
    <label for="name">Project Name:</label>
    <input type="text" id="name" name="name"><br>

    <label for="author">Author:</label>
    <input type="text" id="author" name="author"><br>

    <input type="button" value="Generate TOML" onclick="generateTOML()">
</form>
<textarea id="output" rows="10" cols="50"></textarea>

<script src="/assets/js/tomlGenerator.js"></script>
