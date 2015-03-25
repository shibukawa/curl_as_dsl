package form2curl

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/shibukawa/shell"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"log"
	"math"
	"net/url"
	"strconv"
	"strings"
)

var acceptableOtherInputTypes map[string]bool = map[string]bool{
	"color":          true,
	"date":           true,
	"datetime":       true,
	"datetime-local": true,
	"email":          true,
	"hidden":         true,
	"month":          true,
	"number":         true,
	"password":       true,
	"search":         true,
	"tel":            true,
	"time":           true,
	"url":            true,
	"week":           true,
}

var ignoreInputTypes map[string]bool = map[string]bool{
	"button": true,
	"image":  true,
	"reset":  true,
	"submit": true,
}

func CreateFormFromReader(reader io.Reader) (*Form, error) {
	rootNode, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}
	nodes := findNodes(rootNode, []atom.Atom{atom.Form})
	if len(nodes) > 0 {
		return NewForm(nodes[0]), nil
	}
	return nil, errors.New("Form is not in html source.")
}

func CreateFormFromString(source string) (*Form, error) {
	var buffer bytes.Buffer
	buffer.WriteString(source)
	return CreateFormFromReader(&buffer)
}

func NewForm(formNode *html.Node) *Form {
	result := &Form{
		Node:     formNode,
		Method:   "GET",
		DataList: make(map[string][]string),
	}
	result.readFormParameter()
	result.readInputNodes()
	return result
}

func findNodes(rootNode *html.Node, stopTypes []atom.Atom) []*html.Node {
	stopTypeMap := make(map[atom.Atom]bool)
	for _, stopType := range stopTypes {
		stopTypeMap[stopType] = true
	}
	var result []*html.Node
	nodes := []*html.Node{rootNode}

	// breadth first search
	for {
		if len(nodes) == 0 {
			break
		}
		var nextNodes []*html.Node
		for _, node := range nodes {
			childNode := node.FirstChild
			if childNode == nil {
				continue
			}
			for {
				if childNode.Type == html.ElementNode {
					if stopTypeMap[childNode.DataAtom] {
						result = append(result, childNode)
					} else {
						nextNodes = append(nextNodes, childNode)
					}
				}
				childNode = childNode.NextSibling
				if childNode == nil {
					break
				}
			}
		}
		nodes = nextNodes
	}
	return result
}

type InputElement struct {
	Node             *html.Node
	Type             string // text, radio, select, textarea
	Value            string
	ValueWhenChecked string // radio
	Name             string
	DataListName     string
	Selections       []string // for radio, select, datalist
	Selected         bool     // for radio
}

func (self *InputElement) ValueString(useDummy bool) string {
	var result string

	// already have default value
	if self.Value != "" {
		if self.Type == "file" {
			result = fmt.Sprintf("@%s", self.Value)
		} else {
			result = self.Value
		}
	} else if useDummy {
		switch self.Type {
		case "text":
			if len(self.Selections) > 0 {
				result = self.Selections[0]
			} else {
				result = "dummy text"
			}
		case "select":
			if len(self.Selections) > 0 {
				result = self.Selections[0]
			}
		case "radio":
			if len(self.Selections) > 0 {
				result = self.Selections[0]
			}
		case "checkbox":
			result = self.ValueWhenChecked
		case "file":
			result = "@test.txt"
		case "color":
			result = "#0592D0" // I am a member of Resistance
		case "date":
			result = "2015-03-23"
		case "month":
			result = "2015-03"
		case "tel":
			result = "+1-202-555-0145"
		case "time":
			result = "12:30"
		case "week":
			result = "2015-W04"
		case "number":
			if self.Selections[0] != "" {
				if self.Selections[1] != "" {
					minValue, _ := strconv.ParseInt(self.Selections[0], 10, 32)
					maxValue, _ := strconv.ParseInt(self.Selections[1], 10, 32)
					result = strconv.Itoa(int((minValue + maxValue) / 2))
				} else {
					result = self.Selections[0]
				}
			} else {
				if self.Selections[1] != "" {
					result = self.Selections[1]
				} else {
					result = "1234567"
				}
			}
		}
	}
	return result
}

