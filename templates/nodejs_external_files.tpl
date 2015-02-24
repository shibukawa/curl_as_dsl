var fs = require("fs");
{{ range $key, $_ := .Modules }}var {{ $key }} = require("{{ $key }}");
{{end}}{{ .AdditionalDeclaration }}
Promise.all([
{{ range $i, $externalFile := .ExternalFiles }}    new Promise(function (success, reject) {
        fs.readFile("{{$externalFile.FileName}}"{{if $externalFile.TextType }}, {encoding: "utf8"}{{end}}, function (err, data) {
            if (err) { reject(err); } else { success(data); }
        });
    }),
{{end}}]).then(function (fileContents) {
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
});