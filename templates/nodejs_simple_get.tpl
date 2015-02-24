{{ range $key, $_ := .Modules }}var {{ $key }} = require("{{ $key }}");
{{end}}
{{ .ClientModule }}.get({{ .Url }}, function(res) {
    console.log("Got response: " + res.statusCode + " " + res.statusMessage);
    res.on('data', function (chunk) {
        console.log('BODY: ' + chunk);
    });
}).on('error', function(e) {
    console.log("Got error: " + e.message);
});
