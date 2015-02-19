package httpgen_common

import (
	"os"
	"fmt"
	"strings"
)

type DataType int

const (
	DataAsciiType DataType = iota
	DataBinaryType
	DataUrlEncodeType
	FormType
	FormStringType
)

type DataOption struct {
	Value string
	Type DataType
}

func (self *DataOption) IsFormStyle() bool {
	return strings.IndexByte(self.Value, '=') != -1
}

func (self *DataOption) UseExternalFile() bool {
	if self.Type == FormType {
		index := strings.Index(self.Value, "=")
		if index == -1 {
			return false
		}
		if index < len(self.Value) - 1 {
			nextChar := self.Value[index + 1:index + 2]
			if nextChar == "@" || nextChar == "<" {
				return true
			}
		}
	} else if self.Type != FormStringType {
		return strings.HasPrefix(self.Value, "@")
	}
	return false
}

type DataOptions []DataOption

func (self *DataOptions) Append(data string, typeEmum DataType) {
	*self = append(*self, DataOption{Value:data, Type:typeEmum})
}

func (self *DataOptions) HasAnyData() bool {
	return len(*self) > 0
}

func (self *DataOptions) HasData() bool {
	for _, data := range *self {
		switch data.Type {
		case DataAsciiType:
			return true
		case DataBinaryType:
			return true
		case DataUrlEncodeType:
			return true
		}
	}
	return false
}

func (self *DataOptions) HasForm() bool {
	for _, data := range *self {
		switch data.Type {
		case FormType:
			return true
		case FormStringType:
			return true
		}
	}
	return false
}
type CurlOptions struct {
	// Example of verbosity with level
	Basic bool `long:"basic" description:"Use HTTP Basic Authentication (H)"`
    Compressed func() `long:"compressed" description:"Request compressed response (using deflate or gzip)"`
	Cookie string `short:"b" long:"cookie" value-name:"STRING/FILE" description:"Read cookies from STRING/FILE (H)"`
	CookieJar string `short:"c" long:"cookie-jar" value-name:"FILE" description:"Write cookies to FILE after operation (H)"`
	Data func(string) `short:"d" long:"data" value-name:"DATA" description:"HTTP POST data (H)"`
	DataAscii func(string) `long:"data-ascii" value-name:"DATA" description:"HTTP POST ASCII data (H)"`
	DataBinary func(string) `long:"data-binary" value-name:"DATA" description:"HTTP POST binary data (H)"`
	DataUrlEncode func(string) `long:"data-urlencode" value-name:"DATA" description:"HTTP POST data url encoded (H)"`
	Digest bool `long:"digest" description:"Use HTTP Digest Authentication (H)"`
	Get bool `short:"G" long:"get" description:"Send the -d data with a HTTP GET (H)"`
	Form func(string) `short:"F" long:"form" value-name:"KEY=VALUE" description:"Specify HTTP multipart POST data (H)"`
	FormString func(string) `long:"form-string" value-name:"KEY=VALUE" description:"Specify HTTP multipart POST data (H)"`
	Header []string `short:"H" long:"header" value-name:"LINE" description:"Pass custom header LINE to server (H)"`
	Head bool `short:"I" long:"head" description:"Show document info only"`
	//Http11 bool `long:"http1.1" description:"Use HTTP 1.1 (H)"`
	//Http2 bool `long:"http2" description:"Use HTTP 2 (H)"`
	Proxy string `short:"x" long:"proxy" value-name:"[PROTOCOL://]HOST[:PORT]" description:"Use proxy on given port"`
	Referer func(string) `short:"e" long:"referer" description:"Referer URL (H)"`
	Request string `short:"X" long:"request" value-name:"COMMAND" description:"Specify request command to use"`
	AWSV2 bool `long:"awsv2" description:"AWS V2 style authentication (original)"`
	TrEncoding func() `long:"tr-encoding" description:"Request compressed transfer encoding (H)"`
	Transfer func(string) `short:"T" long:"upload-file" value-name:"FILE" description:"Transfer FILE to destination"`
	Url string `long:"url" value-name:"URL" description:"URL to work with"`
	User string `short:"u" long:"user" value-name:"USER[:PASSWORD]" description:"Server user and password"`
	UserAgent func(string) `short:"A" long:"user-agent" value-name:"STRING" description:"User-Agent to send to server (H)"`
	ProcessedData DataOptions
}

func (self *CurlOptions) Init() {
	self.Compressed = func () {
		self.Header = append(self.Header, "Accept-Encoding: deflate")
		self.Header = append(self.Header, "Accept-Encoding: gzip")
	}

	self.Data = func (data string) {
		self.ProcessedData.Append(data, DataAsciiType)
	}
	self.DataAscii = self.Data;

	self.DataBinary = func (data string) {
		self.ProcessedData.Append(data, DataBinaryType)
	}

	self.DataUrlEncode = func (data string) {
		self.ProcessedData.Append(data, DataUrlEncodeType)
	}

	self.Form = func (data string) {
		self.ProcessedData.Append(data, FormType)
	}

	self.FormString = func (data string) {
		self.ProcessedData.Append(data, FormStringType)
	}

	self.Referer = func (data string) {
		self.Header = append(self.Header, fmt.Sprintf("Referer: %s", data))
	}

	self.Transfer = func (data string) {
		self.ProcessedData.Append(fmt.Sprintf("@%s", data), DataAsciiType)
		if self.Request == "" {
			self.Request = "PUT"
		}
	}

	self.TrEncoding = func () {
		self.Header = append(self.Header, "Te: gzip")
	}

	self.UserAgent = func (data string) {
		self.Header = append(self.Header, fmt.Sprintf("User-Agent: %s", data))
	}

}

func (self *CurlOptions) CheckError() error {
	if self.ProcessedData.HasData() && self.ProcessedData.HasData(){
		return fmt.Errorf("Warning: You can only select one HTTP request!")
	}
	return nil
}

func (self *CurlOptions) Method() string {
	method := strings.ToUpper(self.Request)
	// explicit method is the highest priority
	if method != "" {
		return method
	}
	if self.Get {
		return "GET"
	}
	if self.Head {
		return "HEAD"
	}
	if self.ProcessedData.HasAnyData() {
		return "POST"
	}
	return "GET"
}

func (self *CurlOptions) Headers() [][]string {
	result := make([][]string, 0)
	for _, header := range self.Header {
		words := strings.SplitN(header, ":", 2)
		if len(words) != 2 {
			fmt.Fprintln(os.Stderr, "[warning] %s is wrong style header.\n", header)
			continue
		}
		result = append(result, words)
	}
	return result
}

func (self *CurlOptions) OnlyHasContentTypeHeader() bool {
	headers := self.Headers()
	if len(headers) == 0 || (len(headers) == 1 && strings.ToLower(headers[0][0]) == "content-type") {
		return true
	}
	return false
}

