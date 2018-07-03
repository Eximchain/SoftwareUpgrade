package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"softwareupgrade"
	"time"
)

var (
	d                    softwareupgrade.DebugLog
	commonSSLcertContent []byte
	userSSLcertContent   []byte
)

func main() {
	fmt.Println(softwareupgrade.CEximchainUpgradeTitle)

	var debugLogFilename, JSONFilename string
	var disableNodeVerification, disableFileVerification bool
	flag.BoolVar(&d.PrintDebug, "debug", false, "Specifies debug mode")
	flag.StringVar(&debugLogFilename, "debug-log", `~/Upgrade-debug.log`, "Specifies the debug log filename where logs are written to")
	flag.StringVar(&JSONFilename, "json", "", "Specifies the JSON configuration file")
	flag.BoolVar(&disableNodeVerification, "disable-node-verification", false, "Disables node IP resolution verification")
	flag.BoolVar(&disableFileVerification, "disable-file-verification", false, "Disables source file existence verification")
	flag.Parse()

	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		return
	}

	var err error
	if d.PrintDebug && debugLogFilename != "" {
		err = d.EnableDebugLog(debugLogFilename)
	}
	if err == nil {
		defer d.CloseDebugLog()
	} else {
		d.Println("Error: %v", err)
	}
	d.Debugln(softwareupgrade.CEximchainUpgradeTitle)
	d.EnablePrintConsole()
	appStatus := "aborted"
	defer func() {
		d.Println("Upgrade %s", appStatus)
	}()

	jsonContents, err := softwareupgrade.ReadDataFromFile(JSONFilename)

	// For processing of the JSON configuration
	if err == nil {
		func() {
			var upgradeconfig softwareupgrade.UpgradeConfig
			json.Unmarshal(jsonContents, &upgradeconfig)

			if !disableFileVerification {
				if err := upgradeconfig.VerifyFilesExist(); err != nil {
					d.Printf("%v", err)
					return
				}
			}

			// GroupNames is the name given to each combination of software
			SoftwareGroupNames := upgradeconfig.GetGroupNames()
			d.Print("Groups defined: %+v", SoftwareGroupNames)

			// Nodes contains the list of the nodes to upgrade.
			nodes := upgradeconfig.GetNodes()
			d.Print("Nodes found: %+v", nodes)

			if !disableNodeVerification {
				// Verify all nodes can be looked up using IP address.
				var msg string
				for _, node := range nodes {
					_, err := net.LookupIP(node)
					if err != nil {
						msg = fmt.Sprintf("%sCan't resolve %s\n", msg, node)
					}
				}
				if msg != "" {
					d.Print(msg)
					return
				}
			}

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

						nodeInfo.RunUpgrade(sshConfig) // the upgrade needs to either move or overwrite the older version

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

			appStatus = "completed"

		}()
	} else {
		d.Println(`Error reading from JSON configuration file: "%s", error: %v`, JSONFilename, err)
	}

}
