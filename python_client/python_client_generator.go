package python_client

import (
	"../httpgen_common"
	"fmt"
	"log"
	//"mime"
	"bytes"
	"net/url"
	"os"
	"strings"
)

func escapeDQ(src string) string {
	return strings.Replace(strings.Replace(src, "\"", "\\\"", -1), "\\", "\\\\", -1)
}

type PythonGenerator struct {
	Options *httpgen_common.CurlOptions
	Modules map[string]bool

	HasHeader     bool
	HasBody       bool
	Header        string
	Body          string
	PrepareHeader string
	PrepareBody   string
	extraUrl      string
}

func NewPythonGenerator(options *httpgen_common.CurlOptions) *PythonGenerator {
	result := &PythonGenerator{Options: options}
	result.Modules = make(map[string]bool)
	result.Modules["http.client"] = true

	return result
}

//--- Getter methods called from template

func (self PythonGenerator) ConnectionClass() string {
	var targetUrl string
	if self.Options.Proxy != "" {
		targetUrl = self.Options.Proxy
	} else {
		targetUrl = self.Options.Url
	}
	if strings.HasPrefix(targetUrl, "https") {
		return "HTTPSConnection"
	}
	return "HTTPConnection"
}

func (self PythonGenerator) Host() string {
	var targetUrl string
	if self.Options.Proxy != "" {
		targetUrl = self.Options.Proxy
	} else {
		targetUrl = self.Options.Url
	}
	u, err := url.Parse(targetUrl)
	if err != nil {
		log.Fatal(err)
	}
	return u.Host
}

func (self PythonGenerator) Proxy() string {
	if self.Options.Proxy != "" {
		u, err := url.Parse(self.Options.Url)
		if err != nil {
			log.Fatal(err)
		}
		return fmt.Sprintf("conn.set_tunnel(\"%s\")\n    ", u.Host)
	}
	return ""
}

func (self PythonGenerator) Method() string {
	return self.Options.Method()
}

