package main

import (
	"encoding/json"
	"fmt"
	"os"
	"softwareupgrade"
	"strings"
)

// parses terraform output
// Terraform output looks like this:
/*
given_name1 = {
  region1 = [hostname1 hostname2]
  region2 = [hostname3 hostname4]
  region3 = []
}
given_name2 = {
	region1 = []
	region2 = [hostname1 hostname2 hostname3]
}
*/

// Terrafourm output -json looks like this
/*

 */

type (
	// TerraformNode represents the output for a terraform output -json
	TerraformNode struct {
		Sensitive bool        `json:"sensitive"`
		Type      string      `json:"type"`  // could be map, list
		Value     interface{} `json:"value"` // could be []string, or map[string][]string
	}
)

func convertTerraformJSON(inputTemplate, jsonPath, outputFilename string) {
	var result string
	data, err := softwareupgrade.ReadDataFromFile(inputTemplate)
	if err == nil {
		result = string(data)
	}
	data, err = softwareupgrade.ReadDataFromFile(jsonPath)
	if err == nil {
		var TerraformOutput map[string]TerraformNode
		err = json.Unmarshal(data, &TerraformOutput)
		if err != nil {
			fmt.Printf("Error parsing Terraform JSON: %v\n", err)
			fmt.Println("Most probable cause of error is forgetting to add -json to terraform output")
			return
		}
		for k, v := range TerraformOutput {
			var (
				nodes []string
			)
			switch v.Type {
			case "map":
				{
					for _, v := range v.Value.(map[string]interface{}) {
						for _, node := range v.([]interface{}) {
							nodes = append(nodes, node.(string))
						}
					}
				}
			case "list":
				{
					list := v.Value.([]interface{})
					for i := range list {
						nodes = append(nodes, list[i].(string))
					}
				}
			}

			var nodeStr string
			for _, node := range nodes {
				quotedNode := fmt.Sprintf(`"%s"`, node)
				if nodeStr == "" {
					nodeStr = quotedNode
				} else {
					nodeStr = fmt.Sprintf("%s,\n%s", nodeStr, quotedNode)
				}
			}
			replacementTemplate := fmt.Sprintf("{%%%s}", k)
			result = strings.Replace(result, replacementTemplate, nodeStr, 1)
		}
		data := []byte(result)
		_, err = softwareupgrade.SaveDataToFile(outputFilename, data)
		if err == nil {
			fmt.Println("Terraform output conversion completed.")
		}
	}

}

func main() {
	fmt.Println()
	fmt.Println("Creates configuration file for Upgrading Quorum nodes")
	fmt.Println()
	if len(os.Args) < 4 {
		fmt.Println("Input 1 - Filename of template")
		fmt.Println("Input 2 - Terraform output in JSON format")
		fmt.Println("Input 3 - Filename to write output to")
		fmt.Println()
		return
	}

	var (
		templateFilename      string
		terraformJSONFilename string
		outputFilename        string
		err                   error
	)
	if templateFilename, err = softwareupgrade.Expand(os.Args[1]); err != nil {
		fmt.Printf(cExpandingFilenameErr, os.Args[1], err)
		fmt.Println()
		return
	}

	if terraformJSONFilename, err = softwareupgrade.Expand(os.Args[2]); err != nil {
		fmt.Printf(cExpandingFilenameErr, os.Args[2], err)
		fmt.Println()
		return
	}

	if outputFilename, err = softwareupgrade.Expand(os.Args[3]); err != nil {
		fmt.Printf(cExpandingFilenameErr, os.Args[3], err)
		fmt.Println()
		return
	}

	if !softwareupgrade.FileExists(templateFilename) {
		fmt.Printf(cFilenameDoesntExist, templateFilename)
		return
	}

	if !softwareupgrade.FileExists(terraformJSONFilename) {
		fmt.Printf(cFilenameDoesntExist, terraformJSONFilename)
		return
	}

	convertTerraformJSON(templateFilename, terraformJSONFilename, outputFilename)
}
