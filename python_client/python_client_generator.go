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

type PythonGenerator struct {
	Options *httpgen_common.CurlOptions
	Modules map[string]bool

	HasBody               bool
	Body                  string
	PrepareBody           string
	extraUrl              string
	AdditionalDeclaration string
	specialHeaders        []string
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

func (self PythonGenerator) HasHeader() bool {
	return len(self.Options.Header) != 0 || len(self.specialHeaders) != 0
}

func (self PythonGenerator) Header() string {
	if len(self.Options.Header) == 0 && len(self.specialHeaders) == 0 {
		return ""
	}
	return "headers"
}

func (self PythonGenerator) PrepareHeader() string {
	if len(self.Options.Header) == 0 && len(self.specialHeaders) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("headers = {\n")
	for _, header := range self.Options.Header {
		headers := strings.Split(header, ":")
		buffer.WriteString(fmt.Sprintf("        \"%s\": \"%s\",\n", strings.TrimSpace(headers[0]), strings.TrimSpace(headers[1])))
	}
	for _, header := range self.specialHeaders {
		buffer.WriteString(header)
	}
	buffer.WriteString("    }\n    ")
	return buffer.String()
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

//--- Setter/Getter methods

func (self *PythonGenerator) AddMultiPartCode() {
	self.AdditionalDeclaration = `
BOUNDARY = '----------ThIs_Is_tHe_bouNdaRY_$'

def encode_multipart_formdata(fields, files):
    """
    http://code.activestate.com/recipes/146306-http-client-to-post-using-multipartform-data/
    """
    CRLF = '\r\n'
    L = []
    for key, value, contenttype in fields:
        L.append('--' + BOUNDARY)
        L.append('Content-Disposition: form-data; name="%s"' % key)
        if contenttype:
            L.append('Content-Type: %s' % contenttype)
        L.append('')
        L.append(value)
    for key, sourcefile, filename, contenttype in files:
        L.append('--' + BOUNDARY)
        L.append('Content-Disposition: form-data; name="%s"; filename="%s"' % (key, filename))
        L.append('Content-Type: %s' % contenttype)
        L.append('')
        L.append(open(sourcefile).read())
        L.append('--' + BOUNDARY + '--')
        L.append('')
    return CRLF.join(L)
`
	boundary := "----------ThIs_Is_tHe_bouNdaRY_$"
	self.Options.InsertContentTypeHeader(fmt.Sprintf("multipart/form-data; boundary=%s", boundary))
}

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
		buffer.WriteString("    ]\n    ")
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
			entries[key] = append(entries[key], values[0])
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
	buffer.WriteString("    })\n    ")

	self.PrepareBody = buffer.String()
	self.HasBody = true
	self.Body = "values"
	self.Modules["urllib.parse"] = true
}

func (self *PythonGenerator) SetFormForBody() {
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
		buffer.WriteString("    ]\n    ")
	}
	if len(files) > 0 {
		if len(fields) > 0 {
			buffer.WriteString("    ")
			self.Body = "encode_multipart_formdata(fields, files)"
		} else {
			self.Body = "encode_multipart_formdata([], files)"
		}
		buffer.WriteString("files = [\n")
		for _, value := range files {
			buffer.WriteString(value)
		}
		buffer.WriteString("    ]\n    ")
	} else {
		self.Body = "encode_multipart_formdata(fields, [])"
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
		generator.SetFormForBody()
	}
	if generator.Options.User != "" {
		generator.specialHeaders = append(generator.specialHeaders, fmt.Sprintf("        'Authorization': 'Basic %%s' %% base64.b64encode(b'%s').decode('ascii'),\n", generator.Options.User))
		generator.Modules["base64"] = true

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
			result = ""
			name = fmt.Sprintf("open(r'%s').read().replace('\\n', '')", data.Value[1:])
		} else {
			result = ""
			name = fmt.Sprintf("r'%s'", strings.Replace(data.Value, "\n", "", -1))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = ""
			name = fmt.Sprintf("open(r'%s').read()", data.Value[1:])
		} else {
			result = ""
			name = fmt.Sprintf("'%s'", data.Value)
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = ""
			name = fmt.Sprintf("urllib.parse.quote_plus(open(r'%s').read())", data.Value[1:])
		} else {
			result = ""
			name = fmt.Sprintf("urllib.parse.quote_plus(r'%s')", data.Value)
		}
		generator.Modules["urllib.parse"] = true
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
			result = fmt.Sprintf("        open('%s').read().replace('\\n', ''),\n", data.Value[1:])
		} else {
			result = fmt.Sprintf("        r'%s',\n", strings.Replace(data.Value, "\n", "", -1))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("        open('%s').read(),\n", data.Value[1:])
		} else {
			result = fmt.Sprintf("        r'%s',\n", data.Value)
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("        urllib.parse.quote_plus(open(r'%s').read()),\n", data.Value[1:])
		} else {
			result = fmt.Sprintf("        urllib.parse.quote_plus(r'%s'),\n", data.Value)
		}
		generator.Modules["urllib.parse"] = true
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
			fragments := strings.Split(field[1][1:], ";")

			// field name, source file name
			buffer.WriteString(fmt.Sprintf("        (r'%s', r'%s', ", field[0], fragments[0]))

			var contentType string
			sentFileName := fragments[0]
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "filename=") {
					sentFileName = fragment[9:]
				} else if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
				}
			}
			// sent file name
			buffer.WriteString(fmt.Sprintf("r'%s', ", sentFileName))

			// sent file name
			if contentType != "" {
				buffer.WriteString(fmt.Sprintf("r'%s'", contentType))
			} else {
				buffer.WriteString(fmt.Sprintf("mimetypes.guess_type(r'%s')[0] or 'application/octet-stream'", fragments[0]))
				generator.Modules["mimetypes"] = true
			}
			buffer.WriteString("),\n")
			result = buffer.String()
		} else if strings.HasPrefix(field[1], "<") {
			var buffer bytes.Buffer

			fragments := strings.Split(field[1][1:], ";")

			// field name, content
			buffer.WriteString(fmt.Sprintf("        (r'%s', open(r'%s').read(), ", field[0], fragments[0]))

			var contentType string
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
				}
			}
			if contentType == "" {
				buffer.WriteString("None")
			} else {
				buffer.WriteString(fmt.Sprintf("r'%s'", contentType))
			}
			buffer.WriteString("),\n")
			result = buffer.String()
		} else {
			result = fmt.Sprintf("        (\"%s\", \"%s\", None),\n", field[0], field[1])
		}
	case httpgen_common.FormStringType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
		result = fmt.Sprintf("        (\"%s\", \"%s\", None),\n", field[0], field[1])
	}
	return result
}
