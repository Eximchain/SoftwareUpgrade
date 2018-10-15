package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"softwareupgrade"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	// TerraformNode represents the output for a terraform output -json
	TerraformNode struct {
		Sensitive bool        `json:"sensitive"`
		Type      string      `json:"type"`  // could be map, list
		Value     interface{} `json:"value"` // could be []string, or map[string][]string
	}
)

// ConvertTerraformJSONContent reads the template from the given inputTemplate filename, combines
// it with the data from the JSONContent, and writes to the file specified by the outputFilename
func ConvertTerraformJSONContent(removeQuotes, removeDelimiters bool, inputTemplate string, JSONContent []byte, outputFilename string) {
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

func convertTerraformJSONFile(removeQuotes, removeDelimiters bool, inputTemplate, jsonPath, outputFilename string) {
	data, err := softwareupgrade.ReadDataFromFile(jsonPath)
	if err == nil {
		ConvertTerraformJSONContent(removeQuotes, removeDelimiters, inputTemplate, data, outputFilename)
	} else {
		log.Printf("Unable to read from %s due to error: %v\n", jsonPath, err)
		log.Fatalln("Aborting.")
	}
}

func convertTerraformJSON(removeQuotes, removeDelimiters bool, inputTemplate, jsonPath, outputFilename string) {
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

type (
	// AppInfo is information required by the app
	AppInfo struct {
		templateFilename      string
		terraformJSONFilename string
		outputFilename        string
		terraformMode         string
		workspace             string
		auth                  string
		organization          string
		removeQuotes          bool
		removeDelimiters      bool
		// for AWS
		// networkId int32
		// role      string
		akid          string
		sak           string
		region        string
		separator     string
		filteroutputs FilterOutputs
		debug         bool
	}
)

func mainApp(appInfo *AppInfo) {
	var (
		err error
	)
	switch appInfo.terraformMode {
	default:
		{
			fmt.Println("Wrong mode")
			fmt.Println()

			flag.Usage()

		}
	case "aws":
		{

			if appInfo.debug {
				DumpFilterOutputs(appInfo.filteroutputs)
				debug.EnableDebug()
			}

			a := &AWS{}

			if appInfo.sak != "" && appInfo.akid != "" {
				a.NewAWSSessionCredentials(appInfo.akid, appInfo.sak, "us-east-1")
			}

			if appInfo.region != "" {
				a.SetRegion(appInfo.region)
			}

			if regions, err := a.GetRegions(); err == nil {
				i := -1

				if len(regions) == 0 {
					fmt.Println("Unable to retrieve AWS regions!")
					return
				}

				WithinRegion := func() (result bool) {
					i = i + 1
					result = i < len(regions)
					return
				}

				cachedInstances := make(map[string][]*ec2.Instance)
				aTerraformType := makeTerraformType()

				LoopRegions := func() {
					region := regions[i]
					if region != "" {

						a.SetRegion(region) // change the AWS region so that data is returned correctly
						fmt.Println("Querying region: ", region)

						var (
							instances []*ec2.Instance
							found     bool
						)

						if instances, found = cachedInstances[region]; !found {
							instances = a.GetInstances()
							cachedInstances[region] = instances
						}

						for _, filteroutput := range appInfo.filteroutputs {
							terraformItem := aTerraformType.GetTerraformItem(filteroutput.TerraformName, filteroutput.format)
							terraformItem.SetRegion(region)

							for i, instance := range instances {
								debug.Printf("Instance %d\n", i)
								//fmt.Printf("%+v\n", instance)

								if matchedFilters(instance, filteroutput.Filters) {
									line := getOutputs(instance, &filteroutput) // this is an IP or DNS
									terraformItem.AddIP(line)
								}

							}
						}
						debug.Println("Instance done--------------------")
					}
					debug.Println("=================================")
					debug.Println()
				}

				ForEach(WithinRegion, LoopRegions)
				aTerraformType.SaveToFile(appInfo.outputFilename)
			} else {
				log.Fatalf("Error: %v\n", err)
			}
		}
	case "cli":
		{

			if appInfo.templateFilename, err = softwareupgrade.Expand(appInfo.templateFilename); err != nil {
				log.Fatalf(cExpandingFilenameErr, appInfo.templateFilename, err)
			}

			if appInfo.terraformJSONFilename == "" {
				log.Fatalln("No Terraform JSON filename specified.")
			}
			if appInfo.terraformJSONFilename, err = softwareupgrade.Expand(appInfo.terraformJSONFilename); err != nil {
				log.Fatalf(cExpandingFilenameErr, appInfo.terraformJSONFilename, err)
			}

			if appInfo.outputFilename, err = softwareupgrade.Expand(appInfo.outputFilename); err != nil {
				log.Fatalf(cExpandingFilenameErr, appInfo.outputFilename, err)
			}

			if !softwareupgrade.FileExists(appInfo.templateFilename) {
				log.Fatalf(cFilenameDoesntExist, appInfo.templateFilename)
			}

			if !softwareupgrade.FileExists(appInfo.terraformJSONFilename) {
				log.Fatalf(cFilenameDoesntExist, appInfo.terraformJSONFilename)
			}

			convertTerraformJSONFile(appInfo.removeQuotes, appInfo.removeDelimiters, appInfo.templateFilename, appInfo.terraformJSONFilename, appInfo.outputFilename)
		}
	case "tfe":
		{
			if appInfo.organization == "" || appInfo.workspace == "" || appInfo.auth == "" {
				flag.Usage()
				return
			}
			tfe := NewTFE(appInfo.organization, appInfo.workspace, appInfo.auth)
			defer tfe.Destroy()
			workspaceID, err := tfe.WorkspaceID()
			if err != nil {
				log.Fatalf("Error: %v\n", err)
			}
			fmt.Printf("Workspace: %s, retrieved ID: %s\n", appInfo.workspace, workspaceID)
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
			ConvertTerraformJSONContent(appInfo.removeQuotes, appInfo.removeDelimiters, appInfo.templateFilename, JSONBytes, appInfo.outputFilename)
		}
	}
	if softwareupgrade.FileExists(appInfo.outputFilename) {
		fmt.Printf("Output file: %s created successfully.\n", appInfo.outputFilename)
	}
}

func main() {

	log.SetOutput(os.Stdout)                   // redirect logger output so that the log.Fatal* are redirected to stdout
	log.SetFlags(log.Flags() &^ log.LstdFlags) // remove date/time from any log messages
	fmt.Println()
	fmt.Println("Creates configuration file for Upgrading Quorum nodes")
	fmt.Println()

	var appInfo AppInfo

	flag.StringVar(&appInfo.templateFilename, "template", "", "Filename of template")
	flag.StringVar(&appInfo.terraformJSONFilename, "terraform-json", "", "Filename of Terraform output in JSON format (not required in TFE mode)")
	flag.StringVar(&appInfo.outputFilename, "output", "", "Filename to write output to")
	flag.BoolVar(&appInfo.removeQuotes, "remove-quote", false, "remove-quote=true|false")
	flag.BoolVar(&appInfo.removeDelimiters, "remove-delimiter", false, "remove-delimiter=true|false")
	flag.StringVar(&appInfo.terraformMode, "mode", "cli", "aws|cli|tfe")
	flag.StringVar(&appInfo.workspace, "workspace", "", "value of workspace")
	flag.StringVar(&appInfo.organization, "organization", "", "name of organization")
	flag.StringVar(&appInfo.auth, "auth", "", "authorization token")
	flag.BoolVar(&appInfo.debug, "debug", false, "Enable debug logging during AWS query")

	flag.StringVar(&appInfo.akid, "akid", "", "AWS Access Key ID")
	flag.StringVar(&appInfo.sak, "sak", "", "AWS Secret Access Key")
	flag.StringVar(&appInfo.region, "region", "", "AWS Region")

	// used as below, this creates an array of 10 output conditional flags
	// -output1="Key=NetworkID,Value=64821,Role=Vault;PublicIpAddress" // when Key=NetworkID and its Value=64821, output the PublicIpAddress of the instance
	// -output2="Key=NetworkID,Value="
	outputConditions := make([]string, 10)
	for i := 0; i < len(outputConditions); i++ {
		flagName := fmt.Sprintf("output%d", i+1)
		usage := fmt.Sprintf(`Defines %s=filterFieldName1=value1[,filterFieldName2=value2]+;outputFieldName;(list|map);itemname`, flagName)
		flag.StringVar(&outputConditions[i], flagName, "", usage)
	}

	flag.Parse()

	removeEmptyConditions(&outputConditions)
	appInfo.filteroutputs = makeFilterOutputs(&outputConditions)

	if appInfo.terraformMode != "aws" {
		if appInfo.templateFilename == "" {
			log.Fatalln("No template filename specified.")
		}
		if appInfo.outputFilename == "" {
			log.Fatalln("No output filename specified.")
		}
	}

	fmt.Println("Remove Quotes: ", appInfo.removeQuotes)
	fmt.Println("Remove Delimiters: ", appInfo.removeDelimiters)

	mainApp(&appInfo)

}
