<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>XHR Test</title>
</head>
<body>
<script>
function request() {
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
window.onload = function () {
    request();
};
</script>
</body>
</html>
