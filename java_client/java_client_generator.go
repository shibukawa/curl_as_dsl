package java_client

import (
	"bytes"
	"fmt"
	"github.com/shibukawa/curl_as_dsl/httpgen_common"
	"log"
	"net/url"
	"os"
	"strings"
)

type JavaGenerator struct {
	Options *httpgen_common.CurlOptions
	Modules map[string]bool

	Url                    string
	IsHttps                bool
	HasBody                bool
	Body                   string
	PrepareBody            string
	AdditionalDeclaration  string
	specialHeaders         [][]string
	commonInitialize       []string
	mimeCounter            int
	formFileContentCounter int
}

func NewJavaGenerator(options *httpgen_common.CurlOptions) *JavaGenerator {
	result := &JavaGenerator{Options: options}
	result.Url = fmt.Sprintf("\"%s\"", options.Url)
	result.Modules = make(map[string]bool)
	result.Modules["java.net.URL"] = true
	result.Modules[fmt.Sprintf("java.net.%s", result.ConnectionClass())] = true
	result.Modules["java.net.MalformedURLException"] = true
	result.Modules["java.io.IOException"] = true
	result.Modules["java.io.BufferedReader"] = true
	result.Modules["java.io.InputStreamReader"] = true
	result.mimeCounter = 0
	result.formFileContentCounter = 0

	return result
}

//--- Getter methods called from template

func (self JavaGenerator) ConnectionClass() string {
	var targetUrl string
	if self.Options.Proxy != "" {
		targetUrl = self.Options.Proxy
	} else {
		targetUrl = self.Options.Url
	}
	if strings.HasPrefix(targetUrl, "https") {
		return "HttpsURLConnection"
	}
	return "HttpURLConnection"
}

func (self JavaGenerator) Proxy() string {
	if self.Options.Proxy == "" {
		return ""
	}
	u, err := url.Parse(self.Options.Proxy)
	if err != nil {
		log.Fatal(err)
	}
	hostFragments := strings.Split(u.Host, ":")
	host := hostFragments[0]
	var port string
	if len(hostFragments) == 1 {
		port = "80"
	} else {
		port = hostFragments[1]
	}
	return fmt.Sprintf(`new Proxy(Proxy.Type.%s, new InetSocketAddress("%s", %s))`, strings.ToUpper(u.Scheme), host, port)
}

func (self JavaGenerator) CommonInitialize() string {
	var buffer bytes.Buffer
	indent := func() {
		buffer.WriteString("            ")
	}

	for _, line := range self.commonInitialize {
		buffer.WriteString(line)
		buffer.WriteByte('\n')
		indent()
	}

	return buffer.String()
}

func (self JavaGenerator) PrepareConnection() string {
	var buffer bytes.Buffer
	indent := func() {
		buffer.WriteString("            ")
	}
	method := self.Options.Method()
	if method != "GET" {
		indent()
		buffer.WriteString(fmt.Sprintf("conn.setRequestMethod(\"%s\");\n", method))
	}
	for _, header := range self.Options.Header {
		indent()
		headers := strings.Split(header, ":")
		buffer.WriteString(fmt.Sprintf("conn.setRequestProperty(\"%s\", \"%s\");\n", strings.TrimSpace(headers[0]), strings.TrimSpace(headers[1])))
	}
	for _, header := range self.specialHeaders {
		indent()
		buffer.WriteString(fmt.Sprintf("conn.setRequestProperty(\"%s\", %s);\n", strings.TrimSpace(header[0]), header[1]))
	}
	if self.HasBody {
		indent()
		buffer.WriteString("conn.setRequestProperty(\"Content-Length\", String.valueOf(content.getBytes(\"UTF-8\").length));\n")
		indent()
		buffer.WriteString("conn.setDoOutput(true);\n")
		indent()
		buffer.WriteString("DataOutputStream wr = new DataOutputStream(conn.getOutputStream());\n")
		indent()
		buffer.WriteString("wr.writeBytes(content);\n")
		indent()
		buffer.WriteString("wr.flush();\n")
		indent()
		buffer.WriteString("wr.close();\n")
	}

	return buffer.String()
}

//--- Preparing Java source code methods

