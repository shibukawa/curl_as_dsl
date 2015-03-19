package vim_script_client

import (
	"bytes"
	"fmt"
	"github.com/shibukawa/curl_as_dsl/httpgen_common"
	"log"
	"net/url"
	"os"
	"strings"
)

func escapeDQ(src string) string {
	return strings.Replace(strings.Replace(src, "\"", "\\\"", -1), "\\", "\\\\", -1)
}

type VimScriptGenerator struct {
	Options *httpgen_common.CurlOptions

	HasBody               bool
	Body                  string
	PrepareBody           string
	FinalizeBodyBuffer    bytes.Buffer
	AdditionalDeclaration string
	specialHeaders        []string
}

func NewVimScriptGenerator(options *httpgen_common.CurlOptions) *VimScriptGenerator {
	return &VimScriptGenerator{Options: options}
}

//--- Getter methods called from template

func (self VimScriptGenerator) Url() string {
	return fmt.Sprintf(`'%s'`, self.Options.Url)
}

func (self VimScriptGenerator) HasHeader() bool {
	return len(self.Options.Header) != 0 || len(self.specialHeaders) != 0
}

func (self VimScriptGenerator) BodyContent() string {
	if self.Body == "" {
		return ", ''"
	}
	return fmt.Sprintf(", %s", self.Body)
}

func (self VimScriptGenerator) Header() string {
	if len(self.Options.Header) == 0 && len(self.specialHeaders) == 0 {
		return ""
	}
	return fmt.Sprintf(", s:headers")
}

func (self VimScriptGenerator) PrepareHeader() string {
	if len(self.Options.Header) == 0 && len(self.specialHeaders) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("let s:headers = {\n  ")
	first := true
	for _, header := range self.Options.GroupedHeaders() {
		if first {
			first = false
		} else {
			buffer.WriteString(",\n  ")
		}
		fmt.Fprintf(&buffer, "\\\"%s\": \"%s\"", strings.TrimSpace(header.Key), strings.TrimSpace(header.Values[0]))
	}
	for _, header := range self.specialHeaders {
		if first {
			first = false
		} else {
			buffer.WriteString(",\n  ")
		}
		buffer.WriteString(header)
	}
	buffer.WriteString("\n  \\}\n")
	return buffer.String()
}

func (self VimScriptGenerator) FinalizeBody() string {
	if self.FinalizeBodyBuffer.Len() > 0 {
		return "\n" + self.FinalizeBodyBuffer.String()
	}
	return ""
}

func (self VimScriptGenerator) Method() string {
	method := strings.ToLower(self.Options.Method())
	if method != "get" && method != "post" {
		log.Fatal("VimScript only supports get, post")
	}
	return method
}

//--- Setter/Getter methods

func (self *VimScriptGenerator) AddMultiPartCode() {
	self.AdditionalDeclaration = `let s:BOUNDARY = '----------ThIs_Is_tHe_bouNdaRY_$'

function! s:encode_multipart_formdata(fields, files)
    let lines = []
    for field in a:fields
        call add(lines, '--'. s:BOUNDARY)
        call add(lines, printf('Content-Disposition: form-data; name="%s"', field.key))
        if has_key(field, 'contenttype')
            call add(lines, 'Content-Type: '. field.contenttype)
        endif
        call add(lines, '')
        call add(lines, field.value)
    endfor
    for file in a:files
        call add(lines, '--'. s:BOUNDARY)
        call add(lines, printf('Content-Disposition: form-data; name="%s"; filename="%s"', file.key, file.filename))
        call add(lines, 'Content-Type: '. file.contenttype)
        call add(lines, '')
        call add(lines, join(readfile(file.sourcefile), "\n"))
    endfor
    call add(lines, '--'. s:BOUNDARY. '--')
    return join(lines, "\r\n")
endfunction
`
	boundary := "----------ThIs_Is_tHe_bouNdaRY_$"
	self.Options.InsertContentTypeHeader(fmt.Sprintf("multipart/form-data; boundary=%s", boundary))
}

func (self *VimScriptGenerator) SetDataForBody() {
	var buffer bytes.Buffer
	if self.Options.CanUseSimpleForm() {
		self.SetDataForForm()
		self.FinalizeBodyBuffer.WriteString("unlet! s:body\n")
	} else if len(self.Options.ProcessedData) == 1 {
		self.Body = StringForData(self, &self.Options.ProcessedData[0])
		self.PrepareBody = buffer.String()
	} else {
		for i, data := range self.Options.ProcessedData {
			if i == 0 {
				buffer.WriteString("let s:body = join([\n  \\")
			} else {
				buffer.WriteString(",\n  \\")
			}
			buffer.WriteString(StringForData(self, &data))
		}
		buffer.WriteString("\n  \\], \"&\")\n")
		self.Body = "s:body"
		self.FinalizeBodyBuffer.WriteString("unlet! s:body\n")
		self.PrepareBody = buffer.String()
	}
	self.HasBody = true
}

