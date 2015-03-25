package main

import (
	"fmt"
	"github.com/shibukawa/curl_as_dsl/form2curl"
	"os"
)

func Usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  form2curl [OPTIONS] [input html file]

Options
  -s, --source=SRC    Source form text
  -h, --help          Show this help message	`)
	os.Exit(1)
}

func main() {
	var form *form2curl.Form
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h":
			Usage()
		case "--help":
			Usage()
		case "-s":
			if len(os.Args) > 2 {
				var err error
				form, err = form2curl.CreateFormFromString(os.Args[2])
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			} else {
				fmt.Fprintln(os.Stderr, "-s needs additional content as form source.")
			}
		case "--source":
			if len(os.Args) > 2 {
				var err error
				form, err = form2curl.CreateFormFromString(os.Args[2])
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			} else {
				fmt.Fprintln(os.Stderr, "--source needs additional content as form source.")
			}
		default:
			// input from file
			file, err := os.Open(os.Args[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			form, err = form2curl.CreateFormFromReader(file)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	} else {
		// input from stdin
		var err error
		form, err = form2curl.CreateFormFromReader(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if form != nil {
		if len(form.Warnings) > 0 {
			for _, warning := range form.Warnings {
				fmt.Fprintln(os.Stderr, warning)
			}
		} else {
			fmt.Println(form.MakeCurlCommand())
		}
	}
}