/*
	It returns the content should be removed when the value is empty
*/
func (self *InputElement) CanOmit() bool {
	return self.Type == "checkbox"
}

type Form struct {
	Node          *html.Node
	Inputs        []*InputElement
	Method        string
	Action        string
	EncType       string
	AcceptCharset string
	Accept        string
	DataList      map[string][]string
	Warnings      []string
}

func getAttributes(node *html.Node) map[string][]string {
	result := make(map[string][]string)
	for _, attribute := range node.Attr {
		result[attribute.Key] = append(result[attribute.Key], attribute.Val)
	}
	return result
}

func getTagName(node *html.Node) string {
	orphanNode := &html.Node{
		Type:      node.Type,
		DataAtom:  node.DataAtom,
		Data:      node.Data,
		Namespace: node.Namespace,
		Attr:      node.Attr,
	}
	var buffer bytes.Buffer
	html.Render(&buffer, orphanNode)
	return buffer.String()
}

func getText(rootNode *html.Node) string {
	var buffer bytes.Buffer
	nodes := []*html.Node{rootNode}

	// breadth first search
	for {
		if len(nodes) == 0 {
			break
		}
		var nextNodes []*html.Node
		for _, node := range nodes {
			childNode := node.FirstChild
			if childNode == nil {
				continue
			}
			for {
				if childNode.Type == html.ElementNode {
					nextNodes = append(nextNodes, childNode)
				} else if childNode.Type == html.TextNode {
					buffer.WriteString(childNode.Data)
				}
				childNode = childNode.NextSibling
				if childNode == nil {
					break
				}
			}
		}
		nodes = nextNodes
	}
	return buffer.String()
}

func (self *Form) selectAttribute(node *html.Node, attributeName string, attributes map[string][]string, dumpWarning bool) string {
	names := attributes[attributeName]
	switch len(names) {
	case 0:
		if dumpWarning {
			self.warning("%s doesn't have any %s attribute. This tag will be ignored.", getTagName(node), attributeName)
		}
		return ""
	case 1:
		return names[0]
	default:
		if dumpWarning {
			self.warning("%s has multiple %s attributes. Remove duplication.", getTagName(node), attributeName, attributeName)
		}
		return ""
	}
}

func (self *Form) getOptionValues(node *html.Node) ([]string, string) {
	var values []string
	var activeValue string
	for _, option := range findNodes(node, []atom.Atom{atom.Option}) {
		value := self.selectAttribute(option, "value", getAttributes(option), true)
		if value != "" {
			values = append(values, value)
		}
		if self.selectAttribute(option, "selected", getAttributes(option), false) != "" {
			activeValue = value
		}
	}
	return values, activeValue
}

func (self *Form) warning(format string, a ...interface{}) {
	self.Warnings = append(self.Warnings, fmt.Sprintf(format, a...))
}

func (self *Form) readFormParameter() {
	for _, attribute := range self.Node.Attr {
		switch attribute.Key {
		case "method":
			method := strings.ToUpper(attribute.Val)
			if method == "GET" || method == "POST" {
				self.Method = method
			} else {
				self.warning("<form> can accept only GET or POST as method, but '%s' is specified.", method)
			}
		case "action":
			action, err := url.Parse(attribute.Val)
			if err != nil {
				self.warning("<form> action '%s' is invalid: %s", attribute.Val, err.Error())
			} else if action.Scheme != "" && action.Scheme != "http" && action.Scheme != "https" {
				self.warning("<form> action '%s' should be http or https or path style value", attribute.Val)
			} else {
				self.Action = attribute.Val
			}
		case "enctype":
			switch attribute.Val {
			case "application/x-www-form-urlencoded":
				self.EncType = attribute.Val
			case "multipart/form-data":
				self.EncType = attribute.Val
			case "text/plain":
				self.warning("This tool doesn't support <form> enctype text/plain")
			default:
				self.warning("<form> enctype '%s' is invalid. you can use the following values: application/x-www-form-urlencoded, multipart/form-data.", attribute.Val)
			}
		case "accept-charset":
			self.AcceptCharset = attribute.Val
		case "accept":
			self.Accept = attribute.Val
		}
	}
	if self.EncType != "" && self.Method == "GET" {
		self.warning("<form> enctype attribute is only for POST form")
		self.EncType = ""
	}
}

