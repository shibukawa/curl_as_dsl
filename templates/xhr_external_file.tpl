<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>XHR Test</title>
    <style>
    div.droptarget {
        border : solid 1px black;
        padding : 10px;
        background-color : peachpuff;
    }
    </style>
</head>
<body>
<div id="drop" class="droptarget">drop file here</div>
<script>

function request(file) {
    var xhr = new XMLHttpRequest();{{ .PrepareBody }}
    xhr.open("{{ .Method }}", {{ .Url }}, true);
    {{ .PrepareOptions }}
    xhr.onreadystatechange = function(e) {
        if (this.readyState == 4) {
            document.write("<p>body:" + this.responseText + "</p>");
            document.write("<p>status:" + this.status + "</p>");
        }
    };
    xhr.send({{ .Body }});
}

function handleDragOver(e) {
    e.preventDefault();
}

function handleDrop(e) {
    e.stopPropagation();
    e.preventDefault();
    var file = e.dataTransfer.files[0];
    this.innerHTML = file.name;
    {{ .PrepareFile }}
}

window.onload = function () {
    var tag = document.querySelector('.droptarget');
    tag.addEventListener('dragover', handleDragOver, false);
    tag.addEventListener('drop', handleDrop, false);
}
</script>
</body>
</html>
