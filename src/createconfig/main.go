package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"softwareupgrade"
	"strings"
)

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
		log.Printf("Error parsing Terraform JSON: %v\n", err)
		log.Fatalln("Most probable cause of error is forgetting to add -json to terraform output")
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
		log.Printf("Error saving %s due to error: %v\n", outputFilename, err)
		log.Fatalln("Aborting.")
	}

}

func convertTerraformJSONFile(inputTemplate, jsonPath, outputFilename string) {
	data, err := softwareupgrade.ReadDataFromFile(jsonPath)
	if err == nil {
		ConvertTerraformJSONContent(inputTemplate, data, outputFilename)
	} else {
		log.Printf("Unable to read from %s due to error: %v\n", jsonPath, err)
		log.Fatalln("Aborting.")
	}
}

func convertTerraformJSON(inputTemplate, jsonPath, outputFilename string) {
	var result string
	data, err := softwareupgrade.ReadDataFromFile(inputTemplate)
	if err == nil {
		result = string(data)
	} else {
		log.Printf("Unable to read from %s due to error: %v\n", inputTemplate, err)
		log.Fatalln("Aborting.")
	}
	data, err = softwareupgrade.ReadDataFromFile(jsonPath)
	if err != nil {
		log.Printf("Unable to read from %s due to error: %v\n", jsonPath, err)
		log.Fatalln("Aborting.")
	}
	var TerraformOutput map[string]TerraformNode
	err = json.Unmarshal(data, &TerraformOutput)
	if err != nil {
		log.Printf("Error parsing Terraform JSON: %v\n", err)
		log.Fatalln("Most probable cause of error is forgetting to add -json to terraform output")
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
		log.Println("Terraform output conversion completed.")
	} else {
		log.Printf("Error saving %s due to error: %v\n", outputFilename, err)
		log.Fatalln("Aborting.")
	}

}

func main() {

	log.SetOutput(os.Stdout)                   // redirect logger output so that the log.Fatal* are redirected to stdout
	log.SetFlags(log.Flags() &^ log.LstdFlags) // remove date/time from any log messages
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

	flag.Parse()

	if templateFilename == "" {
		log.Fatalln("No template filename specified.")
	}
	if outputFilename == "" {
		log.Fatalln("No output filename specified.")
	}

	fmt.Println("Remove Quotes: ", removeQuotes)
	fmt.Println("Remove Delimiters: ", removeDelimiters)

	switch terraformMode {
	case "cli":
		{

			if templateFilename, err = softwareupgrade.Expand(templateFilename); err != nil {
				log.Fatalf(cExpandingFilenameErr, templateFilename, err)
			}

			if terraformJSONFilename == "" {
				log.Fatalln("No Terraform JSON filename specified.")
			}
			if terraformJSONFilename, err = softwareupgrade.Expand(terraformJSONFilename); err != nil {
				log.Fatalf(cExpandingFilenameErr, terraformJSONFilename, err)
			}

			if outputFilename, err = softwareupgrade.Expand(outputFilename); err != nil {
				log.Fatalf(cExpandingFilenameErr, outputFilename, err)
			}

			if !softwareupgrade.FileExists(templateFilename) {
				log.Fatalf(cFilenameDoesntExist, templateFilename)
			}

			if !softwareupgrade.FileExists(terraformJSONFilename) {
				log.Fatalf(cFilenameDoesntExist, terraformJSONFilename)
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
			defer tfe.Destroy()
			workspaceID, err := tfe.WorkspaceID()
			if err != nil {
				log.Fatalf("Error: %v\n", err)
			}
			fmt.Printf("Workspace: %s, retrieved ID: %s\n", workspace, workspaceID)
			fmt.Print("Calling Terraform Enterprise API: ")
			workspaceOutput, err := tfe.GetWorkspaceOutput()
			if err != nil {
				fmt.Printf("error retrieving workspace information due to %v\n", err)
				log.Fatalln("Aborting.")
			}
			fmt.Println("data retrieved.")
			nd := NewNodeData(workspaceOutput)
			data := nd.ReadNode()
			JSONBytes, err := json.Marshal(&data)
			if err != nil {
				log.Printf("Failure encountered during conversion: %v\n", err)
				log.Fatalln("Aborting.")
			}
			ConvertTerraformJSONContent(templateFilename, JSONBytes, outputFilename)
		}
	}
	fmt.Printf("Configuration file: %s created successfully.\n", outputFilename)
}
