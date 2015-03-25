package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/shibukawa/curl_as_dsl/common"
	"github.com/shibukawa/curl_as_dsl/generator"
	"log"
	"os"
	"reflect"
)

type GlobalOptions struct {
	Target string `short:"t" long:"target" value-name:"NAME" description:"Target name of code generator" default:"go"`
	Debug  bool   `short:"d" long:"debug" description:"Debug option"`
}

func PrintLangHelp(target string) {
	fmt.Fprintf(os.Stderr, `
'%s' is not supported as a target.
This program supports one of the following targets:

* go, golang         : Golang      (net/http)
* py, python         : Python 3    (http.client)
* node, js.node      : node.js     (http.request)
* xhr, js.xhr        : Browser     (XMLHttpRequest)
* java               : Java        (java.net.HttpURLConnection)
* objc, objc.session : Objective-C (NSURLSession)
* objc.connection    : Objective-C (NSURLConnection)
* php                : PHP         (fopen)
* vim                : Vim script  (webapi-vim)`, target)
}

func main() {
	var globalOptions GlobalOptions
	var curlOptions common.CurlOptions
	curlOptions.Init()

	parser := flags.NewParser(&globalOptions, flags.Default)
	curlCommand, err := parser.AddCommand("curl",
		"Generate code from curl options",
		"This command has almost same options of curl and generate code",
		&curlOptions)
	urls, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	if len(urls) > 1 {
		fmt.Println("It accept only one url. Remained urls are ignored.")
	}
	if parser.Active == curlCommand {
		// --url option has higher priority than params.
		if curlOptions.Url == "" {
			if len(urls) > 0 {
				curlOptions.Url = urls[0]
			} else {
				log.Fatalln("Both --url option and url parameters are missing")
			}
		}
		sourceCode, langName, templateName, option := generator.GenerateCode(globalOptions.Target, &curlOptions)
		if templateName != "" {
			if globalOptions.Debug {
				st := reflect.TypeOf(option)
				v := reflect.ValueOf(option)
				fmt.Fprintf(os.Stderr, "Debug: template name=%s_%s\n", langName, templateName)
				fmt.Fprintf(os.Stderr, "Debug: template context=%s\n", st.Name())
				num := st.NumField()
				for i := 0; i < num; i++ {
					fmt.Fprintf(os.Stderr, "    %s: %s\n", st.Field(i).Name, v.Field(i).String())
				}
			}
			fmt.Println(sourceCode)
		} else {
			PrintLangHelp(globalOptions.Target)
			os.Exit(1)
		}
	}
}
