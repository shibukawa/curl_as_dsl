package nodejs_client

import (
	"bytes"
	"fmt"
	"github.com/shibukawa/curl_as_dsl/httpgen_common"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func escapeDQ(src string) string {
	return strings.Replace(strings.Replace(src, "\"", "\\\"", -1), "\\", "\\\\", -1)
}

type ExternalFile struct {
	FileName string
	TextType bool
}

type NodeJsGenerator struct {
	Options *httpgen_common.CurlOptions
	Modules map[string]bool

	ClientModule          string
	PrepareBody           string
	HasBody               bool
	BodyLines             []string
	ExternalFiles         []ExternalFile
	usedFile              int
	extraUrl              string
	AdditionalDeclaration string
	processedHeaders      []httpgen_common.HeaderGroup
	specialHeaders        []string

	UseSimpleGet bool
}

func NewNodeJsGenerator(options *httpgen_common.CurlOptions) *NodeJsGenerator {
	result := &NodeJsGenerator{Options: options}
	result.Modules = make(map[string]bool)

	return result
}

//--- Getter methods called from template

func (self NodeJsGenerator) Url() string {
	return fmt.Sprintf("\"%s\"", self.Options.Url)
}

func (self NodeJsGenerator) Host() string {
	u, err := url.Parse(self.Options.Url)
	if err != nil {
		log.Fatal(err)
	}
	fragments := strings.SplitN(u.Host, ":", 2)
	if len(fragments) > 1 {
		_, err = strconv.Atoi(fragments[1])
		if err != nil {
			return u.Host
		}
	}
	return fragments[0]
}

func (self NodeJsGenerator) Port() int {
	u, err := url.Parse(self.Options.Url)
	if err != nil {
		log.Fatal(err)
	}
	fragments := strings.SplitN(u.Host, ":", 2)
	if len(fragments) > 1 {
		port, err := strconv.Atoi(fragments[1])
		if err != nil {
			return 0
		}
		return port
	}
	if u.Scheme == "http" {
		return 80
	}
	return 443
}

func (self NodeJsGenerator) Method() string {
	return self.Options.Method()
}

func (self NodeJsGenerator) Path() string {
	u, err := url.Parse(self.Options.Url)
	if err != nil {
		log.Fatal(err)
	}
	path := u.Path
	if u.Path == "" {
		path = "/"
	}
	if self.extraUrl != "" {
		return fmt.Sprintf("\"%s?\" + %s", path, self.extraUrl)
	} else {
		return fmt.Sprintf("\"%s\"", path)
	}
}

func (self NodeJsGenerator) PrepareOptions() string {
	hasIndent := self.Options.ProcessedData.ExternalFileCount() > 0
	indent := func() string {
		if hasIndent {
			return "    "
		} else {
			return ""
		}
	}
	var buffer bytes.Buffer
	if len(self.processedHeaders) != 0 || len(self.specialHeaders) != 0 {
		buffer.WriteString(fmt.Sprintf("\n%s    headers: {\n", indent()))
		for _, header := range self.processedHeaders {
			if len(header.Values) == 1 {
				buffer.WriteString(fmt.Sprintf("%s        \"%s\": \"%s\",\n", indent(), header.Key, header.Values[0]))
			} else {
				buffer.WriteString(fmt.Sprintf("%s        \"%s\": [", indent(), header.Key))
				for i, value := range header.Values {
					if i != 0 {
						buffer.WriteString(", ")
					}
					buffer.WriteString(fmt.Sprintf("\"%s\"", value))
				}
				buffer.WriteString("],\n")
			}
		}
		for _, header := range self.specialHeaders {
			buffer.WriteString(fmt.Sprintf("%s        %s,\n", indent(), header))
		}
		buffer.WriteString(fmt.Sprintf("%s    },", indent()))
	}
	return buffer.String()
}

//--- Setter/Getter methods

func (self *NodeJsGenerator) AddMultiPartCode() {
	self.AdditionalDeclaration = `
BOUNDARY = '----------ThIs_Is_tHe_bouNdaRY_$';

function encodeMultiPartFormData(fields, files) {
    CRLF = "\r\n";
    L = [];
    for (var i = 0; i < fields.length; i++) {
        var field = fields[i];
        L.push('--' + BOUNDARY);
        L.push('Content-Disposition: form-data; name="' + field.key + '"');
        if (field.contentType) {
            L.push('Content-Type: ' + field.contentType);
        }
        L.push('');
        L.push(field.value);
    }
    for (var i = 0; i < files.length; i++) {
    	var file = files[i];
        L.push('--' + BOUNDARY);
        L.push('Content-Disposition: form-data; name="' + file.key + '"; filename="' + file.filename + '"');
        L.push('Content-Type: ' + file.contentType);
        L.push('');
        L.push(file.content);
        L.push('--' + BOUNDARY + '--');
        L.push('');
    }
    return L.join("\r\n");
}
`
	boundary := "----------ThIs_Is_tHe_bouNdaRY_$"
	self.Options.InsertContentTypeHeader(fmt.Sprintf("multipart/form-data; boundary=%s", boundary))
}

func (self *NodeJsGenerator) FileContent() string {
	if len(self.ExternalFiles) == 1 {
		return "fileContent"
	}
	index := self.usedFile
	self.usedFile = self.usedFile + 1
	return fmt.Sprintf("fileContents[%d]", index)
}

func (self *NodeJsGenerator) SetDataForUrl() {
	if self.Options.CanUseSimpleForm() {
		self.SetDataForForm(false)
		self.extraUrl = strings.Join(self.BodyLines, "")
		self.BodyLines = nil
		self.HasBody = false
	} else {
		// Use bytes.Buffer to create URL option string
		self.SetDataForBody()
		self.extraUrl = strings.Join(self.BodyLines, "")
		self.BodyLines = nil
		self.HasBody = false
	}
}

func (self *NodeJsGenerator) SetDataForBody() {
	if len(self.Options.ProcessedData) == 1 {
		self.BodyLines = append(self.BodyLines, NewStringForData(self, &self.Options.ProcessedData[0]))
	} else {
		for i, data := range self.Options.ProcessedData {
			if i != 0 {
				self.BodyLines = append(self.BodyLines, "\"&\"")
			}
			self.BodyLines = append(self.BodyLines, StringForData(self, &data))
		}
	}
	self.HasBody = true
}

func (self *NodeJsGenerator) SetDataForForm(hasIndent bool) {
	entries := make(map[string][]string)
	indent := func() string {
		if hasIndent {
			return "    "
		}
		return ""
	}

	for _, data := range self.Options.ProcessedData {
		singleData, _ := url.ParseQuery(data.Value)
		for key, values := range singleData {
			entries[key] = append(entries[key], values[0])
		}
	}

	var buffer bytes.Buffer
	count := 0
	for key, values := range entries {
		if count == 0 {
			buffer.WriteString("var query = querystring.stringify({\n")
		} else {
			buffer.WriteString(fmt.Sprintf("%s, \"", indent()))
		}
		if len(values) == 1 {
			buffer.WriteString(fmt.Sprintf("%s    \"%s\": \"%s\",\n", indent(), escapeDQ(key), escapeDQ(values[0])))
		} else {
			buffer.WriteString(fmt.Sprintf("%s    \"%s\": [\n%s         ", indent(), key, indent()))
			for i, value := range values {
				if i != 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(fmt.Sprintf("\"%s\"", escapeDQ(value)))
			}
			buffer.WriteString(fmt.Sprintf("],\n%s    ", indent()))
		}
		count++
	}
	buffer.WriteString(fmt.Sprintf("});%s\n", indent()))

	self.PrepareBody = buffer.String()
	self.HasBody = true
	self.BodyLines = append(self.BodyLines, "query")
	self.Modules["querystring"] = true
}

func (self *NodeJsGenerator) SetFormForBody() {
	self.AddMultiPartCode()
	var fields []string
	var files []string

	for _, data := range self.Options.ProcessedData {
		if data.SendAsFormFile() {
			files = append(files, FormString(self, &data))
		} else {
			fields = append(fields, FormString(self, &data))
		}
	}

	var buffer bytes.Buffer
	if len(fields) > 0 {
		buffer.WriteString("fields = [\n")
		for _, value := range fields {
			buffer.WriteString(value)
		}
		buffer.WriteString("];\n")
	}
	if len(files) > 0 {
		if len(fields) > 0 {
			buffer.WriteString("    ")
			self.BodyLines = append(self.BodyLines, "encodeMultiPartFormData(fields, files)")
		} else {
			self.BodyLines = append(self.BodyLines, "encodeMultiPartFormData([], files)")
		}
		buffer.WriteString("files = [\n")
		for _, value := range files {
			buffer.WriteString(value)
		}
		buffer.WriteString("];\n")
	} else {
		self.BodyLines = append(self.BodyLines, "encodeMultiPartFormData(fields, [])")
	}
	self.PrepareBody = buffer.String()
	self.HasBody = true
}

/*
	Dispatcher function of curl command
	This is an exported function and called from httpgen.
*/
func ProcessCurlCommand(options *httpgen_common.CurlOptions) (string, interface{}) {
	generator := NewNodeJsGenerator(options)
	if strings.HasPrefix(options.Url, "https") {
		generator.Modules["https"] = true
		generator.ClientModule = "https"
	} else {
		generator.Modules["http"] = true
		generator.ClientModule = "http"
	}

	for _, data := range options.ProcessedData {
		fileName := data.FileName()
		if fileName != "" {
			isText := data.Type == httpgen_common.DataAsciiType
			generator.ExternalFiles = append(generator.ExternalFiles, ExternalFile{FileName: fileName, TextType: isText})
		}
	}

	generator.processedHeaders = options.GroupedHeaders()

	var templateName string
	switch len(generator.ExternalFiles) {
	case 0:
		templateName = "full"
	case 1:
		templateName = "external_file"
	default:
		templateName = "external_files"
	}

	if generator.Options.User != "" {
		generator.specialHeaders = append(generator.specialHeaders, fmt.Sprintf("\"Authorization\": \"Basic \" + new Buffer(\"%s\").toString(\"base64\")", generator.Options.User))
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
		if templateName == "full" {
			templateName = "simple_get"
		}
	}

	return templateName, *generator
}

// helper functions

func NewStringForData(generator *NodeJsGenerator, data *httpgen_common.DataOption) string {
	var result string
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("%s.replace('\\n', '')", generator.FileContent())
		} else {
			result = fmt.Sprintf("\"%s\"", escapeDQ(strings.Replace(data.Value, "\n", "", -1)))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = generator.FileContent()
		} else {
			result = fmt.Sprintf("\"%s\"", escapeDQ(data.Value))
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("encodeURIComponent(%s)", generator.FileContent())
		} else {
			result = fmt.Sprintf("encodeURIComponent(\"%s\")", escapeDQ(data.Value))
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result
}

func StringForData(generator *NodeJsGenerator, data *httpgen_common.DataOption) string {
	var result string
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("%s.replace('\\n', '')", generator.FileContent())
		} else {
			result = fmt.Sprintf("\"%s\"", escapeDQ(strings.Replace(data.Value, "\n", "", -1)))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("%s", generator.FileContent())
		} else {
			result = fmt.Sprintf("\"%s\"", escapeDQ(data.Value))
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("encodeURIComponent(%s)", generator.FileContent())
		} else {
			result = fmt.Sprintf("encodeURIComponent(\"%s\")", escapeDQ(data.Value))
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result
}

func FormString(generator *NodeJsGenerator, data *httpgen_common.DataOption) string {
	var result string
	switch data.Type {
	case httpgen_common.FormType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
		if strings.HasPrefix(field[1], "@") {
			var buffer bytes.Buffer
			fragments := strings.Split(field[1][1:], ";")

			contentType := "application/octet-stream"
			sentFileName := fragments[0]
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "filename=") {
					sentFileName = fragment[9:]
				} else if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
				}
			}
			// sent file name
			// field name, source file name
			buffer.WriteString(fmt.Sprintf("        {key: \"%s\", filename: \"%s\", content: %s, contentType: \"%s\"},\n", field[0],
				sentFileName, generator.FileContent(), contentType))
			result = buffer.String()
		} else if strings.HasPrefix(field[1], "<") {
			var buffer bytes.Buffer

			fragments := strings.Split(field[1][1:], ";")

			// field name, content
			buffer.WriteString(fmt.Sprintf("        {key:\"%s\", value: %s", field[0], generator.FileContent()))

			var contentType string
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
					break
				}
			}
			if contentType != "" {
				buffer.WriteString(fmt.Sprintf(", contentType: \"%s\"},\n", contentType))
			} else {
				buffer.WriteString("},\n")
			}
			result = buffer.String()
		} else {
			result = fmt.Sprintf("    {key: \"%s\", value: \"%s\"},\n", field[0], field[1])
		}
	case httpgen_common.FormStringType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
		result = fmt.Sprintf("    {key: \"%s\", value: \"%s\"},\n", field[0], field[1])
	}
	return result
}