func (self *JavaGenerator) AppendCommonInitialize(newLine string, check bool) {
	if check {
		found := false
		for _, line := range self.commonInitialize {
			if line == newLine {
				found = true
				break
			}
		}
		if found {
			return
		}
	}
	self.commonInitialize = append(self.commonInitialize, newLine)
}

func (self *JavaGenerator) MimeTypeVariable() string {
	if self.Options.ProcessedData.ExternalFileCount() == 0 {
		return "mimeType"
	}
	self.mimeCounter = self.mimeCounter + 1
	return fmt.Sprintf("mimeType%d", self.mimeCounter)
}

func (self *JavaGenerator) FormFileContentVariable() string {
	if self.Options.ProcessedData.ExternalFileCount() == 0 {
		return "fileContent"
	}
	self.formFileContentCounter = self.formFileContentCounter + 1
	return fmt.Sprintf("fileContent%d", self.formFileContentCounter)
}

func (self *JavaGenerator) AddMultiPartCode() {
	self.AdditionalDeclaration = `
    static String BOUNDARY = "----------ThIs_Is_tHe_bouNdaRY_$";
    static String encodeMultiPartFormData(String[][] fields, String[][] files) {
        try {
            StringWriter writer = new StringWriter();
            char[] buffer = new char[1024];

            for (int i = 0; i < fields.length; i++) {
                String[] field = fields[i];
                writer.write("--");
                writer.write(BOUNDARY);
                writer.write("\r\n");
                writer.write("Content-Disposition: form-data; name=\"");
                writer.write(field[0]);
                writer.write("\"\r\n");
                if (!field[2].equals("")) {
                    writer.write("Content-Type: ");
                    writer.write(field[2]);
                    writer.write("\r\n");
                }
                writer.write("\r\n");
                writer.write(field[1]);
            }
            for (int i = 0; i < files.length; i++) {
                String[] file = files[i];
                writer.write("--");
                writer.write(BOUNDARY);
                writer.write("\r\n");
                writer.write("Content-Disposition: form-data; name=\"");
                writer.write(file[0]);
                writer.write("\"; filename=\"");
                writer.write(file[2]);
                writer.write("\"\r\nContent-Type: ");
                writer.write(file[3]);
                writer.write("\r\n\r\n");
                FileReader input = new FileReader(file[1]);

                for (int n = 0; -1 != (n = input.read(buffer));) {
                    writer.write(buffer, 0, n);
                }
            }
            writer.write("--");
            writer.write(BOUNDARY);
            writer.write("--\r\n\r\n");
            return writer.toString();
        } catch (IOException e) {
            e.printStackTrace();
            return "";
        }
    }
`
	boundary := "----------ThIs_Is_tHe_bouNdaRY_$"
	self.Options.InsertContentTypeHeader(fmt.Sprintf("multipart/form-data; boundary=%s", boundary))
	self.Modules["java.io.StringWriter"] = true
	self.Modules["java.io.FileReader"] = true
}

func (self *JavaGenerator) SetDataForUrl() {
	var buffer bytes.Buffer
	buffer.WriteString("StringWriter writer = new StringWriter();\n")
	indent := func() {
		buffer.WriteString("            ")
	}
	indent()
	fmt.Fprintf(&buffer, "writer.write(\"%s\");\n", self.Options.Url)
	indent()
	buffer.WriteString("writer.write('?');\n")
	indent()
	if self.Options.CanUseSimpleForm() {
		for i, data := range self.Options.ProcessedData {
			if i != 0 {
				indent()
				buffer.WriteString("writer.write('&');\n")
			}
			singleData, _ := url.ParseQuery(data.Value)
			for key, values := range singleData {
				fmt.Fprintf(&buffer, "writer.write(\"%s\");\n", key)
				indent()
				buffer.WriteString("writer.write('=');\n")
				indent()
				fmt.Fprintf(&buffer, "writer.write(\"%s\");\n", values[0])
				indent()
			}
		}
		self.PrepareBody = buffer.String()
	} else {
		if len(self.Options.ProcessedData) == 1 {
			prepareLines, forWriter := NewStringForData(self, &self.Options.ProcessedData[0])
			for _, line := range prepareLines {
				buffer.WriteString(line)
				buffer.WriteByte('\n')
				indent()
			}
			fmt.Fprintf(&buffer, "writer.write(%s);\n", forWriter)
			indent()
		} else {
			for i, data := range self.Options.ProcessedData {
				if i != 0 {
					indent()
					buffer.WriteString("writer.write('&');\n")
				}
				indent()
				prepareLines, forWriter := StringForData(self, &data)
				for _, line := range prepareLines {
					buffer.WriteString(line)
					buffer.WriteByte('\n')
					indent()
				}
				if forWriter != "" {
					fmt.Fprintf(&buffer, "writer.write(%s);\n")
				}
			}
			indent()
		}
	}
	self.Url = "writer.toString()"
	self.PrepareBody = buffer.String()
	self.Modules["java.io.StringWriter"] = true
}