func (self *Form) readInputNodes() {
	nodes := findNodes(self.Node, []atom.Atom{atom.Input, atom.Textarea, atom.Datalist, atom.Select})
	radioNodes := make(map[string][]*InputElement)

	// Process Input elements
	for _, node := range nodes {
		attributes := getAttributes(node)
		switch node.DataAtom {
		case atom.Input:
			typeName := strings.ToLower(self.selectAttribute(node, "type", attributes, false))
			if ignoreInputTypes[typeName] {
				break
			}
			name := self.selectAttribute(node, "name", attributes, true)
			if name == "" {
				break
			}
			if typeName == "" {
				typeName = "text"
			}
			input := &InputElement{
				Node: node,
				Type: typeName,
				Name: name,
			}
			switch typeName {
			case "text":
				input.Value = self.selectAttribute(node, "value", attributes, false)
				input.DataListName = self.selectAttribute(node, "list", attributes, false)
				self.Inputs = append(self.Inputs, input)
			case "radio":
				input.Value = self.selectAttribute(node, "value", attributes, false)
				input.Selected = len(attributes["checked"]) > 0
				radioNodes[name] = append(radioNodes[name], input)
			case "checkbox":
				input.ValueWhenChecked = self.selectAttribute(node, "value", attributes, false)
				input.Selected = len(attributes["checked"]) > 0
				if input.Selected {
					input.Value = input.ValueWhenChecked
				}
				self.Inputs = append(self.Inputs, input)
			case "file":
				if self.EncType != "multipart/form-data" {
					self.warning("'multipart/form-data should be specified when <input type=file> tag exists, but '%s'", self.EncType)
				}
				if self.Method == "GET" {
					errorMessage := "HTML form can't send file content when 'GET' is specified. use 'POST' instead."
					self.Warnings = append(self.Warnings, errorMessage)
				}
				input.Value = self.selectAttribute(node, "accept", attributes, false)
				self.Inputs = append(self.Inputs, input)
			case "range":
				input.Value = self.selectAttribute(node, "value", attributes, false)
				if input.Value == "" {
					var minValue float64 = 0
					var maxValue float64 = 100
					minStr := self.selectAttribute(node, "min", attributes, false)
					if minStr != "" {
						tempMinValue, err := strconv.ParseFloat(minStr, 64)
						if err != nil {
							self.warning("<range>'s min attribute value '%s' is invalid for number.", minStr)
						} else {
							minValue = math.Floor(tempMinValue)
						}
					}
					maxStr := self.selectAttribute(node, "max", attributes, false)
					if maxStr != "" {
						tempMaxValue, err := strconv.ParseFloat(maxStr, 64)
						if err != nil {
							self.warning("<range>'s max attribute value '%s' is invalid for number.", minStr)
						} else {
							maxValue = math.Floor(tempMaxValue)
						}
					}
					if maxValue < minValue {
						input.Value = strconv.Itoa(int(minValue))
					} else {
						value := int(math.Ceil((minValue + maxValue) / 2))
						input.Value = strconv.Itoa(value)
					}
				} else {
					tempValue, err := strconv.ParseFloat(input.Value, 64)
					if err != nil {
						self.warning("<range>'s value attribute '%s' is invalid for number.", input.Value)
						input.Value = ""
					} else {
						input.Value = strconv.Itoa(int(tempValue))
					}
				}
				self.Inputs = append(self.Inputs, input)
			case "number":
				input.Value = self.selectAttribute(node, "value", attributes, false)
				minStr := self.selectAttribute(node, "min", attributes, false)
				if minStr != "" {
					tempMinValue, err := strconv.ParseInt(minStr, 10, 32)
					if err != nil {
						self.warning("<range>'s min attribute value '%s' is invalid for number.", minStr)
						minStr = ""
					} else {
						minStr = strconv.Itoa(int(tempMinValue))
					}
				}
				maxStr := self.selectAttribute(node, "max", attributes, false)
				if maxStr != "" {
					tempMaxValue, err := strconv.ParseInt(maxStr, 10, 32)
					if err != nil {
						self.warning("<range>'s max attribute value '%s' is invalid for number.", minStr)
						maxStr = ""
					} else {
						maxStr = strconv.Itoa(int(tempMaxValue))
					}
				}
				input.Selections = []string{minStr, maxStr}
				if input.Value != "" {
					tempValue, err := strconv.ParseInt(input.Value, 10, 32)
					if err != nil {
						self.warning("<range>'s value attribute '%s' is invalid for number.", input.Value)
						input.Value = ""
					} else {
						input.Value = strconv.Itoa(int(tempValue))
					}
				}
				self.Inputs = append(self.Inputs, input)
			default:
				if acceptableOtherInputTypes[typeName] {
					self.Inputs = append(self.Inputs, input)
				} else {
					self.warning("%s has unknown type '%s'.", getTagName(node), typeName)
				}
			}
		case atom.Textarea:
			name := self.selectAttribute(node, "name", attributes, true)
			if name == "" {
				break
			}
			self.Inputs = append(self.Inputs, &InputElement{
				Type:  "textarea",
				Name:  name,
				Value: getText(node),
			})
		case atom.Select:
			name := self.selectAttribute(node, "name", attributes, true)
			if name == "" {
				break
			}
			values, activeValue := self.getOptionValues(node)
			input := &InputElement{
				Type: "select",
				Name: name,
			}
			input.Value = activeValue
			input.Selections = values
			self.Inputs = append(self.Inputs, input)
		case atom.Datalist:
			values, _ := self.getOptionValues(node)
			dataListId := self.selectAttribute(node, "id", attributes, true)
			if dataListId != "" {
				self.DataList[dataListId] = values
			}
		}
	}
	// Post process: combine radio boxes
	for name, siblingNodes := range radioNodes {
		var values []string
		var activeValue string
		for _, node := range siblingNodes {
			values = append(values, node.Value)
			if node.Selected {
				activeValue = node.Value
			}
		}
		input := &InputElement{
			Type:       "radio",
			Name:       name,
			Selections: values,
			Value:      activeValue,
		}
		self.Inputs = append(self.Inputs, input)
	}
	// Post process: search data list
	for _, input := range self.Inputs {
		if input.DataListName != "" {
			dataList, ok := self.DataList[input.DataListName]
			if !ok {
				self.warning("%s has unknown list name: %s", getTagName(input.Node), input.DataListName)
			} else {
				input.Selections = dataList
			}
		}
	}
}

