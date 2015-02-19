package main

import (
	"./go_client"
	"./httpgen_common"
	"bytes"
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"reflect"
	"text/template"
)

type GlobalOptions struct {
	Target string `short:"t" long:"target" value-name:"NAME" description:"Target name of code generator" default:"go_client"`
	Debug  bool   `short:"d" long:"debug" description:"Debug option"`
}

func render(lang, key string, options interface{}) string {
	src, _ := Asset(fmt.Sprintf("templates/%s_%s.tpl", lang, key))
	tpl := template.Must(template.New(key).Parse(string(src)))
	var buffer bytes.Buffer
	err := tpl.Execute(&buffer, options)
	if err != nil {
		log.Fatal(err)
	}
	return buffer.String()
}

func main() {
	var globalOptions GlobalOptions
	var curlOptions httpgen_common.CurlOptions
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
		var langName string
		var templateName string
		var option interface{}

		switch globalOptions.Target {
		case "go_client":
			langName = "go"
			templateName, option = go_client.ProcessCurlCommand(&curlOptions)
		}
		if templateName != "" {
			if globalOptions.Debug {
				st := reflect.TypeOf(option)
				v := reflect.ValueOf(option)
				fmt.Fprintf(os.Stderr, "Debug: template name=%s\n", templateName)
				fmt.Fprintf(os.Stderr, "Debug: template context=%s\n", st.Name())
				num := st.NumField()
				for i := 0; i < num; i++ {
					fmt.Fprintf(os.Stderr, "    %s: %s\n", st.Field(i).Name, v.Field(i).String())
				}
			}
			fmt.Println(render(langName, templateName, option))
		}
	}
}
