package golang

import (
	"fmt"
	"os"
	//"log"
	"bytes"
	"github.com/shibukawa/curl_as_dsl/common"
	"net/url"
	"strings"
)

type GoGenerator struct {
	Options *common.CurlOptions
	Modules map[string]bool

	Data         string
	DataVariable string
	ContentType  string
	HasBoundary  bool

	extraUrl string
}

func NewGoGenerator(options *common.CurlOptions) *GoGenerator {
	result := &GoGenerator{Options: options}
	result.Modules = make(map[string]bool)
	result.Modules["net/http"] = true
	result.Modules["log"] = true
	result.Modules["io/ioutil"] = true
	result.DataVariable = "nil"
	return result
}

//--- Getter methods called from template

func (self GoGenerator) Url() string {
	return fmt.Sprintf("\"%s\"%s", self.Options.Url, self.extraUrl)
}

func (self GoGenerator) Method() string {
	return self.Options.Method()
}

func (self GoGenerator) FilePath() string {
	return ""
}

func (self GoGenerator) PrepareClient() string {
	if self.Options.Proxy != "" {
		return fmt.Sprintf("proxyUrl, err := url.Parse(\"%s\")", self.Options.Proxy)
	}
	return ""
}

func (self GoGenerator) ClientBody() string {
	if !self.Options.Insecure && self.Options.Proxy == "" {
		return "{}"
	}
	var buffer bytes.Buffer
	buffer.WriteString(`{Transport: &http.Transport{`)
	if self.Options.Insecure {
		buffer.WriteString("TLSClientConfig: &tls.Config{InsecureSkipVerify: true},\n")
	}
	if self.Options.Proxy != "" {
		buffer.WriteString("Proxy: http.ProxyURL(proxyUrl),\n")
	}
	buffer.WriteString("}}")
	return buffer.String()
}

func (self GoGenerator) ModifyRequest() string {
	var buffer bytes.Buffer
	contentTypeInHeader := self.Options.FindContentTypeHeader()
	if contentTypeInHeader != "" {
		self.ContentType = contentTypeInHeader
	}

	// Set headers
	for _, header := range self.Options.Header {
		headers := strings.SplitN(header, ":", 2)
		if len(headers) == 2 {
			key := strings.TrimSpace(headers[0])
			value := strings.TrimSpace(headers[1])
			if strings.ToLower(key) == "content-type" && self.HasBoundary {
				fmt.Fprintf(&buffer, "request.Header.Add(\"%s\", \"%s; boundary=\" + writer.Boundary())\n", key, value)
			} else {
				fmt.Fprintf(&buffer, "request.Header.Add(\"%s\", \"%s\")\n", key, value)
			}
		}
	}

	if self.Options.User != "" {
		fmt.Fprintf(&buffer, "request.Header.Add(\"Authorization\", \"Basic \" + base64.StdEncoding.EncodeToString([]byte(\"%s\")))\n", self.Options.User)
	}

	for _, cookie := range self.Options.Cookie {
		fragments := strings.SplitN(cookie, "=", 2)
		if len(fragments) == 2 {
			name := strings.TrimSpace(fragments[0])
			value := strings.TrimSpace(fragments[1])
			fmt.Fprintf(&buffer, "request.AddCookie(&http.Cookie{Name: \"%s\", Value: \"%s\"})\n", name, value)
		}
	}

	if self.Options.AWSV2 != "" {
		buffer.WriteString(fmt.Sprintf("SignAWSV2(request, \"\", \"%s\")\n", self.ContentType))
	}

	return buffer.String()
}

func (self GoGenerator) AdditionalDeclaration() string {
	var buffer bytes.Buffer

	if self.Options.AWSV2 != "" {
		fragments := strings.SplitN(self.Options.AWSV2, ":", 2)
		if len(fragments) == 2 {
			fmt.Fprintf(&buffer, `
				func SignAWSV2(req *http.Request, md5, contentType string) {
					dateStr := time.Now().UTC().Format(time.RFC1123Z)
					req.Header.Set("Date", dateStr)
					if md5 != "" {
						req.Header.Set("Content-MD5", md5)
					}
					strToSign := fmt.Sprintf("%%s\n%%s\n%%s\n%%s\n%%s", req.Method, md5, contentType, dateStr, req.URL.Path)
					hash := hmac.New(sha1.New, []byte("%s"))
					hash.Write([]byte(strToSign))
					signature := make([]byte, base64.StdEncoding.EncodedLen(hash.Size()))
					base64.StdEncoding.Encode(signature, hash.Sum(nil))
					req.Header.Set("Authorization", fmt.Sprintf("AWS %%s:%%s", "%s", string(signature)))
				}
			`, fragments[1], fragments[0])
		}
	}

	return buffer.String()
}

//--- Setter/Getter methods

