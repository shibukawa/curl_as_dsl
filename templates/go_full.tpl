package main

import (
{{ range $key, $_ := .Modules }}    "{{ $key }}"
{{end}})

func main() {
    client := &http.Client{{ .ClientBody }}{{ if urlprepare}}
    {{ .UrlPrepare }} {{end}}{{ if bodyprepare }}
    {{ .BodyPrepare }}
    request, err := http.NewRequest("{{ .Method }}", "{{ .Url }}{{ .UrlExtra }}", {{ .Body }})
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
