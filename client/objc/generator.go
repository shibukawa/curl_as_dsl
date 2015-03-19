package objc

import (
	"bytes"
	"fmt"
	"github.com/shibukawa/curl_as_dsl/common"
	"log"
	"net/url"
	"os"
	"strings"
)

type ObjCGenerator struct {
	Options *common.CurlOptions

	Url                   string
	IsHttps               bool
	HasBody               bool
	Body                  string
	PrepareBody           string
	AdditionalDeclaration string
	specialHeaders        [][]string
	commonInitialize      []string
	Modules               map[string]bool
}

func NewObjCGenerator(options *common.CurlOptions) *ObjCGenerator {
	result := &ObjCGenerator{Options: options}
	result.Url = fmt.Sprintf(`@"%s"`, options.Url)
	result.Modules = make(map[string]bool)
	result.Modules["Foundation/Foundation.h"] = true

	return result
}

//--- Getter methods called from template

func (self ObjCGenerator) Proxy() string {
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

func (self ObjCGenerator) CommonInitialize() string {
	var buffer bytes.Buffer
	indent := func() {
		buffer.WriteString("        ")
	}

	for _, line := range self.commonInitialize {
		buffer.WriteString(line)
		buffer.WriteByte('\n')
		indent()
	}

	return buffer.String()
}

func (self ObjCGenerator) ModifyRequest() string {
	var buffer bytes.Buffer
	indent := func() {
		buffer.WriteString("        ")
	}
	method := self.Options.Method()
	if method != "GET" {
		indent()
		buffer.WriteString(fmt.Sprintf("[request setHTTPMethod:@\"%s\"];\n", method))
	}
	for _, headerStr := range self.Options.Header {
		indent()
		header := strings.Split(headerStr, ":")
		buffer.WriteString(fmt.Sprintf("[request setValue:@\"%s\" forHTTPHeaderField:@\"%s\"];\n", strings.TrimSpace(header[1]), strings.TrimSpace(header[0])))
	}
	for _, header := range self.specialHeaders {
		indent()
		buffer.WriteString(fmt.Sprintf("[request setValue:%s forHTTPHeaderField:@\"%s\"];\n", header[1], header[0]))
	}
	if self.HasBody {
		indent()
		buffer.WriteString("[request setValue:[NSString stringWithFormat:@\"%lu\", [content length]] forHTTPHeaderField:@\"Content-length\"];\n")
		indent()
		buffer.WriteString("[request setHTTPBody:content];\n")
	}

	return buffer.String()
}

//--- Preparing ObjC source code methods

func (self *ObjCGenerator) AppendCommonInitialize(newLine string, check bool) {
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

func (self *ObjCGenerator) AddMultiPartCode() {
	/*
		Thank you for following questions and answers!
		http://stackoverflow.com/questions/24250475/post-multipart-form-data-with-objective-c
		http://stackoverflow.com/questions/300618/in-memory-mime-type-detection-with-cocoa-os-x
	*/
	self.AdditionalDeclaration = `
NSData* encodeMultiPartBody(NSString* boundary, NSArray* fields, NSArray* files)
{
    NSMutableData *httpBody = [NSMutableData data];

    for (NSArray* field in fields) {
        NSString* key = [field objectAtIndex:0];
        NSString* value = [field objectAtIndex:1];
        [httpBody appendData:[[NSString stringWithFormat:@"--%@\r\n", boundary] dataUsingEncoding:NSUTF8StringEncoding]];
        [httpBody appendData:[[NSString stringWithFormat:@"Content-Disposition: form-data; name=\"%@\"\r\n", key] dataUsingEncoding:NSUTF8StringEncoding]];
        if ([field count] == 3) {
            [httpBody appendData:[[NSString stringWithFormat:@"Content-Type: %@\r\n", [field objectAtIndex:2]] dataUsingEncoding:NSUTF8StringEncoding]];
        }
        [httpBody appendData:[@"\r\n" dataUsingEncoding:NSUTF8StringEncoding]];
        [httpBody appendData:[[NSString stringWithFormat:@"%@\r\n", value] dataUsingEncoding:NSUTF8StringEncoding]];
    }

    for (NSArray *file in files) {
        NSString* key            = [file objectAtIndex:0];
        NSData*   data           = [NSData dataWithContentsOfFile:[file objectAtIndex:1]];
        NSString* remoteFileName = [file objectAtIndex:2];
        NSString* contentType    = [file objectAtIndex:3];

        [httpBody appendData:[[NSString stringWithFormat:@"--%@\r\n", boundary] dataUsingEncoding:NSUTF8StringEncoding]];
        [httpBody appendData:[[NSString stringWithFormat:@"Content-Disposition: form-data; name=\"%@\"; filename=\"%@\"\r\n", key, remoteFileName] dataUsingEncoding:NSUTF8StringEncoding]];
        [httpBody appendData:[[NSString stringWithFormat:@"Content-Type: %@\r\n\r\n", contentType] dataUsingEncoding:NSUTF8StringEncoding]];
        [httpBody appendData:data];
        [httpBody appendData:[@"\r\n" dataUsingEncoding:NSUTF8StringEncoding]];
    }

    [httpBody appendData:[[NSString stringWithFormat:@"--%@--\r\n", boundary] dataUsingEncoding:NSUTF8StringEncoding]];

    return httpBody;
}

NSString* getMimeTypeFromPath(NSString* path)
{
    NSString *mimeType = nil;
#ifdef TARGET_OS_MAC
    CFStringRef uti = (__bridge CFStringRef)[[NSWorkspace sharedWorkspace] typeOfFile:path error:nil];
#else
    CFStringRef extension = (__bridge CFStringRef)[path pathExtension];
    CFStringRef uti = UTTypeCreatePreferredIdentifierForTag(kUTTagClassFilenameExtension, extension, NULL);
    CFRelease(extension);
#endif
    if (uti) {
        CFStringRef cfMimeType = UTTypeCopyPreferredTagWithClass(uti, kUTTagClassMIMEType);
        if (cfMimeType) {
            mimeType = (__bridge NSString*)cfMimeType;
            CFRelease(cfMimeType);
        }
    }
    if (!mimeType) {
        mimeType = @"application/octet-stream";
    }
    return mimeType;
}
`
	self.specialHeaders = append(self.specialHeaders, []string{"Content-type", `[NSString stringWithFormat: @"multipart/form-data; boundary=%@", boundary]`})
}

func (self *ObjCGenerator) SetDataForUrl() {
	var buffer bytes.Buffer
	indent := func() {
		buffer.WriteString("        ")
	}
	if self.Options.CanUseSimpleForm() {
		fmt.Fprintf(&buffer, "NSMutableString* url = [@\"%s\" mutableCopy];\n", self.Options.Url)
		indent()
		buffer.WriteString("[url appendString:@\"?\"];\n")
		indent()
		for i, data := range self.Options.ProcessedData {
			if i != 0 {
				indent()
				buffer.WriteString("[url appendString:@\"&\"];\n")
			}
			singleData, _ := url.ParseQuery(data.Value)
			for key, values := range singleData {
				fmt.Fprintf(&buffer, "[url appendFormat:@\"%%@=%%@\", @\"%s\", @\"%s\"];\n", key, values[0])
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
			fmt.Fprintf(&buffer, "NSString* url = [[NSString alloc] initWithFormat:@\"%%@?%%@\", @\"%s\", %s];\n", self.Options.Url, forWriter)
			indent()
		} else {
			fmt.Fprintf(&buffer, "NSMutableString* query = [@\"%s\" mutableCopy];\n", self.Options.Url)
			indent()
			buffer.WriteString("[url appendString:@\"?\"];\n")
			indent()
			for i, data := range self.Options.ProcessedData {
				if i != 0 {
					indent()
					buffer.WriteString("[url appendString:@\"&\"];\n")
				}
				indent()
				prepareLines, forWriter := StringForData(self, &data)
				for _, line := range prepareLines {
					buffer.WriteString(line)
					buffer.WriteByte('\n')
					indent()
				}
				if forWriter != "" {
					fmt.Fprintf(&buffer, "[query appendString:%s];\n", forWriter)
				}
			}
			indent()
		}
	}
	self.Url = "url"
	self.PrepareBody = buffer.String()
}

func (self *ObjCGenerator) SetDataForBody() {
	var buffer bytes.Buffer
	indent := func() {
		buffer.WriteString("        ")
	}
	if len(self.Options.ProcessedData) == 1 {
		prepareLines, forWriter := NewBinaryForData(self, &self.Options.ProcessedData[0])
		for _, line := range prepareLines {
			buffer.WriteString(line)
			buffer.WriteByte('\n')
			indent()
		}
		if forWriter != "" {
			fmt.Fprintf(&buffer, "NSData *content = %s;\n", forWriter)
			indent()
		}
	} else {
		buffer.WriteString("NSMutableData* content = [NSMutableData data];\n")
		for i, data := range self.Options.ProcessedData {
			if i != 0 {
				indent()
				buffer.WriteString("[content appendBytes:\"&\" length:1];\n")
			}
			prepareLines, forWriter := BinaryForData(self, &data)
			for _, line := range prepareLines {
				indent()
				buffer.WriteString(line)
				buffer.WriteByte('\n')
			}
			if forWriter != "" {
				indent()
				fmt.Fprintf(&buffer, "[content appendData:%s];\n", forWriter)
			}
		}
		indent()
	}
	self.PrepareBody = buffer.String()
	self.HasBody = true
}

func (self *ObjCGenerator) SetDataForForm() {
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
}

func (self *ObjCGenerator) SetFormForBody() {
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
		buffer.WriteString("        ")
	}

	if len(fields) > 0 {
		buffer.WriteString("NSArray* fields = [NSArray arrayWithObjects:\n")
		for _, value := range fields {
			indent()
			buffer.WriteString(value)
		}
		indent()
		buffer.WriteString("    nil\n")
		indent()
		buffer.WriteString("];\n")
	} else {
		buffer.WriteString("NSArray* fields = [NSArray array];\n")
	}
	indent()
	if len(files) > 0 {
		buffer.WriteString("NSArray* files = [NSArray arrayWithObjects:\n")
		for _, value := range files {
			indent()
			buffer.WriteString(value)
		}
		indent()
		buffer.WriteString("    nil\n")
		indent()
		buffer.WriteString("];\n")
	} else {
		buffer.WriteString("NSArray* files = [NSArray array];\n")
	}
	indent()
	buffer.WriteString("NSString *boundary = [NSString stringWithFormat:@\"Boundary-%@\", [[NSUUID UUID] UUIDString]];\n")
	indent()
	buffer.WriteString("NSData *content = encodeMultiPartBody(boundary, fields, files);\n")
	indent()
	self.PrepareBody = buffer.String()
	self.HasBody = true
	self.Modules["AppKit/NSWorkspace.h"] = true
}

/*
	Dispatcher function of curl command
	This is an exported function and called from common.
*/
func ProcessCurlCommand(options *common.CurlOptions) (string, interface{}) {
	generator := NewObjCGenerator(options)

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
		generator.specialHeaders = append(generator.specialHeaders,
			[]string{
				"Authorization",
				fmt.Sprintf(`[NSString stringWithFormat:@"Basic %%@", [[@"%s"dataUsingEncoding:NSUTF8StringEncoding] base64EncodedStringWithOptions:NSDataBase64EncodingEndLineWithLineFeed]]`, generator.Options.User)})
	}
	if generator.HasBody {
	}

	return "full", *generator
}

// helper functions

func NewBinaryForData(generator *ObjCGenerator, data *common.DataOption) ([]string, string) {
	var result []string
	var resultForWriter string
	switch data.Type {
	case common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result,
				`NSMutableData* content = [NSMutableData data];`,
				fmt.Sprintf(`NSString *fileContents = [NSString stringWithContentsOfFile:@"%s" encoding:NSUTF8StringEncoding error:NULL];`, data.Value[1:]),
				`for (NSString *line in [fileContents componentsSeparatedByCharactersInSet:[NSCharacterSet newlineCharacterSet]]) {`,
				`    [content appendData:[line dataUsingEncoding:NSUTF8StringEncoding]];`,
				`}`)
		} else {
			resultForWriter = fmt.Sprintf(`[@"%s" dataUsingEncoding:NSUTF8StringEncoding]`, strings.Replace(data.Value, "\n", "", -1))
		}
	case common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			resultForWriter = fmt.Sprintf(`[NSData dataWithContentsOfFile:@"%s"]`, data.Value[1:])
		} else {
			resultForWriter = fmt.Sprintf(`@"%s"`, data.Value)
		}
	case common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result,
				`NSError *error = nil;`,
				fmt.Sprintf(`NSString *source = [NSString stringWithContentsOfFile:@"%s" encoding: NSUTF8StringEncoding error:&error];`, data.Value[1:]),
				`NSString *encodedSource = [source stringByAddingPercentEncodingWithAllowedCharacters:[NSCharacterSet URLQueryAllowedCharacterSet]];`)
			resultForWriter = "[encodedSource dataUsingEncoding:NSUTF8StringEncoding]"
		} else {
			resultForWriter = fmt.Sprintf(`[[@"%s" stringByAddingPercentEncodingWithAllowedCharacters:[NSCharacterSet URLQueryAllowedCharacterSet]] dataUsingEncoding:NSUTF8StringEncoding]`, data.Value)
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, resultForWriter
}

