package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestConvertTerraformJSONContent(t *testing.T) {
	inputContent := "{%hello} {%there}\n"
	inputTemplateFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Errorf("Error in generating template file: %v", err)
		return
	}
	inputTemplate := inputTemplateFile.Name()
	defer os.Remove(inputTemplate)
	output, err := ioutil.TempFile("", "")
	if err != nil {
		t.Errorf("Error in generating output file: %v", err)
		return
	}
	outputFilename := output.Name()
	defer os.Remove(outputFilename)
	output.Close()
	err = ioutil.WriteFile(inputTemplate, []byte(inputContent), 0644)
	content := map[string]interface{}{}
	C1 := make(map[string]interface{})
	C2 := make(map[string]interface{})
	C1["type"] = "list"
	C1["value"] = []string{"world", "me"} // this is value of "hello"

	C2["type"] = "map"
	C2Map := make(map[string]interface{})
	C2Map["any1"] = []string{"nowhere", "here"} // this is the value of "there", part 1
	C2Map["any2"] = []string{"out", "there"}    // this is the value of "there", part 2
	C2["value"] = C2Map

	content["hello"] = C1
	content["there"] = C2

	JSONContent, err := json.Marshal(&content)
	fmt.Printf("%s", string(JSONContent))

	SetRemoveQuotes(true)
	SetRemoveDelimiters(true)
	ConvertTerraformJSONContent(inputTemplate, JSONContent, outputFilename)
	outputBytes, err := ioutil.ReadFile(outputFilename)
	if err != nil {
		t.Errorf("Failure reading output file %s due to %v", outputFilename, err)
	}
	outputContent := string(outputBytes)
	expectedContent := "world me nowhere here out there\n"
	if outputContent != expectedContent {
		t.Errorf("Output content: %s does not match expected content: %s", outputContent, expectedContent)
	}
}
