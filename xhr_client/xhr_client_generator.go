package xhr_client

import (
	"bytes"
	"fmt"
	"github.com/shibukawa/curl_as_dsl/httpgen_common"
	"net/url"
	"os"
	"strings"
)

func escapeDQ(src string) string {
	return strings.Replace(strings.Replace(src, "\"", "\\\"", -1), "\\", "\\\\", -1)
}

type ExternalFile struct {
	Data         *httpgen_common.DataOption
	FileName     string
	VariableName string
	TextType     bool
}

type XHRGenerator struct {
	Options               *httpgen_common.CurlOptions
	prepareFile           bytes.Buffer
	PrepareBody           string
	Body                  string
	HasBody               bool
	ExternalFiles         map[int]*ExternalFile
	usedFile              int
	extraUrl              string
	AdditionalDeclaration string
	processedHeaders      []httpgen_common.HeaderGroup
	specialHeaders        [][]string

	UseSimpleGet bool
}

func NewXHRGenerator(options *httpgen_common.CurlOptions) *XHRGenerator {
	result := &XHRGenerator{
		Options:       options,
		ExternalFiles: make(map[int]*ExternalFile),
	}

	return result
}

//--- Getter methods called from template

func (self XHRGenerator) Url() string {
	if self.extraUrl != "" {
		return fmt.Sprintf(`"%s" + "?" + %s`, self.Options.Url, self.extraUrl)
	}
	return fmt.Sprintf(`"%s"`, self.Options.Url)
}

func (self XHRGenerator) Method() string {
	return self.Options.Method()
}

func (self XHRGenerator) PrepareOptions() string {
	var buffer bytes.Buffer
	if len(self.processedHeaders) != 0 || len(self.specialHeaders) != 0 {
		for _, header := range self.processedHeaders {
			for i, value := range header.Values {
				if i != 0 {
					buffer.WriteString("    ")
				}
				fmt.Fprintf(&buffer, "xhr.setRequestHeader(\"%s\", \"%s\")\n", header.Key, value)
			}
		}
		for _, headers := range self.specialHeaders {
			fmt.Fprintf(&buffer, "xhr.setRequestHeader(\"%s\", %s)\n", headers[0], headers[1])
		}
	}
	return buffer.String()
}

func (self XHRGenerator) FileNames() []string {
	var result []string
	for _, file := range self.ExternalFiles {
		result = append(result, file.VariableName)
	}
	return result
}

func (self XHRGenerator) PrepareFile() string {
	return self.prepareFile.String()
}

//--- Setter/Getter methods

func (self *XHRGenerator) FileReader() string {
	if len(self.ExternalFiles) == 1 {
		return "fileReader"
	}
	index := self.usedFile
	self.usedFile = self.usedFile + 1
	return fmt.Sprintf("fileReader%d", index)
}

func (self *XHRGenerator) SetDataForUrl() {
	if self.Options.CanUseSimpleForm() {
		self.SetDataForForm(false)
		self.extraUrl = self.Body
		self.Body = ""
		self.HasBody = false
	} else {
		// Use bytes.Buffer to create URL option string
		self.SetDataForBody()
		self.extraUrl = self.Body
		self.Body = ""
		self.HasBody = false
	}
}

func (self *XHRGenerator) VariableName(data *httpgen_common.DataOption) string {
	for _, file := range self.ExternalFiles {
		if file.Data == data {
			return file.VariableName
		}
	}
	return ""
}

func (self *XHRGenerator) SetDataForBody() {
	var prepareFile string
	if len(self.Options.ProcessedData) == 1 {
		self.Body, prepareFile = NewStringForData(self, &self.Options.ProcessedData[0])
		self.prepareFile.WriteString(prepareFile)
	} else {
		var buffer bytes.Buffer
		buffer.WriteString("\n    var content = [\n")
		for _, data := range self.Options.ProcessedData {
			dataStr, _ := StringForData(self, &data)
			fmt.Fprintf(&buffer, "        %s,\n", dataStr)
		}
		buffer.WriteString("    ];\n")
		self.PrepareBody = buffer.String()
		self.Body = `content.join("&")`
	}
	self.HasBody = true
}

