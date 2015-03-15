package httpgen_generator

import (
	"bytes"
	"fmt"
	"github.com/shibukawa/curl_as_dsl/go_client"
	"github.com/shibukawa/curl_as_dsl/httpgen_common"
	"github.com/shibukawa/curl_as_dsl/java_client"
	"github.com/shibukawa/curl_as_dsl/nodejs_client"
	"github.com/shibukawa/curl_as_dsl/objc_client"
	"github.com/shibukawa/curl_as_dsl/php_client"
	"github.com/shibukawa/curl_as_dsl/python_client"
	"github.com/shibukawa/curl_as_dsl/xhr_client"
	"go/format"
	"log"
	"text/template"
)

var LanguageMap map[string]string = map[string]string{
	"go":                 "go",
	"golang":             "go",
	"py":                 "python",
	"python":             "python",
	"node":               "node",
	"nodejs":             "node",
	"js.node":            "node",
	"javascript.node":    "node",
	"xhr":                "xhr",
	"js.xhr":             "xhr",
	"javascript.xhr":     "xhr",
	"js.browser":         "xhr",
	"javascript.browser": "xhr",
	"java":               "java",
	"objc":               "objc_nsurlsession",
	"objc.session":       "objc_nsurlsession",
	"objc.nsurlsession":  "objc_nsurlsession",
	"objc.connection":    "objc_nsurlconnection",
	"objc.urlconnection": "objc_nsurlconnection",
	"php":                "php",
}

func render(lang, key string, options interface{}) string {
	src, _ := Asset(fmt.Sprintf("templates/%s_%s.tpl", lang, key))
	tpl := template.Must(template.New(key).Parse(string(src)))
	var buffer bytes.Buffer
	err := tpl.Execute(&buffer, options)
	if err != nil {
		log.Fatal(err)
	}
	if lang == "go" {
		gosrc, err := format.Source(buffer.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		return string(gosrc)
	}
	return buffer.String()
}

func GenerateCode(target string, curlOptions *httpgen_common.CurlOptions) (string, string, string, interface{}) {
	var langName string
	var templateName string
	var option interface{}

	lang, ok := LanguageMap[target]
	if !ok {
		return "", "", "", nil
	}

	switch lang {
	case "go":
		langName = "go"
		templateName, option = go_client.ProcessCurlCommand(curlOptions)
	case "python":
		langName = "python"
		templateName, option = python_client.ProcessCurlCommand(curlOptions)
	case "node":
		langName = "nodejs"
		templateName, option = nodejs_client.ProcessCurlCommand(curlOptions)
	case "java":
		langName = "java"
		templateName, option = java_client.ProcessCurlCommand(curlOptions)
	case "objc_nsurlsession":
		langName = "objc_nsurlsession"
		templateName, option = objc_client.ProcessCurlCommand(curlOptions)
	case "objc_nsurlconnection":
		langName = "objc_nsurlconnection"
		templateName, option = objc_client.ProcessCurlCommand(curlOptions)
	case "xhr":
		langName = "xhr"
		templateName, option = xhr_client.ProcessCurlCommand(curlOptions)
	case "php":
		langName = "php"
		templateName, option = php_client.ProcessCurlCommand(curlOptions)
	default:
	}
	sourceCode := render(langName, templateName, option)
	return sourceCode, langName, templateName, option
}