func (self *GoGenerator) SetDataForBody() {
	var buffer bytes.Buffer
	if len(self.Options.ProcessedData) == 1 {
		var body string
		body, self.DataVariable = NewStringForData(self, &self.Options.ProcessedData[0])
		buffer.WriteString(body)
	} else {
		for i, data := range self.Options.ProcessedData {
			if i > 0 {
				buffer.WriteString("buffer.WriteByte('&')\n")
			} else {
				buffer.WriteString("var buffer bytes.Buffer\n")
			}
			buffer.WriteString(StringForData(self, &data))
		}
		self.DataVariable = "&buffer"
	}
	self.Data = buffer.String()
}

func (self *GoGenerator) SetDataForUrl() {
	if self.Options.CanUseSimpleForm() {
		// Use url.Values to create URL option string
		self.SetDataForPostForm()
		self.extraUrl = " + \"?\" + values.Encode()"
	} else {
		// Use bytes.Buffer to create URL option string
		self.SetDataForBody()
		self.extraUrl = " + \"?\" + buffer.String()"
	}
}

func (self *GoGenerator) SetFormForBody() {
	var buffer bytes.Buffer
	buffer.WriteString("var buffer bytes.Buffer\n")
	buffer.WriteString("    writer := multipart.NewWriter(&buffer)\n")
	for _, data := range self.Options.ProcessedData {
		buffer.WriteString(FormString(self, &data))
	}
	buffer.WriteString("    writer.Close()\n")
	self.Data = buffer.String()
}

func (self *GoGenerator) SetDataForPostForm() {
	entries := make(map[string][]string)
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
			buffer.WriteString("values := url.Values{\n")
		}
		buffer.WriteString(`"` + key)
		buffer.WriteString(`": {`)
		for j, value := range values {
			if j == 0 {
				buffer.WriteString(`"`)
			} else {
				buffer.WriteString(`, "`)
			}
			buffer.WriteString(value)
			buffer.WriteString(`"`)
		}
		count++
		buffer.WriteString("},\n")
	}
	buffer.WriteString("}\n")

	self.Data = buffer.String()
	self.DataVariable = "values"
	self.Modules["net/url"] = true
}

func NewStringForData(generator *GoGenerator, data *common.DataOption) (string, string) {
	var result string
	var name string
	switch data.Type {
	case common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("var buffer bytes.Buffer\n")
			fmt.Fprintf(&buffer, "content, err := ioutil.ReadFile(\"%s\")\n", data.Value[1:])
			buffer.WriteString("if err != nil {\n")
			buffer.WriteString("    log.Fatal(err)\n")
			buffer.WriteString("}\n")
			buffer.WriteString("buffer.WriteString(strings.Replace(string(content), \"\\n\", \"\", -1))")
			result = buffer.String()
			name = "&buffer"
			generator.Modules["strings"] = true
		} else {
			result = fmt.Sprintf("buffer := bytes.NewBufferString(\"%s\")\n", escapeDQ(strings.Replace(data.Value, "\n", "", -1)))
			name = "buffer"
		}
		generator.Modules["bytes"] = true
	case common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			fmt.Fprintf(&buffer, "file, err := os.Open(\"%s\")\n", data.Value[1:])
			buffer.WriteString("if err != nil {\n")
			buffer.WriteString("    log.Fatal(err)\n")
			buffer.WriteString("}\n")
			result = buffer.String()
			name = "file"
			generator.Modules["os"] = true
		} else {
			result = fmt.Sprintf("buffer := bytes.NewBufferString(\"%s\")\n", escapeDQ(data.Value))
			name = "buffer"
			generator.Modules["bytes"] = true
		}
	case common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("var buffer bytes.Buffer\n")
			fmt.Fprintf(&buffer, "content, err := ioutil.ReadFile(\"%s\")\n", data.Value[1:])
			buffer.WriteString("if err != nil {\n")
			buffer.WriteString("    log.Fatal(err)\n")
			buffer.WriteString("}\n")
			buffer.WriteString("buffer.WriteString(url.QueryEscape(string(content)))")
			result = buffer.String()
			name = "&buffer"
		} else {
			result = fmt.Sprintf("buffer := bytes.NewBufferString(url.QueryEscape(\"%s\"))\n", escapeDQ(data.Value))
			name = "buffer"
		}
		generator.Modules["bytes"] = true
		generator.Modules["net/url"] = true
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, name
}

