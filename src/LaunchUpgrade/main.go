package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"softwareupgrade"
	"time"
)

var (
	d                    softwareupgrade.DebugLog
	commonSSLcertContent []byte
	userSSLcertContent   []byte
)

func main() {
	debugFlagPtr := flag.Bool("debug", false, "Specifies debug mode")
	debugLogFilePtr := flag.String("debug-log", "", "Specifies the debug log filename where logs are written to")
	// configFilePtr := flag.String("conf", "/Users/chuacw/Documents/GitHub/SoftwareUpgrade/config.ini", "Specifies the configuration file")
	jsonFilePtr := flag.String("json", "LaunchUpgrade.json", "Specifies the JSON configuration file")
	flag.Parse()
	d.PrintDebug = *debugFlagPtr
	err := d.EnableDebugLog(debugLogFilePtr)
	if err == nil {
		defer d.CloseDebugLog()
	} else {
		fmt.Printf("Error: %v", err)
	}
	d.Println(softwareupgrade.CEximchainUpgradeTitle)
	defer d.Println("Upgrade completed")

	JSONFilename := *jsonFilePtr
	jsonContents, err := softwareupgrade.ReadDataFromFile(JSONFilename)

	// For processing of the JSON configuration
	if err == nil {
		func() {
			var upgradeconfig softwareupgrade.UpgradeConfig
			json.Unmarshal(jsonContents, &upgradeconfig)
			// GroupNames is the name given to each combination of software
			SoftwareGroupNames := upgradeconfig.GetGroupNames()
			d.Print("Groups defined: %+v", SoftwareGroupNames)

			// Nodes contains the list of the nodes to upgrade.
			nodes := upgradeconfig.GetNodes()
			d.Print("Nodes found: %+v", nodes)

			for _, softwareGroup := range SoftwareGroupNames {
				var doPause bool
				// Look up the software for each softwareGroup
				d.Printf("Performing upgrade for software group: %s\n", softwareGroup)
				groupSoftware := upgradeconfig.GetGroupSoftware(softwareGroup)
				// fmt.Printf("Group Software: %+v\n", groupSoftware)
				groupNodes := upgradeconfig.GetGroupNodes(softwareGroup)
				for _, node := range groupNodes {
					if len(groupSoftware) == 0 {
						continue
					}
					doPause = true
					for _, software := range groupSoftware {
						nodeInfo := upgradeconfig.GetNodeUpgradeInfo(node, software)
						nodeDNS := node
						d.Print("Upgrading node: %s with software: %s\n", nodeDNS, software)
						sshConfig := softwareupgrade.NewSSHConfig(nodeInfo.SSHUserName, nodeInfo.SSHCert, nodeDNS)

						// Stop the running software, upgrade it, then start the software
						StopCmd := nodeInfo.StopCmd
						StopResult, err := sshConfig.Run(StopCmd)
						if err != nil { // If stop failed, skip the upgrade!
							d.Printf(softwareupgrade.CNodeMsgSSS, nodeDNS, softwareupgrade.CStop, err)
							continue
						}
						d.Printf(softwareupgrade.CNodeMsgSSS, nodeDNS, softwareupgrade.CStop, StopResult)

						nodeInfo.RunUpgrade(sshConfig) // the upgrade needs to either move or delete the older version

						StartCmd := nodeInfo.StartCmd
						StartResult, err := sshConfig.Run(StartCmd)
						if err != nil {
							d.Printf(softwareupgrade.CNodeMsgSSS, nodeDNS, softwareupgrade.CStart, err)
							continue
						}
						d.Printf(softwareupgrade.CNodeMsgSSS, nodeDNS, softwareupgrade.CStart, StartResult)

					}
				}
				if doPause { // pause only if upgrade has been run
					d.Printf("Pausing for %s...", upgradeconfig.Common.GroupPause)
					time.Sleep(upgradeconfig.Common.GroupPause.Duration)
					d.Println("Pause completed!")
				}
			}

		}()
	} else {
		log := fmt.Sprintf("Error reading from JSON configuration file: %s, error: %v", JSONFilename, err)
		d.Println(log)
	}

}