func (self *Form) MakeCurlCommand() string {
	commands := []string{"curl"}
	log.Println("@1", commands)

	if self.Method == "GET" {
		commands = append(commands, "-G")
		log.Println("@2", commands)
		for _, param := range self.Inputs {
			value := param.ValueString(true)
			if value != "" {
				commands = append(commands, "--data-urlencode", shell.Escape(fmt.Sprintf(`%s=%s`, param.Name, value)))
				log.Println("@3", commands)
			} else if !param.CanOmit() {
				commands = append(commands, "--data-urlencode", shell.Escape(param.Name))
				log.Println("@4", commands)
			}
		}
	} else {
		for _, param := range self.Inputs {
			value := param.ValueString(true)
			if value != "" {
				commands = append(commands, "-F", shell.Escape(fmt.Sprintf(`%s=%s`, param.Name, value)))
				log.Println("@5", commands)
			} else if !param.CanOmit() {
				commands = append(commands, "-F", shell.Escape(fmt.Sprintf(`%s=`, param.Name)))
				log.Println("@6", commands)
			}
		}
	}
	if self.Accept != "" {
		commands = append(commands, "-H", shell.Escape(fmt.Sprintf(`Accept: %s`, self.Accept)))
		log.Println("@7", commands)
	}
	if self.AcceptCharset != "" {
		commands = append(commands, "-H", shell.Escape(fmt.Sprintf(`Accept-Charset: %s`, self.AcceptCharset)))
		log.Println("@8", commands)
	}
	if self.Action != "" {
		commands = append(commands, self.Action)
		log.Println("@9", commands)
	} else {
		commands = append(commands, "http://example.com")
		log.Println("@10", commands)
	}
	return strings.Join(commands, " ")
}