func (self *VimScriptGenerator) SetDataForForm() {
	entries := make(map[string][]string)
	for _, data := range self.Options.ProcessedData {
		singleData, _ := url.ParseQuery(data.Value)
		for key, values := range singleData {
			entries[key] = append(entries[key], values[0])
		}
	}

	var buffer bytes.Buffer
	count := 1
	for key, values := range entries {
		if count == 1 {
			buffer.WriteString("let s:body = {")
		} else {
			buffer.WriteString(`, `)
		}
		fmt.Fprintf(&buffer, `"%s": "%s"`, key, values[0])
		count++
	}
	buffer.WriteString("}\n")

	self.PrepareBody = buffer.String()
	self.HasBody = true
	self.Body = "s:body"
}

func (self *VimScriptGenerator) SetFormForBody() {
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
		buffer.WriteString("\nlet s:fields = [\n")
		for _, value := range fields {
			buffer.WriteString(value)
		}
		buffer.WriteString("  \\]\n")
		self.FinalizeBodyBuffer.WriteString("unlet! s:fields\n")
	}
	if len(files) > 0 {
		if len(fields) > 0 {
			self.Body = "s:encode_multipart_formdata(s:fields, s:files)"
		} else {
			self.Body = "s:encode_multipart_formdata([], s:files)"
		}
		buffer.WriteString("\nlet s:files = [\n")
		for _, value := range files {
			buffer.WriteString(value)
		}
		buffer.WriteString("  \\]\n")
		self.FinalizeBodyBuffer.WriteString("unlet! s:files\n")
	} else {
		self.Body = "s:encode_multipart_formdata(s:fields, [])"
	}
	self.PrepareBody = buffer.String()
	self.HasBody = true
}

/*
	Dispatcher function of curl command
	This is an exported function and called from httpgen.
*/
func ProcessCurlCommand(options *httpgen_common.CurlOptions) (string, interface{}) {
	generator := NewVimScriptGenerator(options)

	if options.ProcessedData.HasData() {
		generator.Options.InsertContentTypeHeader("application/x-www-form-urlencoded")
		generator.SetDataForBody()
	} else if options.ProcessedData.HasForm() {
		generator.SetFormForBody()
	}
	if generator.Options.User != "" {
		generator.specialHeaders = append(generator.specialHeaders, fmt.Sprintf("\\'Authorization': 'Basic '. webapi#base64#b64encode('%s')", generator.Options.User))
	}

	return "full", *generator
}

// helper functions

func StringForData(generator *VimScriptGenerator, data *httpgen_common.DataOption) string {
	var result string
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("join(readfile('%s'), '')", data.Value[1:])
		} else {
			result = fmt.Sprintf(`"%s"`, escapeDQ(strings.Replace(data.Value, "\n", "", -1)))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("join(readfile('%s'), \"\\n\")", data.Value[1:])
		} else {
			result = fmt.Sprintf(`"%s"`, escapeDQ(data.Value))
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = fmt.Sprintf("webapi#http#encodeURIComponent(join(readfile('%s'), \"\\n\"))", data.Value[1:])
		} else {
			result = fmt.Sprintf(`webapi#http#encodeURIComponent("%s")`, escapeDQ(data.Value))
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result
}

func FormString(generator *VimScriptGenerator, data *httpgen_common.DataOption) string {
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
			fmt.Fprintf(&buffer, "  \\{'key': '%s', 'sourcefile': '%s', ", field[0], fragments[0])

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
			fmt.Fprintf(&buffer, "'filename': '%s', ", sentFileName)

			// sent file name
			if contentType != "" {
				fmt.Fprintf(&buffer, "'contenttype': '%s'", contentType)
			} else {
				buffer.WriteString("'contenttype': 'application/octet-stream'")
			}
			buffer.WriteString("},\n")
			result = buffer.String()
		} else if strings.HasPrefix(field[1], "<") {
			var buffer bytes.Buffer

			fragments := strings.Split(field[1][1:], ";")

			// field name, content
			fmt.Fprintf(&buffer, "  \\{'key':'%s', 'value': join(readfile('%s'), \"\\n\")", field[0], fragments[0])

			var contentType string
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
				}
			}
			if contentType != "" {
				fmt.Fprintf(&buffer, ", 'contenttype': '%s'", contentType)
			}
			buffer.WriteString("},\n")
			result = buffer.String()
		} else {
			result = fmt.Sprintf("  \\{'key': '%s', 'value': \"%s\"},\n", field[0], escapeDQ(field[1]))
		}
	case httpgen_common.FormStringType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
		result = fmt.Sprintf("  \\{'key': '%s', 'value': \"%s\"},\n", field[0], escapeDQ(field[1]))
	}
	return result
}
