package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/jessevdk/go-flags"
	"github.com/shibukawa/curl_as_dsl/httpgen_common"
	"github.com/shibukawa/curl_as_dsl/httpgen_generator"
	"github.com/shibukawa/optstring_parser"
	"honnef.co/go/js/console"
	"html"
	"strings"
)

type GlobalOptions struct{}

func GenerateCode(target, options string) (string, string) {
	var globalOptions GlobalOptions
	var curlOptions httpgen_common.CurlOptions
	curlOptions.Init()

	parser := flags.NewParser(&globalOptions, flags.Default)
	curlCommand, err := parser.AddCommand("curl",
		"Generate code from curl options",
		"This command has almost same options of curl and generate code",
		&curlOptions)

	if !strings.HasPrefix(options, "curl ") {
		options = "curl " + options
	}
	args := optstring_parser.Parse(options)

	urls, err := parser.ParseArgs(args)
	if err != nil {
		console.Log(err)
		return "", err.Error()
	}
	if len(urls) > 1 {
		return "", "It accept only one url. Remained urls are ignored:" + strings.Join(urls, ", ")
	}
	if parser.Active == curlCommand {
		// --url option has higher priority than params.
		if curlOptions.Url == "" {
			if len(urls) > 0 {
				curlOptions.Url = urls[0]
			} else {
				console.Error("Both --url option and url parameters are missing")
				return "", "Both --url option and url parameters are missing"
			}
		}
		sourceCode, _, _, _ := httpgen_generator.GenerateCode(target, &curlOptions)
		return html.EscapeString(sourceCode), ""
	}
	return "", ""
}

func main() {
	js.Global.Set("CurlAsDsl", map[string]interface{}{
		"generate": GenerateCode,
	})
}
