package php_client

import (
	"bytes"
	"fmt"
	"github.com/shibukawa/curl_as_dsl/httpgen_common"
	"net/url"
	"os"
	"strings"
)

type PHPGenerator struct {
	Options *httpgen_common.CurlOptions

	HasBody               bool
	Body                  string
	PrepareBody           string
	queries               [][]string
	extraUrl              string
	AdditionalDeclaration string
	specialHeaders        []string
}

func NewPHPGenerator(options *httpgen_common.CurlOptions) *PHPGenerator {
	result := &PHPGenerator{Options: options}

	return result
}

//--- Getter methods called from template

func (self PHPGenerator) Url() string {
	return fmt.Sprintf(`"%s"%s`, self.Options.Url, self.extraUrl)
}

func (self PHPGenerator) HasHeader() bool {
	return len(self.Options.Header) != 0 || len(self.specialHeaders) != 0
}

func (self PHPGenerator) Header() string {
	if len(self.Options.Header) == 0 && len(self.specialHeaders) == 0 {
		return ""
	}
	return ",\n    \"header\" => $headers"
}

func (self PHPGenerator) PrepareHeader() string {
	if len(self.Options.Header) == 0 && len(self.specialHeaders) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("\n$headers = ")
	if len(self.Options.Header)+len(self.specialHeaders) == 1 {
		for _, header := range self.Options.Header {
			fmt.Fprintf(&buffer, "\"%s\";\n", header)
		}
		for _, header := range self.specialHeaders {
			buffer.WriteString(header)
			buffer.WriteString(";\n")
		}
	} else {
		buffer.WriteString("\n")
		for i, header := range self.Options.Header {
			fmt.Fprintf(&buffer, `  "%s\n"`, header)
			if i != len(self.Options.Header)-1 || len(self.specialHeaders) > 0 {
				buffer.WriteString(" .\n")
			} else {
				buffer.WriteString(";\n")
			}
		}
		for i, header := range self.specialHeaders {
			fmt.Fprintf(&buffer, `  %s`, header)
			if i == len(self.specialHeaders)-1 {
				buffer.WriteString(";\n")
			} else {
				buffer.WriteString(".\n")
			}
		}
	}
	return buffer.String()
}

func (self PHPGenerator) Method() string {
	return self.Options.Method()
}

func (self PHPGenerator) Content() string {
	if self.Body != "" {
		var buffer bytes.Buffer
		buffer.WriteString(",\n    \"content\" => ")
		buffer.WriteString(self.Body)
		return buffer.String()
	}
	return ""
}

//--- Setter/Getter methods

func (self *PHPGenerator) AddMultiPartCode() {
	self.AdditionalDeclaration = `
$BOUNDARY = "---------------------".substr(md5(rand(0,32000)), 0, 10);

function encode_multipart_formdata($fields, $files, $boundary) {
  $result = "";
  $finfo = finfo_open(FILEINFO_MIME_TYPE);
  foreach($fields as $field) {
    $result .= "--" . $boundary . "\r\n";
    $result .= "Content-Disposition: form-data; name=\"" . $field["key"] . "\"\r\n";
    if ($field["content_type"] != '') {
      $result .= "Content-Type: " . $field["content_type"] . "\r\n";
    }
    $result .= "\r\n" . $field["value"] . "\r\n";
  }
  foreach($files as $file) {
    $result .= "--" . $boundary . "\r\n";
    $result .= "Content-Disposition: form-data; name=\"" . $file["key"] . "\"; filename=\"" . $file["filename"] . "\"\r\n";
    if ($file["content_type"] != '') {
      $result .= "Content-Type: " . $file["content_type"] . "\r\n";
    } else {
      $result .= "Content-Type: " . finfo_file($finfo, $file["source_file"]) . "\r\n";
    }
    $result .= "\r\n" . file_get_contents($file["source_file"]) . "\r\n";
  }
  $result .= $boundary . "--\r\n";
  return $result;
}
`
	self.Options.InsertContentTypeHeader("multipart/form-data; boundary={$BOUNDARY}")
}

func (self *PHPGenerator) SetDataForUrl() {
	var buffer bytes.Buffer
	if self.Options.CanUseSimpleForm() {
		entries := make(map[string][]string)
		for _, data := range self.Options.ProcessedData {
			singleData, _ := url.ParseQuery(data.Value)
			for key, values := range singleData {
				entries[key] = append(entries[key], values[0])
			}
		}
		fmt.Fprintf(&buffer, "\n%s = http_build_query([", "$query")
		count := 0
		for key, values := range entries {
			if count != 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(fmt.Sprintf("\n  \"%s\" => \"%s\"", key, values[0]))
			count++
		}
		buffer.WriteString("\n], PHP_QUERY_RFC1738);")
		self.extraUrl = ` . "?" . $query`
	} else {
		// Use bytes.Buffer to create URL option string
		if len(self.Options.ProcessedData) == 1 {
			self.extraUrl = fmt.Sprintf(` . "?" . %s`, NewStringForData(self, &self.Options.ProcessedData[0]))
		} else {
			for i, data := range self.Options.ProcessedData {
				if i == 0 {
					buffer.WriteString("\n$query = \n  ")
				}
				buffer.WriteString(StringForData(self, &data))
				if i != len(self.Options.ProcessedData)-1 {
					buffer.WriteString(".\n  ")
				}
			}
			buffer.WriteString(";\n")
			self.extraUrl = ` . "?" . $query`
		}
	}
	self.PrepareBody = buffer.String()
}