func (self *XHRGenerator) SetDataForForm(hasIndent bool) {
	var buffer bytes.Buffer
	buffer.WriteString("\n    var query = [\n")
	for _, data := range self.Options.ProcessedData {
		singleData, _ := url.ParseQuery(data.Value)
		for key, values := range singleData {
			for _, value := range values {
				buffer.WriteString(fmt.Sprintf("        encodeURIComponent(\"%s\") + \"=\" + encodeURIComponent(\"%s\"),\n", escapeDQ(key), escapeDQ(value)))
			}
		}
	}
	buffer.WriteString(fmt.Sprintf("    ];\n"))

	self.PrepareBody = buffer.String()
	self.HasBody = true
	self.Body = `query.join("&")`
}

func (self *XHRGenerator) SetFormForBody() {
	var buffer bytes.Buffer
	buffer.WriteString("\n    var form = new FormData();\n")

	for _, data := range self.Options.ProcessedData {
		body, prepareFile := FormString(self, &data)
		buffer.WriteString(body)
		self.prepareFile.WriteString(prepareFile)
	}

	self.PrepareBody = buffer.String()
	self.Body = "form"
	self.HasBody = true
}

/*
	Dispatcher function of curl command
	This is an exported function and called from httpgen.
*/
func ProcessCurlCommand(options *httpgen_common.CurlOptions) (string, interface{}) {
	generator := NewXHRGenerator(options)

	for i, data := range options.ProcessedData {
		fileName := data.FileName()
		if fileName != "" {
			isText := data.Type == httpgen_common.DataAsciiType
			file := &ExternalFile{
				Data:     &data,
				FileName: fileName,
				TextType: isText,
			}
			generator.ExternalFiles[i] = file
		}
	}

	if len(generator.ExternalFiles) == 1 {
		generator.ExternalFiles[0].VariableName = "file"
	} else {
		for i, externalFile := range generator.ExternalFiles {
			externalFile.VariableName = fmt.Sprintf("file_%d", i+1)
		}
	}

	generator.processedHeaders = options.GroupedHeaders()

	var templateName string
	switch len(generator.ExternalFiles) {
	case 0:
		templateName = "simple"
	case 1:
		templateName = "external_file"
	default:
		templateName = "external_files"
	}

	if generator.Options.User != "" {
		generator.specialHeaders = append(generator.specialHeaders, []string{"Authorization", fmt.Sprintf(`"Basic " + btoa("%s")`, generator.Options.User)})
	}

	if options.ProcessedData.HasData() {
		if options.Get {
			generator.SetDataForUrl()
		} else {
			generator.Options.InsertContentTypeHeader("application/x-www-form-urlencoded")
			generator.SetDataForBody()
		}
	} else if options.ProcessedData.HasForm() {
		generator.SetFormForBody()
	} else if options.Method() == "GET" && len(generator.processedHeaders) == 0 && len(generator.specialHeaders) == 0 {
	}

	return templateName, *generator
}

// helper functions

func NewStringForData(generator *XHRGenerator, data *httpgen_common.DataOption) (string, string) {
	var result string
	var prepare bytes.Buffer
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = "file"
			prepare.WriteString(`
    var reader = new FileReader();
    reader.onloadend = function(evt) {
        if (evt.target.readyState == FileReader.DONE) {
            request(evt.target.result.replace(/\n/g, ""));
        }
    };
    reader.readAsText(file, "UTF-8");`)
		} else {
			result = fmt.Sprintf("\"%s\"", escapeDQ(data.Value))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = "file"
			prepare.WriteString(`
    var reader = new FileReader();
    reader.onloadend = function(evt) {
        if (evt.target.readyState == FileReader.DONE) {
            request(evt.target.result);
        }
    };
    reader.readAsText(file, "UTF-8");`)
		} else {
			result = fmt.Sprintf("\"%s\"", escapeDQ(data.Value))
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = "file"
			prepare.WriteString(`
    var reader = new FileReader();
    reader.onloadend = function(evt) {
        if (evt.target.readyState == FileReader.DONE) {
            request(encodeURIComponent(evt.target.result));
        }
    };
    reader.readAsText(file, "UTF-8");`)
		} else {
			result = fmt.Sprintf(`encodeURIComponent("%s")`, escapeDQ(data.Value))
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, prepare.String()
}

