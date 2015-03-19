package generator

import (
	"bytes"
	"fmt"
	"github.com/shibukawa/curl_as_dsl/client/golang"
	"github.com/shibukawa/curl_as_dsl/client/java"
	"github.com/shibukawa/curl_as_dsl/client/nodejs"
	"github.com/shibukawa/curl_as_dsl/client/objc"
	"github.com/shibukawa/curl_as_dsl/client/php"
	"github.com/shibukawa/curl_as_dsl/client/python"
	"github.com/shibukawa/curl_as_dsl/client/vimscript"
	"github.com/shibukawa/curl_as_dsl/client/xhr"
	"github.com/shibukawa/curl_as_dsl/common"
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
	"vim":                "vim",
}

func render(lang, key string, options interface{}) string {
	src, _ := Asset(fmt.Sprintf("templates/%s_%s.tpl", lang, key))
	var buffer bytes.Buffer
	var err error
	tmpTpl, err := template.New(key).Parse(string(src))
	tpl := template.Must(tmpTpl, err)
	err = tpl.Execute(&buffer, options)
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

func GenerateCode(target string, curlOptions *common.CurlOptions) (string, string, string, interface{}) {
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
		templateName, option = golang.ProcessCurlCommand(curlOptions)
	case "python":
		langName = "python"
		templateName, option = python.ProcessCurlCommand(curlOptions)
	case "node":
		langName = "nodejs"
		templateName, option = nodejs.ProcessCurlCommand(curlOptions)
	case "java":
		langName = "java"
		templateName, option = java.ProcessCurlCommand(curlOptions)
	case "objc_nsurlsession":
		langName = "objc_nsurlsession"
		templateName, option = objc.ProcessCurlCommand(curlOptions)
	case "objc_nsurlconnection":
		langName = "objc_nsurlconnection"
		templateName, option = objc.ProcessCurlCommand(curlOptions)
	case "xhr":
		langName = "xhr"
		templateName, option = xhr.ProcessCurlCommand(curlOptions)
	case "php":
		langName = "php"
		templateName, option = php.ProcessCurlCommand(curlOptions)
	case "vim":
		langName = "vim_script"
		templateName, option = vimscript.ProcessCurlCommand(curlOptions)
	default:
	}
	sourceCode := render(langName, templateName, option)
	return sourceCode, langName, templateName, option
}