func (self *PHPGenerator) SetDataForBody(varName string) {
	var buffer bytes.Buffer
	if len(self.Options.ProcessedData) == 1 {
		self.Body = NewStringForData(self, &self.Options.ProcessedData[0])
	} else {
		for i, data := range self.Options.ProcessedData {
			if i == 0 {
				fmt.Fprintf(&buffer, "\n%s = \n  ", varName)
			} else {
				buffer.WriteString(" . \"&\" .\n  ")
			}
			buffer.WriteString(StringForData(self, &data))
		}
		buffer.WriteString(";\n")
		self.Body = varName
	}
	self.PrepareBody = buffer.String()
}

func (self *PHPGenerator) SetDataForForm(varName string) {
	entries := make(map[string][]string)
	for _, data := range self.Options.ProcessedData {
		singleData, _ := url.ParseQuery(data.Value)
		for key, values := range singleData {
			entries[key] = append(entries[key], values[0])
		}
	}

	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "\n%s = http_build_query([", varName)
	count := 0
	for key, values := range entries {
		if count != 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(fmt.Sprintf("\n  \"%s\" => \"%s\"", key, values[0]))
		count++
	}
	buffer.WriteString("\n], PHP_QUERY_RFC1738);")

	self.PrepareBody = buffer.String()
	self.HasBody = true
	self.Body = "values"
}

func (self *PHPGenerator) SetFormForBody() {
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
		buffer.WriteString("\n$fields = [\n")
		for _, value := range fields {
			buffer.WriteString(value)
		}
		buffer.WriteString("];\n")
	}
	if len(files) > 0 {
		if len(fields) > 0 {
			self.Body = "encode_multipart_formdata($fields, $files, $BOUNDARY)"
		} else {
			self.Body = "encode_multipart_formdata(array(), $files, $BOUNDARY)"
		}
		buffer.WriteString("\n$files = [\n")
		for _, value := range files {
			buffer.WriteString(value)
		}
		buffer.WriteString("];\n")
	} else {
		self.Body = "encode_multipart_formdata($fields, array(), $BOUNDARY)"
	}
	self.PrepareBody = buffer.String()
	self.HasBody = true
}

/*
	Dispatcher function of curl command
	This is an exported function and called from httpgen.
*/
func ProcessCurlCommand(options *httpgen_common.CurlOptions) (string, interface{}) {
	generator := NewPHPGenerator(options)

	if options.ProcessedData.HasData() {
		if options.Get {
			generator.SetDataForUrl()
		} else {
			generator.Options.InsertContentTypeHeader("application/x-www-form-urlencoded")
			generator.SetDataForBody("$content")
		}
	} else if options.ProcessedData.HasForm() {
		generator.SetFormForBody()
	}
	if generator.Options.User != "" {
		generator.specialHeaders = append(generator.specialHeaders, fmt.Sprintf(`"Authorization: Basic " . base64_encode('%s') . "\n"`, generator.Options.User))
	}

	return "full", *generator
}

// helper functions

func NewStringForData(generator *PHPGenerator, data *httpgen_common.DataOption) string {
	var result string
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf(`str_replace(array("\r\n", "\n", "\r"), "", file_get_contents("%s"))`, data.Value[1:])
		} else {
			result = fmt.Sprintf(`"%s"`, strings.Replace(data.Value, "\n", "", -1))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf(`file_get_contents("%s")`, data.Value[1:])
		} else {
			result = fmt.Sprintf(`"%s"`, data.Value)
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("urlencode(file_get_contents('%s'))", data.Value[1:])
		} else {
			result = fmt.Sprintf("urlencode('%s')", data.Value)
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result
}

func StringForData(generator *PHPGenerator, data *httpgen_common.DataOption) string {
	var result string
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("str_replace(array(\"\\r\\n\", \"\\n\", \"\\r\"), '', file_get_contents('%s'))", data.Value[1:])
		} else {
			result = fmt.Sprintf(`"%s"`, strings.Replace(data.Value, "\n", "", -1))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("file_get_contents(\"%s\")", data.Value[1:])
		} else {
			result = fmt.Sprintf(`"%s"`, data.Value)
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("urlencode(file_get_contents(\"%s\"))", data.Value[1:])
		} else {
			result = fmt.Sprintf("urlencode(\"%s\")", data.Value)
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result
}

func FormString(generator *PHPGenerator, data *httpgen_common.DataOption) string {
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
			fmt.Fprintf(&buffer, "  array(\"key\"=>\"%s\", \"source_file\"=>\"%s\", ", field[0], fragments[0])

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
			fmt.Fprintf(&buffer, "\"filename\"=>\"%s\", ", sentFileName)

			// sent file name
			if contentType != "" {
				fmt.Fprintf(&buffer, "\"content_type\"=>\"%s\"),\n", contentType)
			} else {
				buffer.WriteString("\"content_type\"=>\"\"),\n")
			}
			result = buffer.String()
		} else if strings.HasPrefix(field[1], "<") {
			var buffer bytes.Buffer

			fragments := strings.Split(field[1][1:], ";")

			// field name, content
			fmt.Fprintf(&buffer, `  array("key"=>"%s", "value"=>file_get_contents("%s"), "content_type"=>`, field[0], fragments[0])

			var contentType string
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
				}
			}
			if contentType == "" {
				buffer.WriteString("\"\"),\n")
			} else {
				fmt.Fprintf(&buffer, "\"%s\"),\n", contentType)
			}
			result = buffer.String()
		} else {
			result = fmt.Sprintf("  array(\"key\"=>\"%s\", \"value\"=>\"%s\", \"content_type\"=>\"\"),\n", field[0], field[1])
		}
	case httpgen_common.FormStringType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
		result = fmt.Sprintf("  array(\"key\"=>\"%s\", \"value\"=>\"%s\", \"content_type\"=>\"\"),\n", field[0], field[1])
	}
	return result
}