func BinaryForData(generator *ObjCGenerator, data *common.DataOption) ([]string, string) {
	var result []string
	var resultForWriter string
	switch data.Type {
	case common.DataAsciiType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result,
				"{",
				fmt.Sprintf(`    NSString *fileContents = [NSString stringWithContentsOfFile:@"%s" encoding:NSUTF8StringEncoding error:NULL];`, data.Value[1:]),
				`    for (NSString *line in [fileContents componentsSeparatedByCharactersInSet:[NSCharacterSet newlineCharacterSet]]) {`,
				`        [content appendData:[line dataUsingEncoding:NSUTF8StringEncoding]];`,
				`    }`,
				`}`)
		} else {
			resultForWriter = fmt.Sprintf(`[@"%s" dataUsingEncoding:NSUTF8StringEncoding]`, strings.Replace(data.Value, "\n", "", -1))
		}
	case common.DataBinaryType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result, fmt.Sprintf(`[content appendData:[NSData dataWithContentsOfFile:@"%s"]];`, data.Value[1:]))
		} else {
			resultForWriter = fmt.Sprintf(`[@"%s" dataUsingEncoding:NSUTF8StringEncoding]`, data.Value)
		}
	case common.DataUrlEncodeType:
		if strings.HasPrefix(data.Value, "@") {
			result = append(result,
				"{",
				`    NSError *error = nil;`,
				fmt.Sprintf(`    NSString *source = [NSString stringWithContentsOfFile:@"%s" encoding: NSUTF8StringEncoding error:&error];`, data.Value[1:]),
				`    NSString *encodedSource = [source stringByAddingPercentEncodingWithAllowedCharacters:[NSCharacterSet URLQueryAllowedCharacterSet]];`,
				`    [content appendData:[encodedSource dataUsingEncoding:NSUTF8StringEncoding]];`,
				"}")
		} else {
			resultForWriter = fmt.Sprintf("URLEncoder.encode(\"%s\", \"UTF-8\")", data.Value)
		}
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, resultForWriter
}

