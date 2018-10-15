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

	removeQuotes := true
	removeDelimiters := true
	ConvertTerraformJSONContent(removeQuotes, removeDelimiters, inputTemplate, JSONContent, outputFilename)
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

func makeBootnodeValues() (result map[string][]string) {
	result = make(map[string][]string)
	result["ap-northeast-1"] = []string{}
	result["ap-northeast-2"] = []string{}
	result["ap-south-1"] = []string{"52.66.115.215", "13.232.221.120", "13.127.182.4", "52.66.114.148", "13.127.55.162"}
	result["ap-southeast-1"] = []string{}
	result["ap-southeast-2"] = []string{}
	result["ca-central-1"] = []string{}
	result["au-central-1"] = []string{}
	return
}

func makeBootnodeIPs() TerraformItem {
	return TerraformItem{Sensitive: false, Type: "map", Value: makeBootnodeValues()}
}

func makeBootnodeIPs2() (result TerraformItem) {
	result = makeTerraformItem("map")
	result.SetRegion("us-east-1")
	result.AddIP("ec2-34-229-88-28.compute-1.amazonaws.com")
	result.AddIP("ec2-18-209-160-89.compute-1.amazonaws.com")
	return
}

func makeVaultServerIPs() TerraformItem {
	return TerraformItem{Sensitive: false, Type: "list", Value: []string{"18.213.245.54"}}
}

func makeVaultServerIPs2() (result TerraformItem) {
	result = makeTerraformItem("list")
	result.AddIP("18.213.245.54")
	result.AddIP("165.21.100.88")
	return
}

func Test_Format(t *testing.T) {
	aTerraformType := makeTerraformType()
	item := aTerraformType.GetTerraformItem("bootnode_ips", "map")
	item.SetRegion("us-east-1")
	item.AddIP("192.168.0.1")
	item.SetRegion("us-east-2")
	item.AddIP("192.168.0.2")
	aTerraformType.SaveToFile("/tmp/terraform-AWS.json")
}

func Test_format2(t *testing.T) {
	var terraformOutput map[string]TerraformItem
	// terraformOutput := makeTerraformType()
	terraformOutput["bootnode_ips"] = makeBootnodeIPs2()
	terraformOutput["vault_server_ips"] = makeVaultServerIPs2()
	bytes, err := json.MarshalIndent(&terraformOutput, "", "    ")
	if err == nil {
		bytes = append(bytes, []byte("\n")...)
	}

	ioutil.WriteFile("/tmp/terraform-test-02.json", bytes, 0777)
}