func (self *JavaGenerator) SetDataForBody() {
	var buffer bytes.Buffer
	indent := func() {
		buffer.WriteString("            ")
	}
	if len(self.Options.ProcessedData) == 1 {
		prepareLines, forWriter := NewStringForData(self, &self.Options.ProcessedData[0])
		for _, line := range prepareLines {
			buffer.WriteString(line)
			buffer.WriteByte('\n')
			indent()
		}
		fmt.Fprintf(&buffer, "String content = %s;\n", forWriter)
		indent()
	} else {
		self.Modules["java.net.URLEncoder"] = true
		buffer.WriteString("StringWriter writer = new StringWriter();\n")
		for i, data := range self.Options.ProcessedData {
			if i != 0 {
				indent()
				buffer.WriteString("writer.write('&');\n")
			}
			prepareLines, forWriter := StringForData(self, &data)
			for _, line := range prepareLines {
				indent()
				buffer.WriteString(line)
				buffer.WriteByte('\n')
			}
			if forWriter != "" {
				indent()
				fmt.Fprintf(&buffer, "writer.write(%s);\n", forWriter)
			}
		}
		indent()
		buffer.WriteString("String content = writer.toString();\n")
		self.Modules["java.io.StringWriter"] = true
		indent()
	}
	self.PrepareBody = buffer.String()
	self.HasBody = true
}

func (self *JavaGenerator) SetDataForForm() {
	var buffer bytes.Buffer
	indent := func() {
		buffer.WriteString("            ")
	}

	buffer.WriteString("StringWriter writer = new StringWriter();\n")
	for i, data := range self.Options.ProcessedData {
		if i != 0 {
			indent()
			buffer.WriteString("writer.write('&');\n")
		}
		singleData, _ := url.ParseQuery(data.Value)
		for key, values := range singleData {
			fmt.Fprintf(&buffer, "writer.write(\"%s\");\n", key)
			buffer.WriteString("writer.write('=');\n")
			fmt.Fprintf(&buffer, "writer.write(\"%s\");\n", values[0])
		}
	}

	self.PrepareBody = buffer.String()
	self.HasBody = true
	self.Body = "values"
	self.Modules["java.net.URLEncoder"] = true
}

func (self *JavaGenerator) SetFormForBody() {
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
	indent := func() {
		buffer.WriteString("            ")
	}

	if len(fields) > 0 {
		buffer.WriteString("String[][] fields = {\n")
		for _, value := range fields {
			indent()
			buffer.WriteString(value)
		}
		indent()
		buffer.WriteString("};\n")
	} else {
		buffer.WriteString("String[][] fields = {};\n")
	}
	indent()
	if len(files) > 0 {
		buffer.WriteString("String[][] files = {\n")
		for _, value := range files {
			indent()
			buffer.WriteString(value)
		}
		indent()
		buffer.WriteString("};\n")
	} else {
		buffer.WriteString("String[][] files = {};\n")
	}
	indent()
	buffer.WriteString("String content = encodeMultiPartFormData(fields, files);\n")
	indent()
	self.PrepareBody = buffer.String()
	self.HasBody = true
}

/*
	Dispatcher function of curl command
	This is an exported function and called from httpgen.
*/
func ProcessCurlCommand(options *httpgen_common.CurlOptions) (string, interface{}) {
	generator := NewJavaGenerator(options)

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
		generator.specialHeaders = append(generator.specialHeaders, []string{"Authorization", fmt.Sprintf("\"Basic \" + Base64.getEncoder().encodeToString(\"%s\".getBytes(StandardCharsets.UTF_8))", generator.Options.User)})
		generator.Modules["java.util.Base64"] = true
		generator.Modules["java.nio.charset.StandardCharsets"] = true
	}
	if generator.HasBody {
		generator.Modules["java.io.DataOutputStream"] = true
	}

	return "full", *generator
}

