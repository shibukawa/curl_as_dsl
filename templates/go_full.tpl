package main

import (
{{ range $key, $_ := .Modules }}    "{{ $key }}"
{{end}})
{{ .AdditionalDeclaration }}
func main() {
    {{ .PrepareClient }}
    client := &http.Client{{ .ClientBody }}
    {{ .Data }}
    request, err := http.NewRequest("{{ .Method }}", {{ .Url }}, {{ .DataVariable }})
    {{ .ModifyRequest }}
    resp, err := client.Do(request)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    log.Print(string(body))
}
