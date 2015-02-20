{{ range $key, $_ := .Modules }}import {{ $key }}
{{end}}{{ .AdditionalDeclaration }}
def main():
    conn = http.client.{{ .ConnectionClass }}("{{ .Host }}")
    {{ .Proxy }}{{ .PrepareBody }}{{ .PrepareHeader }}
    conn.request("{{ .Method }}", {{ .Path }}{{if .HasBody}}, body={{ .Body }}{{end}}{{if .HasHeader}}, header={{ .Header }}{{end}})
    res = conn.getresponse()
    print(res.status, res.reason)
    print(res.read())
    conn.close()

if __name__ == "__main__":
    main()
