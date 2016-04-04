package golang

import (
	"fmt"
	"github.com/shibukawa/curl_as_dsl/common"
	"mime"
	"strings"
)

func escapeDQ(src string) string {
	return strings.Replace(strings.Replace(src, "\"", "\\\"", -1), "\\", "\\\\", -1)
}

func ClientNeeded(options *common.CurlOptions) bool {
	if options.Insecure || options.Proxy != "" || options.User != "" || len(options.Cookie) > 0 {
		return true
	}
	if len(options.AWSV2) > 0 {
		return true
	}
	if options.OnlyHasContentTypeHeader() {
		method := options.Method()
		if method != "GET" && method != "POST" {
			return true
		}
		return false
	}
	return true
}

/*
	Dispatcher function of curl command
	This is an exported function and called from httpgen.
*/
func ProcessCurlCommand(options *common.CurlOptions) (string, interface{}) {

	generator := NewGoGenerator(options)

	if ClientNeeded(options) {
		return processCurlFullFeatureRequest(generator)
	}

	method := options.Method()
	onlyHasContentTypeHeader := options.OnlyHasContentTypeHeader()

	if method == "POST" && onlyHasContentTypeHeader {
		if options.ProcessedData.HasData() {
			if options.Get {
				return processCurlPostDataWithUrl(generator)
			} else if len(options.ProcessedData) == 1 && options.ProcessedData[0].UseExternalFile() {
				return processCurlPostSingleFile(generator)
			} else {
				return processCurlPostData(generator)
			}
		} else if options.ProcessedData.HasForm() {
			return processCurlPostData(generator)
		} else {
			return processCurlSimple(generator)
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
	options := generator.Options

	if options.ProcessedData.HasData() {
		if options.Get {
			generator.SetDataForUrl()
			generator.DataVariable = "nil"
		} else {
			generator.DataVariable = "&buffer"
			generator.Options.InsertContentTypeHeader("application/x-www-form-urlencoded")
			generator.SetDataForBody()
		}
	} else if options.ProcessedData.HasForm() {
		generator.DataVariable = "&buffer"
		generator.Options.InsertContentTypeHeader("multipart/form-data")
		generator.HasBoundary = true

		generator.SetFormForBody()
	}
	if options.Proxy != "" {
		generator.Modules["net/url"] = true
	}
	if options.User != "" {
		generator.Modules["encoding/base64"] = true
	}
	if options.Insecure {
		generator.Modules["crypto/tls"] = true
	}
	if generator.Options.AWSV2 != "" {
		generator.Modules["encoding/base64"] = true
		generator.Modules["crypto/hmac"] = true
		generator.Modules["crypto/sha1"] = true
		generator.Modules["time"] = true
		generator.Modules["fmt"] = true
	}
	return "full", *generator
}

func processCurlPostSingleFile(generator *GoGenerator) (string, interface{}) {
	fileName := generator.Options.ProcessedData[0].Value[1:]
	contentType := ""
	headers := generator.Options.Headers()
	if len(headers) > 0 {
		contentType = headers[0][1]
	} else {
		contentType = mime.TypeByExtension(fileName)
	}
	var value struct {
		Url         string
		FilePath    string
		ContentType string
	}
	value.Url = fmt.Sprintf("\"%s\"", generator.Options.Url)
	value.FilePath = fileName
	value.ContentType = contentType
	return "post_single_file", value
}

func processCurlPostData(generator *GoGenerator) (string, interface{}) {
	var contentType string
	headers := generator.Options.Headers()
	if len(headers) > 0 {
		contentType = headers[0][1]
	}
	if !generator.Options.ProcessedData.HasForm() && generator.Options.CanUseSimpleForm() && contentType == "" {
		generator.Modules["net/url"] = true
		generator.SetDataForPostForm()
		return "post_form", generator
	}
	if generator.Options.ProcessedData.HasForm() {
		generator.DataVariable = "&buffer"
		contentType = "multipart/form-data"
		generator.SetFormForBody()
		generator.HasBoundary = true
	} else {
		generator.DataVariable = "&buffer"
		contentType = "application/x-www-form-urlencoded"
		generator.SetDataForBody()
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
