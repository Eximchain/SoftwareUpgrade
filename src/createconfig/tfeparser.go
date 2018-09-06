package main

import "fmt"

// NewNodeData returns a pointer to a NodeData structure, initialized with the provided data
// parsing Terraform output, reference: https://github.com/hashicorp/terraform/blob/master/command/output.go
func NewNodeData(data []byte) (result *NodeData) {
	result = &NodeData{string(data), -1}
	return
}

func (nd *NodeData) skipsSpaces() {
	index := nd.index
	if index == -1 {
		index++
	}
	for index < len(nd.bytes) {
		ch := rune(nd.bytes[index])
		if ch == ' ' || ch == '\r' || ch == '\n' || ch == '\t' {
			index++
			continue
		}
		break
	}
	nd.index = index
}

func (nd *NodeData) skipOptionalChar() {
	index := nd.index
	for index < len(nd.bytes) {
		ch := rune(nd.bytes[index])
		if ch == '"' || ch == ',' {
			index++
			continue
		}
		break
	}
	nd.index = index
}

// ReadOptionalValue reads an optional token
func (nd *NodeData) ReadOptionalValue() (result string) {
	nd.skipsSpaces()
	for nd.index < len(nd.bytes) {
		ch := rune(nd.bytes[nd.index])
		if ch == ',' || ch == ']' || ch == '}' || ch == '\n' {
			break
		}
		result = result + string(ch)
		nd.index++
	}
	nd.skipsSpaces()
	return
}

func (nd *NodeData) eof() (result bool) {
	result = nd.index >= len(nd.bytes)
	return
}

// ReadValue reads a value from the specified NodeData
func (nd *NodeData) ReadValue() (result string) {
	nd.skipsSpaces()
	index := nd.index
	var name string

	for index < len(nd.bytes) {
		ch := nd.bytes[index]
		if ch != ' ' && ch != '\n' {
			name = name + string(ch)
			index++
		} else {
			break
		}
	}
	nd.index = index
	nd.skipsSpaces()
	result = name
	return
}

func (nd *NodeData) readLeftSquare() {
	nd.readChar('[')
}

func (nd *NodeData) readRightSquare() {
	nd.readChar(']')
}

func (nd *NodeData) readChar(chExpected rune) {
	nd.skipsSpaces()
	chSeen := rune(nd.bytes[nd.index])
	if chSeen == chExpected {
		nd.index++
	} else {
		panic(fmt.Sprintf(`"%s" expected, but found "%s"`, string(chExpected), string(chSeen)))
	}
	nd.skipsSpaces()
}

func (nd *NodeData) readEqual() {
	nd.readChar('=')
}

func (nd *NodeData) readLeftSquareOrLeftBrace() (result NodeType) {
	ch := nd.bytes[nd.index]
	switch ch {
	case '[':
		nd.index++
		result = ntList
	case '{':
		nd.index++
		result = ntMap
	}
	return
}

// ReadList reads an array of string. The List nomenclature is from Terraform
func (nd *NodeData) ReadList() (result []string) {
	nd.skipsSpaces()
readName:
	if nd.eof() {
		return
	}
	ch := rune(nd.bytes[nd.index])
	// empty list
	if ch == ']' { // no need to increment index
		return
	}
	nd.skipOptionalChar()
	value := nd.ReadOptionalValue()
	if value != "" {
		result = append(result, value)
	}
	nd.skipOptionalChar()
	goto readName
}

// ReadMap reads a dictionary from the given NodeData structure
func (nd *NodeData) ReadMap() (result map[string]interface{}) {
	result = make(map[string]interface{})
	values := []string{}
	nd.skipsSpaces()
readName:
	if nd.eof() {
		return
	}
	ch := rune(nd.bytes[nd.index])
	// empty map
	if ch == '}' || ch == ']' {
		return
	}
	name := nd.ReadValue()
	nd.readEqual()
	nd.readLeftSquare()
readvalue:
	value := nd.ReadOptionalValue()
	if value != "" {
		values = append(values, value)
	}
	result[name] = values
	ch = rune(nd.bytes[nd.index])
	if !(ch == ']') {
		goto readvalue
	}
	nd.readRightSquare()
	values = []string{}
	goto readName
}

// ReadRightBrace expects a right brace as the next char
func (nd *NodeData) ReadRightBrace() {
	nd.readChar('}')
}

// ReadRightSquare expects a right square as the next char
func (nd *NodeData) ReadRightSquare() {
	nd.readChar(']')
}

// ReadNode reads a node which consists of a sequence of a list, or a map, or a combination of both.
func (nd *NodeData) ReadNode() (result map[string]interface{}) {
	result = make(map[string]interface{})
	if !nd.eof() {
	loop:
		NodeName := nd.ReadValue()
		nd.readEqual()
		nodeType := nd.readLeftSquareOrLeftBrace()

		switch nodeType {
		case ntList:
			{
				value := nd.ReadList()
				nd.ReadRightSquare()
				amap := make(map[string]interface{})
				amap["value"] = value
				amap["type"] = "list"
				result[NodeName] = amap
			}
		case ntMap:
			{
				value := nd.ReadMap()
				nd.ReadRightBrace()
				amap := make(map[string]interface{})
				amap["value"] = value
				amap["type"] = "map"
				result[NodeName] = amap
			}
		}
		if !nd.eof() {
			goto loop
		}
	}

	return

}
