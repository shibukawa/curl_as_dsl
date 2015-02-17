package go_client

import (
	"fmt"
	"log"
	"strings"
	"../httpgen_common"
	"mime"
)

func escapeDQ(src string) string {
	return strings.Replace(strings.Replace(src, "\"", "\\\"", -1), "\\", "\\\\", -1)
}

func ClientNeeded(options *httpgen_common.CurlOptions) bool {
	if options.Proxy != "" {
		return true
	}
	if options.OnlyHasContentTypeHeader() {
		method := options.Method()
		if method != "GET" && method != "POST" {
			return true
		}
	}
	return false
}

/*
	Dispatcher function of curl command
	This is an exported function and called from httpgen.
 */
func ProcessCurlCommand(options *httpgen_common.CurlOptions) (string, interface{}) {

	generator := NewGoGenerator(options)

	if ClientNeeded(options) {
		return processCurlFullFeatureRequest(generator)
	}

	method := options.Method()
	onlyHasContentTypeHeader := options.OnlyHasContentTypeHeader()

	if method == "POST" && onlyHasContentTypeHeader {
		if options.Transfer != "" {
			return processCurlPostSingleFile(generator)
		} else {
			if !options.ProcessedData.HasData() {

			}
			if options.ProcessedData.HasData() {
				if options.Get {
					return processCurlPostDataWithUrl(generator)
				} else {
					return processCurlPostData(generator, options.ProcessedData)
				}
			} else if options.ProcessedData.HasForm() {
				return processCurlPostForm(generator)
			} else {
				return processCurlSimple(generator)
			}
		}
	}

	if method == "GET" {
		if len(options.ProcessedData) > 0 {
			return processCurlGetDataWithUrl(generator)
		} else {
			return processCurlSimple(generator)
		}
	}

	if !options.ProcessedData.HasData() && !options.ProcessedData.HasForm() {
		return processCurlSimple(generator)
	}

	return "", nil
}

func processCurlFullFeatureRequest(generator *GoGenerator) (string, interface{}) {
	log.Println("processCurlFullFeatureRequest")
	return "full", *generator
}

func processCurlPostSingleFile(generator *GoGenerator) (string, interface{}) {
	contentType := ""
	headers := generator.Options.Headers()
	if len(headers) > 0 {
		contentType = headers[0][1]
	} else {
		contentType = mime.TypeByExtension(generator.Options.Transfer)
	}
	var value struct {
		Url string
		FilePath string
		ContentType string
	}
	value.Url = fmt.Sprintf("\"%s\"", generator.Options.Url)
	value.FilePath = generator.Options.Transfer
	value.ContentType = contentType
	value.ContentType = contentType
	return "post_single_file", value
}

func processCurlPostForm(generator *GoGenerator) (string, interface{}) {
	if !canUseSimpleForm(&generator.Options.ProcessedData) {
		return processCurlPostData(generator, generator.Options.ProcessedData)
	}
	generator.Modules["net/url"] = true
	generator.SetDataForForm()
	return "post_form", generator
}

func processCurlPostData(generator *GoGenerator, inputs []httpgen_common.DataOption) (string, interface{}) {
	generator.DataVariable = "&buffer"
	var contentType string
	if generator.Options.ProcessedData.HasForm() {
		contentType = "multipart/form-data"
		generator.SetFormForBody()
	} else {
		contentType = "application/x-www-form-urlencoded"
		generator.SetDataForBody()
	}
	headers := generator.Options.Headers()
	if len(headers) > 0 {
		contentType = headers[0][1]
	}

	generator.ContentType = contentType
	return "post_text", *generator
}

func processCurlPostDataWithUrl(generator *GoGenerator) (string, interface{}) {
	generator.SetDataForUrl()
	return "post_with_data_url", *generator
}

func processCurlGetDataWithUrl(generator *GoGenerator) (string, interface{}) {
	generator.SetDataForUrl()
	return "get_with_data_url", *generator
}

func processCurlSimple(generator *GoGenerator) (string, interface{}) {
	method := generator.Options.Method()
	if method == "GET" {
		return "simple_get", *generator
	} else { // "POST"
		return "simple_post", *generator
	}
}

func canUseSimpleForm(dataOptions *httpgen_common.DataOptions) bool {
	for _, form := range *dataOptions {
		if form.UseExternalFile() {
			return false
		}
		if strings.Index(form.Value, "=") == -1 {
			return false
		}
	}
	return true
}