func StringForData(generator *GoGenerator, data *common.DataOption) string {
	var result string
	switch data.Type {
	case common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("{\n")
			fmt.Fprintf(&buffer, "content, err := ioutil.ReadFile(\"%s\")\n", data.Value[1:])
			buffer.WriteString("if err != nil {\n")
			buffer.WriteString("    log.Fatal(err)\n")
			buffer.WriteString("}\n")
			buffer.WriteString("buffer.WriteString(strings.Replace(string(content), \"\\n\", \"\", -1))\n")
			buffer.WriteString("}\n")
			result = buffer.String()
			generator.Modules["strings"] = true
		} else {
			result = fmt.Sprintf("    buffer.WriteString(\"%s\")\n", escapeDQ(strings.Replace(data.Value, "\n", "", -1)))
		}
	case common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("{\n")
			fmt.Fprintf(&buffer, "file, err := os.Open(\"%s\")\n", data.Value[1:])
			buffer.WriteString("if err != nil {\n")
			buffer.WriteString("    log.Fatal(err)\n")
			buffer.WriteString("}\n")
			buffer.WriteString("io.Copy(&buffer, file)\n")
			buffer.WriteString("}\n")
			result = buffer.String()
			generator.Modules["os"] = true
			generator.Modules["io"] = true
		} else {
			result = fmt.Sprintf("buffer.WriteString(\"%s\")\n", escapeDQ(data.Value))
		}
	case common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			var buffer bytes.Buffer
			buffer.WriteString("{\n")
			fmt.Fprintf(&buffer, "content, err := ioutil.ReadFile(\"%s\")\n", data.Value[1:])
			buffer.WriteString("if err != nil {\n")
			buffer.WriteString("    log.Fatal(err)\n")
			buffer.WriteString("}\n")
			buffer.WriteString("buffer.WriteString(url.QueryEscape(string(content)))\n")
			buffer.WriteString("}\n")
			result = buffer.String()
		} else {
			result = fmt.Sprintf("buffer.WriteString(url.QueryEscape(\"%s\"))\n", escapeDQ(data.Value))
		}
		generator.Modules["net/url"] = true
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	generator.Modules["bytes"] = true
	return result
}

func FormString(generator *GoGenerator, data *common.DataOption) string {
	var result string
	switch data.Type {
	case common.FormType:
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
			buffer.WriteString("{\n")
			if contentType != "" {
				buffer.WriteString("header := make(textproto.MIMEHeader)\n")
				fmt.Fprintf(&buffer, "header.Add(\"Content-Disposition\", \"form-data; name=\\\"%s\\\"; filename=\\\"%s\\\"\")\n", field[0], sentFileName)
				fmt.Fprintf(&buffer, "header.Add(\"Content-Type\", \"%s\")\n", contentType)
				buffer.WriteString("fileWriter, err := writer.CreatePart(header)\n")
				buffer.WriteString("if err != nil {\n")
				buffer.WriteString("    log.Fatal(err)\n")
				buffer.WriteString("}\n")
				generator.Modules["net/textproto"] = true
			} else {
				fmt.Fprintf(&buffer, "fileWriter, err := writer.CreateFormFile(\"%s\", \"%s\")\n", field[0], sentFileName)
				buffer.WriteString("if err != nil {\n")
				buffer.WriteString("    log.Fatal(err)\n")
				buffer.WriteString("}\n")
			}
			fmt.Fprintf(&buffer, "file, err := os.Open(\"%s\")\n", sourceFile)
			buffer.WriteString("if err != nil {\n")
			buffer.WriteString("    log.Fatal(err)\n")
			buffer.WriteString("}\n")
			buffer.WriteString("io.Copy(fileWriter, file)\n")
			buffer.WriteString("}\n")
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
			buffer.WriteString("{\n")
			buffer.WriteString("header := make(textproto.MIMEHeader)\n")
			fmt.Fprintf(&buffer, "header.Add(\"Content-Disposition\", \"form-data; name=\\\"%s\\\"\")\n", field[0])
			if contentType != "" {
				fmt.Fprintf(&buffer, "header.Add(\"Content-Type\", \"%s\")\n", contentType)
			}
			buffer.WriteString("fileWriter, err := writer.CreatePart(header)\n")
			buffer.WriteString("if err != nil {\n")
			buffer.WriteString("    log.Fatal(err)\n")
			buffer.WriteString("}\n")
			fmt.Fprintf(&buffer, "file, err := os.Open(\"%s\")\n", sourceFile)
			buffer.WriteString("if err != nil {\n")
			buffer.WriteString("    log.Fatal(err)\n")
			buffer.WriteString("}\n")
			buffer.WriteString("io.Copy(fileWriter, file)\n")
			buffer.WriteString("}\n")
			result = buffer.String()
			generator.Modules["net/textproto"] = true
			generator.Modules["os"] = true
			generator.Modules["io"] = true
		} else {
			result = fmt.Sprintf("writer.WriteField(\"%s\", \"%s\")\n", field[0], field[1])
		}
	case common.FormStringType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
		result = fmt.Sprintf("writer.WriteField(\"%s\", \"%s\")\n", field[0], field[1])
	}
	generator.Modules["bytes"] = true
	generator.Modules["mime/multipart"] = true
	return result
}
