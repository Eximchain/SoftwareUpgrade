package main

import (
	"encoding/json"
	"flag"
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

// Terraform output -json looks like this
/*

{
    "vault_server_ips": {
        "sensitive": false,
        "type": "list",
        "value": [
            "107.23.75.56"
        ]
    }
}

*/

type (
	// TerraformNode represents the output for a terraform output -json
	TerraformNode struct {
		Sensitive bool        `json:"sensitive"`
		Type      string      `json:"type"`  // could be map, list
		Value     interface{} `json:"value"` // could be []string, or map[string][]string
	}
)

var (
	removeQuotes     bool
	removeDelimiters bool
)

// SetRemoveQuotes set the bool removeQuotes
func SetRemoveQuotes(value bool) {
	removeQuotes = value
}

// SetRemoveDelimiters set the bool removeDelimiters
func SetRemoveDelimiters(value bool) {
	removeDelimiters = value
}

// ConvertTerraformJSONContent reads the template from the given inputTemplate filename, combines
// it with the data from the JSONContent, and writes to the file specified by the outputFilename
func ConvertTerraformJSONContent(inputTemplate string, JSONContent []byte, outputFilename string) {
	var result string
	data, err := softwareupgrade.ReadDataFromFile(inputTemplate)
	if err != nil {
		fmt.Printf("Error reading from: %s due to %v\n", inputTemplate, err)
		return
	}
	result = string(data)
	data = JSONContent
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

		var nodeStr, nodeFormat, appendFormat string
		if removeQuotes || removeDelimiters {
			nodeFormat = "%s"
		} else {
			nodeFormat = `"%s"`
		}
		if removeDelimiters {
			appendFormat = "%s %s"
		} else {
			appendFormat = "%s,\n%s"
		}
		for _, node := range nodes {
			quotedNode := fmt.Sprintf(nodeFormat, node)
			if nodeStr == "" {
				nodeStr = quotedNode
			} else {
				nodeStr = fmt.Sprintf(appendFormat, nodeStr, quotedNode)
			}
		}
		replacementTemplate := fmt.Sprintf("{%%%s}", k)
		result = strings.Replace(result, replacementTemplate, nodeStr, 1)
	}
	data = []byte(result)
	_, err = softwareupgrade.SaveDataToFile(outputFilename, data)
	if err == nil {
		fmt.Println("Terraform output conversion completed.")
	} else {
		fmt.Printf("Error saving %s due to error: %v\n", outputFilename, err)
		fmt.Println("Aborting.")
	}

}

func convertTerraformJSONFile(inputTemplate, jsonPath, outputFilename string) {
	data, err := softwareupgrade.ReadDataFromFile(jsonPath)
	if err == nil {
		ConvertTerraformJSONContent(inputTemplate, data, outputFilename)
	} else {
		fmt.Printf("Unable to read from %s due to error: %v\n", jsonPath, err)
		fmt.Println("Aborting.")
	}
}

func convertTerraformJSON(inputTemplate, jsonPath, outputFilename string) {
	var result string
	data, err := softwareupgrade.ReadDataFromFile(inputTemplate)
	if err == nil {
		result = string(data)
	} else {
		fmt.Printf("Unable to read from %s due to error: %v\n", inputTemplate, err)
		fmt.Println("Aborting.")
		return
	}
	data, err = softwareupgrade.ReadDataFromFile(jsonPath)
	if err != nil {
		fmt.Printf("Unable to read from %s due to error: %v\n", jsonPath, err)
		fmt.Println("Aborting.")
		return
	}
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

		var nodeStr, nodeFormat, appendFormat string
		if removeQuotes || removeDelimiters {
			nodeFormat = "%s"
		} else {
			nodeFormat = `"%s"`
		}
		if removeDelimiters {
			appendFormat = "%s %s"
		} else {
			appendFormat = "%s,\n%s"
		}
		for _, node := range nodes {
			quotedNode := fmt.Sprintf(nodeFormat, node)
			if nodeStr == "" {
				nodeStr = quotedNode
			} else {
				nodeStr = fmt.Sprintf(appendFormat, nodeStr, quotedNode)
			}
		}
		replacementTemplate := fmt.Sprintf("{%%%s}", k)
		result = strings.Replace(result, replacementTemplate, nodeStr, 1)
	}
	data = []byte(result)
	_, err = softwareupgrade.SaveDataToFile(outputFilename, data)
	if err == nil {
		fmt.Println("Terraform output conversion completed.")
	} else {
		fmt.Printf("Error saving %s due to error: %v\n", outputFilename, err)
		fmt.Println("Aborting.")
	}

}

