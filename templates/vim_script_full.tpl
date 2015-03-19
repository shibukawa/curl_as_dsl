{{ .AdditionalDeclaration }}{{ .PrepareBody }}{{ .PrepareHeader }}let s:res = webapi#http#{{ .Method }}({{ .Url }}{{ .BodyContent }}{{ .Header }})
echo s:res.status
echo s:res.message
echo s:res.content
unlet! s:res{{if .HasHeader}}
unlet! s:headers{{end}}{{ .FinalizeBody }}