func StringForData(generator *XHRGenerator, data *httpgen_common.DataOption) (string, string) {
	var result string
	var prepare bytes.Buffer
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			fmt.Fprintf(os.Stderr, "XHR generator doesn't support sending multiple files except form(-F).")
			os.Exit(1)
		} else {
			result = fmt.Sprintf("\"%s\"", escapeDQ(data.Value))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			fmt.Fprintf(os.Stderr, "XHR generator doesn't support sending multiple files except form(-F).")
			os.Exit(1)
		} else {
			result = fmt.Sprintf("\"%s\"", escapeDQ(data.Value))
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("encodeURIComponent(%s)", "@@")
		} else {
			result = fmt.Sprintf("encodeURIComponent(\"%s\")", escapeDQ(data.Value))
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, prepare.String()
}

func FormString(generator *XHRGenerator, data *httpgen_common.DataOption) (string, string) {
	var buffer bytes.Buffer
	var prepare bytes.Buffer
	switch data.Type {
	case httpgen_common.FormType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
		if strings.HasPrefix(field[1], "@") {
			fragments := strings.Split(field[1][1:], ";")

			sentFileName := ""
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "filename=") {
					sentFileName = fragment[9:]
				} else if strings.HasPrefix(fragment, "type=") {
					fmt.Fprintf(os.Stderr, "XHR doesn't support sending content-type.")
					os.Exit(1)
				}
			}
			if len(generator.ExternalFiles) == 1 {
				if sentFileName != "" {
					fmt.Fprintf(&buffer, "    form.append(\"%s\", file, \"%s\");\n", field[0], sentFileName)
				} else {
					fmt.Fprintf(&buffer, "    form.append(\"%s\", file);\n", field[0])
				}
				prepare.WriteString("request(file);")
			} else {
				if sentFileName != "" {
					fmt.Fprintf(&buffer, "    form.append(\"%s\", files.%s, \"%s\");\n", field[0], generator.VariableName(data), sentFileName)
				} else {
					fmt.Fprintf(&buffer, "    form.append(\"%s\", files.%s);\n", field[0], generator.VariableName(data))
				}
				fmt.Fprintf(&buffer, `files.%s = file;
    request();`, generator.VariableName(data))
			}
			// sent file name
			// field name, source file name
		} else if strings.HasPrefix(field[1], "<") {
			if len(generator.ExternalFiles) > 1 {
				fmt.Fprintf(os.Stderr, "XHR generator doesn't support sending multiple files except multipart form.")
				os.Exit(1)
			}
			fragments := strings.Split(field[1][1:], ";")
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "type=") {
					fmt.Fprintf(os.Stderr, "XHR doesn't support sending content-type.")
					os.Exit(1)
				}
			}
			fmt.Fprintf(&buffer, "    form.append(\"%s\", file);\n", field[0])
			prepare.WriteString(`
    var reader = new FileReader();
    reader.onloadend = function(evt) {
        if (evt.target.readyState == FileReader.DONE) {
            request(evt.target.result);
        }
    };
    reader.readAsText(file, "UTF-8");`)
		} else {
			fmt.Fprintf(&buffer, "    form.append(\"%s\", \"%s\");\n", field[0], field[1])
		}
	case httpgen_common.FormStringType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
	}
	return buffer.String(), prepare.String()
}