func main() {

	fmt.Println()
	fmt.Println("Creates configuration file for Upgrading Quorum nodes")
	fmt.Println()

	var (
		err                   error
		templateFilename      string
		terraformJSONFilename string
		outputFilename        string
		terraformMode         string
		workspace             string
		auth                  string
		organization          string
	)

	flag.StringVar(&templateFilename, "template", "", "Filename of template")
	flag.StringVar(&terraformJSONFilename, "terraform-json", "", "Filename of Terraform output in JSON format (not required in TFE mode)")
	flag.StringVar(&outputFilename, "output", "", "Filename to write output to")
	flag.BoolVar(&removeQuotes, "remove-quote", false, "remove-quote=true|false")
	flag.BoolVar(&removeDelimiters, "remove-delimiter", false, "remove-delimiter=true|false")
	flag.StringVar(&terraformMode, "mode", "cli", "cli|tfe")
	flag.StringVar(&workspace, "workspace", "", "value of workspace")
	flag.StringVar(&organization, "organization", "", "name of organization")
	flag.StringVar(&auth, "auth", "", "authorization token")

	savedOsArgs := os.Args
	flag.Parse()

	if len(os.Args) < 4 && terraformMode == "cli" {
		fmt.Println("--mode=cli")
		fmt.Println("\tInput 1 - Filename of template")
		fmt.Println("\tInput 2 - Terraform output in JSON format (only in cli mode)")
		fmt.Println("\tInput 3 - Filename to write output to")
		fmt.Println("--mode=tfe")
		flag.Usage()
		fmt.Println()
		return
	}

	// legacy parsing
	if len(os.Args) >= 5 {
		// expect these flags to be in 3rd position or later...
		for i := 4; i < len(os.Args); i++ {
			removeQuotes = removeQuotes || strings.Contains(os.Args[i], "-remove-quote")             // support both -remove-quote and -remove-quotes
			removeDelimiters = removeDelimiters || strings.Contains(os.Args[i], "-remove-delimiter") // support -remove-delimiter and -remove-delimiters
		}
	}

	switch terraformMode {
	case "cli":
		{
			if len(os.Args) < len(savedOsArgs) {
				os.Args = savedOsArgs
			}
			fmt.Println("Remove Quotes: ", removeQuotes)
			fmt.Println("Remove Delimiters: ", removeDelimiters)

			var arg1, arg2, arg3 string
			if templateFilename != "" {
				arg1 = templateFilename
			} else {
				arg1 = os.Args[1]
			}
			if templateFilename, err = softwareupgrade.Expand(arg1); err != nil {
				fmt.Printf(cExpandingFilenameErr, arg1, err)
				fmt.Println()
				return
			}

			if terraformJSONFilename != "" {
				arg2 = terraformJSONFilename
			} else {
				arg2 = os.Args[2]
			}
			if terraformJSONFilename, err = softwareupgrade.Expand(arg2); err != nil {
				fmt.Printf(cExpandingFilenameErr, arg2, err)
				fmt.Println()
				return
			}

			if outputFilename != "" {
				arg3 = outputFilename
			} else {
				arg3 = os.Args[3]
			}
			if outputFilename, err = softwareupgrade.Expand(arg3); err != nil {
				fmt.Printf(cExpandingFilenameErr, arg3, err)
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

			convertTerraformJSONFile(templateFilename, terraformJSONFilename, outputFilename)
		}
	case "tfe":
		{
			if organization == "" || workspace == "" || auth == "" {
				flag.Usage()
				return
			}
			tfe := NewTFE(organization, workspace, auth)
			workspaceID, err := tfe.WorkspaceID()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Workspace: %s, retrieved ID: %s\n", workspace, workspaceID)
			fmt.Print("Calling Terraform Enterprise API: ")
			workspaceOutput, err := tfe.GetWorkspaceOutput()
			if err != nil {
				fmt.Printf("error retrieving workspace information due to %v", err)
				fmt.Println("Aborting.")
				return
			}
			fmt.Println("data retrieved.")
			nd := NewNodeData(workspaceOutput)
			data := nd.ReadNode()
			JSONBytes, err := json.Marshal(&data)
			if err == nil {
				ConvertTerraformJSONContent(templateFilename, JSONBytes, outputFilename)
			} else {
				fmt.Printf("Failure encountered during conversion: %v\n", err)
			}

		}
	}

}
