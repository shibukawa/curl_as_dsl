package main

import (
	"fmt"
	"honnef.co/go/js/console"
    "github.com/gopherjs/gopherjs/js"
	"github.com/jessevdk/go-flags"
	"github.com/shibukawa/optstring_parser"
	"github.com/shibukawa/curl_as_dsl/httpgen_common"
	"github.com/shibukawa/curl_as_dsl/httpgen_generator"
)

type GlobalOptions struct {}

func GenerateCode(target, options string) (string, string) {
	var globalOptions GlobalOptions
	var curlOptions httpgen_common.CurlOptions
	curlOptions.Init()

	parser := flags.NewParser(&globalOptions, flags.Default)
	curlCommand, err := parser.AddCommand("curl",
		"Generate code from curl options",
		"This command has almost same options of curl and generate code",
	&curlOptions)


	commandLine := fmt.Sprintf("httpgen -t %s curl %s", target, options)
	args := optstring_parser.Parse(commandLine)

	urls, err := parser.ParseArgs(args)
	if err != nil {
		console.Log(err)
		return "", err.Error()
	}
	if len(urls) > 1 {
		return "", "It accept only one url. Remained urls are ignored."
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
		return sourceCode, ""
	}
	return "", ""
}

func main() {
    js.Global.Set("CurlAsDsl", map[string]interface{}{
        "Generate": GenerateCode,
    })
}
