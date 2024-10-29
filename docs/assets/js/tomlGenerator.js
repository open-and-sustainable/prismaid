function generateTOML() {
    var name = document.getElementById('name').value;
    var author = document.getElementById('author').value;

    var tomlContent = `name = "${name}"\nauthor = "${author}"`;
    document.getElementById('output').value = tomlContent;
}