func NewStringForData(generator *ObjCGenerator, data *common.DataOption) ([]string, string) {
	var result []string
	var resultForWriter string
	switch data.Type {
	case common.DataAsciiType:
		resultForWriter = fmt.Sprintf(`@"%s"`, strings.Replace(data.Value, "\n", "", -1))
	case common.DataBinaryType:
		resultForWriter = fmt.Sprintf(`@"%s"`, data.Value)
	case common.DataUrlEncodeType:
		resultForWriter = fmt.Sprintf(`[[@"%s"
		stringByAddingPercentEncodingWithAllowedCharacters:[NSCharacterSet URLQueryAllowedCharacterSet]] dataUsingEncoding:NSUTF8StringEncoding]`, data.Value)
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, resultForWriter
}

func StringForData(generator *ObjCGenerator, data *common.DataOption) ([]string, string) {
	var result []string
	var resultForWriter string
	switch data.Type {
	case common.DataAsciiType:
		resultForWriter = fmt.Sprintf(`[@"%s" dataUsingEncoding:NSUTF8StringEncoding]`, strings.Replace(data.Value, "\n", "", -1))
	case common.DataBinaryType:
		resultForWriter = fmt.Sprintf(`[@"%s" dataUsingEncoding:NSUTF8StringEncoding]`, data.Value)
	case common.DataUrlEncodeType:
		resultForWriter = fmt.Sprintf("URLEncoder.encode(\"%s\", \"UTF-8\")", data.Value)
	default:
		panic(fmt.Sprintf("unknown type: %d", data.Type))
	}
	return result, resultForWriter
}

func FormString(generator *ObjCGenerator, data *common.DataOption) string {
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
			fragments := strings.Split(field[1][1:], ";")

			// field name, source file name
			fmt.Fprintf(&buffer, "    [NSArray arrayWithObjects:@\"%s\", @\"%s\", ", field[0], fragments[0])

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
			fmt.Fprintf(&buffer, "@\"%s\", ", sentFileName)

			// sent file name
			if contentType != "" {
				fmt.Fprintf(&buffer, "@\"%s\"", contentType)
			} else {
				fmt.Fprintf(&buffer, "getMimeTypeFromPath(@\"%s\")", fragments[0])
			}
			buffer.WriteString(", nil],\n")
			result = buffer.String()
		} else if strings.HasPrefix(field[1], "<") {
			var buffer bytes.Buffer

			fragments := strings.Split(field[1][1:], ";")

			// field name, content
			fmt.Fprintf(&buffer, "    [NSArray arrayWithObjects:@\"%s\", [NSString stringWithContentsOfFile:@\"%s\" encoding:NSUTF8StringEncoding error: nil], ", field[0], fragments[0])

			var contentType string
			for _, fragment := range fragments[1:] {
				if strings.HasPrefix(fragment, "type=") {
					contentType = fragment[5:]
				}
			}
			if contentType != "" {
				fmt.Fprintf(&buffer, `@"%s", `, contentType)
			}
			buffer.WriteString("nil],\n")
			result = buffer.String()
		} else {
			result = fmt.Sprintf("    [NSArray arrayWithObjects:@\"%s\", @\"%s\", nil],\n", field[0], field[1])
		}
	case common.FormStringType:
		field := strings.SplitN(data.Value, "=", 2)
		if len(field) != 2 {
			fmt.Fprintln(os.Stderr, "Warning: Illegally formatted input field!\ncurl: option -F: is badly used here")
			os.Exit(1)
		}
		result = fmt.Sprintf("    [NSArray arrayWithObjects:@\"%s\", @\"%s\", nil],\n", field[0], field[1])
	}
	return result
}
