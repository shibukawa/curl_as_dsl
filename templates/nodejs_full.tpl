{{ range $key, $_ := .Modules }}var {{ $key }} = require("{{ $key }}");
{{end}}{{ .AdditionalDeclaration }}
{{ .PrepareBody }}var req = {{ .ClientModule }}.request({
    host: "{{ .Host }}",
    path: {{ .Path }},{{if ne .Port 0}}
    port: {{ .Port }},{{end}}
    method: "{{ .Method }}",{{ .PrepareOptions }}
}, function(res) {
    console.log("Got response: " + res.statusCode + " " + res.statusMessage);
    res.on('data', function (chunk) {
        console.log('BODY: ' + chunk);
    });
});
{{ range $_, $line := .BodyLines}}req.write({{ $line }});
{{end}}req.end();
req.on('error', function(e) {
    console.log("Got error: " + e.message);
});