func (self PythonGenerator) Path() string {
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

func (self PythonGenerator) AdditionalDeclaration() string {
	return ""
}

//--- Setter/Getter methods

func (self *PythonGenerator) SetDataForUrl() {
	if self.Options.CanUseSimpleForm() {
		self.SetDataForForm()
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

func (self *PythonGenerator) SetDataForBody() {
	var buffer bytes.Buffer
	if len(self.Options.ProcessedData) == 1 {
		var body string
		body, self.Body = NewStringForData(self, &self.Options.ProcessedData[0])
		buffer.WriteString(body)
	} else {
		for i, data := range self.Options.ProcessedData {
			if i == 0 {
				buffer.WriteString("body = [\n")
			}
			buffer.WriteString(StringForData(self, &data))
		}
		buffer.WriteString("    ]\n")
		self.Body = "'&'.join(body)"
	}
	self.PrepareBody = buffer.String()
	self.HasBody = true
}

func (self *PythonGenerator) SetDataForForm() {
	entries := make(map[string][]string)
	for _, data := range self.Options.ProcessedData {
		singleData, _ := url.ParseQuery(data.Value)
		for key, values := range singleData {
			entry, ok := entries[key]
			if !ok {
				entry = make([]string, 0)
			}
			entries[key] = append(entry, values[0])
		}
	}

	var buffer bytes.Buffer
	count := 0
	for key, values := range entries {
		if count == 0 {
			buffer.WriteString("values = urllib.parse.urlencode({\n")
		} else {
			buffer.WriteString(", \"")
		}
		buffer.WriteString(fmt.Sprintf("        \"%s\": \"%s\",\n", key, values[0]))
		count++
	}
	buffer.WriteString("    })\n")

	self.PrepareBody = buffer.String()
	self.HasBody = true
	self.Body = "values"
	self.Modules["urllib.parse"] = true
}

func (self *PythonGenerator) SetFormForBody() {
	var buffer bytes.Buffer
	buffer.WriteString("var buffer bytes.Buffer\n")
	buffer.WriteString("    writer := multipart.NewWriter(&buffer)\n")
	for _, data := range self.Options.ProcessedData {
		buffer.WriteString(FormString(self, &data))
	}
	self.PrepareBody = buffer.String()
	self.HasBody = true
}

/*
	Dispatcher function of curl command
	This is an exported function and called from httpgen.
*/
func ProcessCurlCommand(options *httpgen_common.CurlOptions) (string, interface{}) {
	generator := NewPythonGenerator(options)

	if options.ProcessedData.HasData() {
		if options.Get {
			generator.SetDataForUrl()
		} else {
			generator.Options.InsertContentTypeHeader("application/x-www-form-urlencoded")
			generator.SetDataForBody()
		}
	} else if options.ProcessedData.HasForm() {
		generator.Options.InsertContentTypeHeader("multipart/form-data")
		generator.SetFormForBody()
	}
	if options.Proxy != "" {
		generator.Modules["net/url"] = true
	}
	if options.User != "" {
		generator.Modules["encoding/base64"] = true
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

// helper functions

func NewStringForData(generator *PythonGenerator, data *httpgen_common.DataOption) (string, string) {
	var result string
	var name string
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("var buffer bytes.Buffer\n")
			buffer.WriteString("    content, err := ioutil.ReadFile(\"")
			buffer.WriteString(data.Value[1:])
			buffer.WriteString("\")\n")
			buffer.WriteString("    if err != nil {\n")
			buffer.WriteString("        log.Fatal(err)\n")
			buffer.WriteString("    }\n")
			buffer.WriteString("    buffer.WriteString(strings.Replace(string(content), \"\\n\", \"\", -1))")
			result = buffer.String()
			name = "&buffer"
			generator.Modules["strings"] = true
		} else {
			result = ""
			name = fmt.Sprintf("\"%s\"", escapeDQ(strings.Replace(data.Value, "\n", "", -1)))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("file, err := os.Open(\"")
			buffer.WriteString(data.Value[1:])
			buffer.WriteString("\")\n")
			buffer.WriteString("    if err != nil {\n")
			buffer.WriteString("        log.Fatal(err)\n")
			buffer.WriteString("    }\n")
			result = buffer.String()
			name = "file"
			generator.Modules["os"] = true
		} else {
			result = fmt.Sprintf("buffer := bytes.NewBufferString(\"%s\")\n", escapeDQ(data.Value))
			name = "buffer"
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("var buffer bytes.Buffer\n")
			buffer.WriteString("    content, err := ioutil.ReadFile(\"")
			buffer.WriteString(data.Value[1:])
			buffer.WriteString("\")\n")
			buffer.WriteString("    if err != nil {\n")
			buffer.WriteString("        log.Fatal(err)\n")
			buffer.WriteString("    }\n")
			buffer.WriteString("    buffer.WriteString(url.QueryEscape(string(content)))")
			result = buffer.String()
			name = "&buffer"
		} else {
			result = ""
			name = fmt.Sprintf("urllib.parse.quote_plus(\"%s\")", escapeDQ(data.Value))
			generator.Modules["urllib.parse"] = true
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, name
}

func StringForData(generator *PythonGenerator, data *httpgen_common.DataOption) string {
	var result string
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("    {\n")
			buffer.WriteString(fmt.Sprintf("        content, err := ioutil.ReadFile(\"%s\")\n", data.Value[1:]))
			buffer.WriteString("        if err != nil {\n")
			buffer.WriteString("            log.Fatal(err)\n")
			buffer.WriteString("        }\n")
			buffer.WriteString("        buffer.WriteString(strings.Replace(string(content), \"\\n\", \"\", -1))\n")
			buffer.WriteString("    }\n")
			result = buffer.String()
			generator.Modules["strings"] = true
		} else {
			result = fmt.Sprintf("        \"%s\",\n", escapeDQ(strings.Replace(data.Value, "\n", "", -1)))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("    {\n")
			buffer.WriteString(fmt.Sprintf("        file, err := os.Open(\"%s\")\n", data.Value[1:]))
			buffer.WriteString("        if err != nil {\n")
			buffer.WriteString("            log.Fatal(err)\n")
			buffer.WriteString("        }\n")
			buffer.WriteString("        io.Copy(&buffer, file)\n")
			buffer.WriteString("    }\n")
			result = buffer.String()
			generator.Modules["os"] = true
			generator.Modules["io"] = true
		} else {
			result = fmt.Sprintf("    buffer.WriteString(\"%s\")\n", escapeDQ(data.Value))
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("    {\n")
			buffer.WriteString(fmt.Sprintf("        content, err := ioutil.ReadFile(\"%s\")\n", data.Value[1:]))
			buffer.WriteString("        if err != nil {\n")
			buffer.WriteString("            log.Fatal(err)\n")
			buffer.WriteString("        }\n")
			buffer.WriteString("        buffer.WriteString(url.QueryEscape(string(content)))\n")
			buffer.WriteString("    }\n")
			result = buffer.String()
		} else {
			result = fmt.Sprintf("    buffer.WriteString(url.QueryEscape(\"%s\"))\n", escapeDQ(data.Value))
		}
		generator.Modules["net/url"] = true
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result
}

func FormString(generator *PythonGenerator, data *httpgen_common.DataOption) string {
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
			var contentType string
			fragments := strings.Split(field[1][1:], ";")
			sourceFile := fragments[0]
			sentFileName := fragments[0]
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "filename=") {
					sentFileName = fragment[9:]
				} else if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
				}
			}
			buffer.WriteString("    {\n")
			if contentType != "" {
				buffer.WriteString("        header := make(textproto.MIMEHeader)\n")
				buffer.WriteString(fmt.Sprintf("        header.Add(\"Content-Disposition\", \"form-data; name=\\\"%s\\\"; filename=\\\"%s\\\"\")\n", field[0], sentFileName))
				buffer.WriteString(fmt.Sprintf("        header.Add(\"Content-Type\", \"%s\")\n", contentType))
				buffer.WriteString("        fileWriter, err := writer.CreatePart(header)\n")
				buffer.WriteString("        if err != nil {\n")
				buffer.WriteString("            log.Fatal(err)\n")
				buffer.WriteString("        }\n")
				generator.Modules["net/textproto"] = true
			} else {
				buffer.WriteString(fmt.Sprintf("        fileWriter, err := writer.CreateFormFile(\"%s\", \"%s\")\n", field[0], sentFileName))
				buffer.WriteString("        if err != nil {\n")
				buffer.WriteString("            log.Fatal(err)\n")
				buffer.WriteString("        }\n")
			}
			buffer.WriteString(fmt.Sprintf("        file, err := os.Open(\"%s\")\n", sourceFile))
			buffer.WriteString("        if err != nil {\n")
			buffer.WriteString("            log.Fatal(err)\n")
			buffer.WriteString("        }\n")
			buffer.WriteString("        io.Copy(fileWriter, file)\n")
			buffer.WriteString("    }\n")
			result = buffer.String()
			generator.Modules["os"] = true
			generator.Modules["io"] = true
		} else if strings.HasPrefix(field[1], "<") {
			var buffer bytes.Buffer
			var contentType string
			fragments := strings.Split(field[1][1:], ";")
			sourceFile := fragments[0]
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
				}
			}
			buffer.WriteString("    {\n")
			buffer.WriteString("        header := make(textproto.MIMEHeader)\n")
			buffer.WriteString(fmt.Sprintf("        header.Add(\"Content-Disposition\", \"form-data; name=\\\"%s\\\"\")\n", field[0]))
			if contentType != "" {
				buffer.WriteString(fmt.Sprintf("        header.Add(\"Content-Type\", \"%s\")\n", contentType))
			}
			buffer.WriteString("        fileWriter, err := writer.CreatePart(header)\n")
			buffer.WriteString("        if err != nil {\n")
			buffer.WriteString("            log.Fatal(err)\n")
			buffer.WriteString("        }\n")
			buffer.WriteString(fmt.Sprintf("        file, err := os.Open(\"%s\")\n", sourceFile))
			buffer.WriteString("        if err != nil {\n")
			buffer.WriteString("            log.Fatal(err)\n")
			buffer.WriteString("        }\n")
			buffer.WriteString("        io.Copy(fileWriter, file)\n")
			buffer.WriteString("    }\n")
			result = buffer.String()
			generator.Modules["net/textproto"] = true
			generator.Modules["os"] = true
			generator.Modules["io"] = true
		} else {
			result = fmt.Sprintf("    writer.WriteField(\"%s\", \"%s\")\n", field[0], field[1])
		}
	case httpgen_common.FormStringType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
		result = fmt.Sprintf("    writer.WriteField(\"%s\", \"%s\")\n", field[0], field[1])
	}
	generator.Modules["mime/multipart"] = true
	return result
}