// helper functions

func NewStringForData(generator *JavaGenerator, data *httpgen_common.DataOption) ([]string, string) {
	var result []string
	var resultForWriter string
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result, "StringWriter writer = new StringWriter();")
			result = append(result, fmt.Sprintf(`FileReader fileReader = new FileReader("%s");`, data.Value[1:]))
			result = append(result, "BufferedReader bufferedReader = new BufferedReader(fileReader);")
			result = append(result, "String str = bufferedReader.readLine();")
			result = append(result, "while (str != null) {")
			result = append(result, "	 writer.write(str);")
			result = append(result, "	 str = bufferedReader.readLine();")
			result = append(result, "}")
			result = append(result, "bufferedReader.close();")
			resultForWriter = "writer.toString()"
			generator.Modules["java.io.StringWriter"] = true
			generator.Modules["java.io.FileReader"] = true
		} else {
			resultForWriter = fmt.Sprintf(`"%s"`, strings.Replace(data.Value, "\n", "", -1))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result, "StringWriter writer = new StringWriter();")
			result = append(result, fmt.Sprintf(`FileReader fileReader = new FileReader("%s");`, data.Value[1:]))
			result = append(result, "for (int n = 0; -1 != (n = fileReader.read(buffer));) {")
			result = append(result, "    writer.write(buffer, 0, n);")
			result = append(result, "}")
			result = append(result, "fileReader.close();")
			resultForWriter = "writer.toString()"
			generator.AppendCommonInitialize("char[] buffer = new char[1024];", true)
			generator.Modules["java.io.StringWriter"] = true
			generator.Modules["java.io.FileReader"] = true
		} else {
			resultForWriter = fmt.Sprintf(`"%s"`, data.Value)
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result, "StringWriter writer = new StringWriter();")
			result = append(result, fmt.Sprintf(`FileReader fileReader = new FileReader("%s");`, data.Value[1:]))
			result = append(result, "BufferedReader bufferedReader = new BufferedReader(fileReader);")
			result = append(result, "String str = bufferedReader.readLine();")
			result = append(result, "while (str != null) {")
			result = append(result, "	 writer.write(URLEncoder.encode(str + \"\\n\", \"UTF-8\"));")
			result = append(result, "	 str = bufferedReader.readLine();")
			result = append(result, "}")
			result = append(result, "bufferedReader.close();")
			resultForWriter = "writer.toString()"
			generator.Modules["java.io.StringWriter"] = true
			generator.Modules["java.io.FileReader"] = true
		} else {
			resultForWriter = fmt.Sprintf("URLEncoder.encode(\"%s\", \"UTF-8\")", data.Value)
		}
		generator.Modules["java.net.URLEncoder"] = true
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, resultForWriter
}

func StringForData(generator *JavaGenerator, data *httpgen_common.DataOption) ([]string, string) {
	var result []string
	var resultForWriter string
	switch data.Type {
	case httpgen_common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result, "{")
			result = append(result, fmt.Sprintf(`    FileReader fileReader = new FileReader("%s");`, data.Value[1:]))
			result = append(result, "    BufferedReader bufferedReader = new BufferedReader(fileReader);")
			result = append(result, "    String str = bufferedReader.readLine();")
			result = append(result, "    while (str != null) {")
			result = append(result, "	     writer.write(str);")
			result = append(result, "	     str = bufferedReader.readLine();")
			result = append(result, "    }")
			result = append(result, "    bufferedReader.close();")
			result = append(result, "}")
			generator.Modules["java.io.FileReader"] = true
		} else {
			resultForWriter = fmt.Sprintf("\"%s\"", strings.Replace(data.Value, "\n", "", -1))
		}
	case httpgen_common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result, "{")
			result = append(result, fmt.Sprintf(`    FileReader fileReader = new FileReader("%s");`, data.Value[1:]))
			result = append(result, "    for (int n = 0; -1 != (n = fileReader.read(buffer));) {")
			result = append(result, "        writer.write(buffer, 0, n);")
			result = append(result, "    }")
			result = append(result, "    fileReader.close();")
			result = append(result, "}")
			generator.AppendCommonInitialize("char[] buffer = new char[1024];", true)
			generator.Modules["java.io.FileReader"] = true
			generator.Modules["java.io.BufferedReader"] = true
		} else {
			resultForWriter = fmt.Sprintf("\"%s\"", data.Value)
		}
	case httpgen_common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result, "{")
			result = append(result, fmt.Sprintf(`    FileReader fileReader = new FileReader("%s");`, data.Value[1:]))
			result = append(result, "    BufferedReader bufferedReader = new BufferedReader(fileReader);")
			result = append(result, "    String str = bufferedReader.readLine();")
			result = append(result, "    while (str != null) {")
			result = append(result, "	     writer.write(URLEncoder.encode(str + \"\\n\", \"UTF-8\"));")
			result = append(result, "	     str = bufferedReader.readLine();")
			result = append(result, "    }")
			result = append(result, "    bufferedReader.close();")
			result = append(result, "}")
			generator.Modules["java.io.FileReader"] = true
		} else {
			resultForWriter = fmt.Sprintf("URLEncoder.encode(\"%s\", \"UTF-8\")", data.Value)
		}
		generator.Modules["java.net.URLEncoder"] = true
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, resultForWriter
}

func FormString(generator *JavaGenerator, data *httpgen_common.DataOption) string {
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
			buffer.WriteString(fmt.Sprintf("    {\"%s\", \"%s\", ", field[0], fragments[0]))

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
			buffer.WriteString(fmt.Sprintf("\"%s\", ", sentFileName))

			// sent file name
			var mimeTypeVariable string
			if contentType != "" {
				buffer.WriteString(fmt.Sprintf("\"%s\"", contentType))
			} else {
				mimeTypeVariable = generator.MimeTypeVariable()
				buffer.WriteString(fmt.Sprintf("%s != null ? %s : \"application/octet-stream\"", mimeTypeVariable, mimeTypeVariable))
			}
			buffer.WriteString("},\n")
			result = buffer.String()
			generator.AppendCommonInitialize("FileNameMap fileNameMap = URLConnection.getFileNameMap();", true)
			if mimeTypeVariable != "" {
				generator.AppendCommonInitialize(fmt.Sprintf("String %s = fileNameMap.getContentTypeFor(\"%s\");", mimeTypeVariable, fragments[0]), false)
			}
			generator.Modules["java.net.URLConnection"] = true
			generator.Modules["java.net.FileNameMap"] = true
		} else if strings.HasPrefix(field[1], "<") {
			var buffer bytes.Buffer

			fragments := strings.Split(field[1][1:], ";")

			// field name, content
			generator.AppendCommonInitialize("char[] buffer = new char[1024];", true)
			contentVariable := generator.FormFileContentVariable()
			generator.AppendCommonInitialize(fmt.Sprintf("String %s;", contentVariable), false)
			generator.AppendCommonInitialize("{", false)
			generator.AppendCommonInitialize("    StringWriter writer = new StringWriter();", false)
			generator.AppendCommonInitialize(fmt.Sprintf(`    FileReader fileReader = new FileReader("%s");`, fragments[0]), false)
			generator.AppendCommonInitialize("    for (int n = 0; -1 != (n = fileReader.read(buffer));) {", false)
			generator.AppendCommonInitialize("        writer.write(buffer, 0, n);", false)
			generator.AppendCommonInitialize("    }", false)
			generator.AppendCommonInitialize("    fileReader.close();", false)
			generator.AppendCommonInitialize(fmt.Sprintf("    %s = writer.toString();", contentVariable), false)
			generator.AppendCommonInitialize("}", false)

			buffer.WriteString(fmt.Sprintf(`    {"%s", %s, `, field[0], contentVariable))

			var contentType string
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
				}
			}
			if contentType == "" {
				buffer.WriteString(`""`)
			} else {
				buffer.WriteString(fmt.Sprintf(`"%s"`, contentType))
			}
			buffer.WriteString("},\n")
			result = buffer.String()
		} else {
			result = fmt.Sprintf("    {\"%s\", \"%s\", \"\"},\n", field[0], field[1])
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
