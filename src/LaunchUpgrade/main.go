package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"path"
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

	var (
		debugLogFilename, JSONFilename, FailedNodesFilename      string
		disableNodeVerification, disableFileVerification, dryRun bool
		disableTargetDirVerification                             bool
	)
	flag.BoolVar(&d.PrintDebug, "debug", false, "Specifies debug mode")
	flag.StringVar(&debugLogFilename, "debug-log", `~/Upgrade-debug.log`, "Specifies the debug log filename where logs are written to")
	flag.StringVar(&JSONFilename, "json", "", "Specifies the JSON configuration file")
	flag.StringVar(&FailedNodesFilename, "failed-nodes", "~/Upgrade-Failed.session", "Specifes the file to load/save nodes that failed to upgrade")
	flag.BoolVar(&disableNodeVerification, "disable-node-verification", false, "Disables node IP resolution verification")
	flag.BoolVar(&disableFileVerification, "disable-file-verification", false, "Disables source file existence verification")
	flag.BoolVar(&disableTargetDirVerification, "disable-target-dir-verification", false, "Disables target directory existence verification")
	flag.BoolVar(&dryRun, "dry-run", true, "Enables testing mode, doesn't upgrade")
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

	// Read JSON configuration file
	if expandedJSONFilename, err := softwareupgrade.Expand(JSONFilename); err == nil {
		JSONFilename = expandedJSONFilename
	} else {
		d.Println("Unable to interpret/parse %s due to %v", JSONFilename, err)
		return
	}
	jsonContents, err := softwareupgrade.ReadDataFromFile(JSONFilename)

	// For processing of the JSON configuration
	if err == nil {
		func() {
			var upgradeconfig softwareupgrade.UpgradeConfig
			// Parse the JSON
			json.Unmarshal(jsonContents, &upgradeconfig)

			if !disableFileVerification {
				if err := upgradeconfig.VerifyFilesExist(); err != nil {
					d.Printf("%v", err)
					return
				}
				d.Println("All source files verified.")
			}

			// GroupNames is the name given to each combination of software
			SoftwareGroupNames := upgradeconfig.GetGroupNames()
			d.Println("%d groups defined: %v", len(SoftwareGroupNames), SoftwareGroupNames)

			// Nodes contains the list of the nodes to upgrade.
			nodes := upgradeconfig.GetNodes()
			d.Println("%d nodes found: %v", len(nodes), nodes)

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
				d.Println("All nodes verified to be resolvable to IP addresses.")
			}

			if !disableTargetDirVerification {
				// Verify all target directories exist. This is also an opportunity
				// to ensure all nodes can be connected to.
				type (
					DirExistStruct struct {
						dir   string
						exist bool
					}
				)
				var (
					msg string
				)
				d.Println("Verifying target directories, please wait.")
				hostDirsCache := make(map[string]DirExistStruct)
				dupErr := make(map[string]bool)
				for _, softwareGroup := range SoftwareGroupNames {
					// Look up the software for each softwareGroup
					groupSoftware := upgradeconfig.GetGroupSoftware(softwareGroup)

					// Get the nodes for this group
					groupNodes := upgradeconfig.GetGroupNodes(softwareGroup)
					for _, node := range groupNodes {
						if len(groupSoftware) == 0 {
							continue
						}
						for _, software := range groupSoftware {
							nodeInfo := upgradeconfig.GetNodeUpgradeInfo(node, software)
							nodeDNS := node
							sshConfig := softwareupgrade.NewSSHConfig(nodeInfo.SSHUserName, nodeInfo.SSHCert, nodeDNS)
							for _, dirInfo := range nodeInfo.Copy {
								remoteDir := path.Dir(dirInfo.DestFilePath)
								hostDir := fmt.Sprintf("%s-%s", nodeDNS, remoteDir)
								hostDirStruct := hostDirsCache[hostDir]
								if hostDirStruct == (DirExistStruct{}) {
									var err error
									hostDirStruct.dir = remoteDir
									hostDirStruct.exist, err = sshConfig.DirectoryExists(remoteDir)
									if err == nil {
										hostDirsCache[hostDir] = hostDirStruct
										if !hostDirStruct.exist {
											msg = fmt.Sprintf("%sRemote directory: %s doesn't exist on node: %s\n",
												msg, remoteDir, nodeDNS)
										}
									} else {
										errmsg := fmt.Sprintf("Node: %s error: %v", nodeDNS, err)
										errExist := dupErr[errmsg]
										if !errExist {
											msg = fmt.Sprintf("%s%s\n", msg, errmsg)
											dupErr[errmsg] = true
										}
									}
								}
							}
						}
					}
				}
				if msg != "" {
					d.Println("Error(s) encountered in target directory verification.")
					d.Printf("%v", msg)
					return
				}
				d.Println("All remote directories verified.")
			}

			failedUpgradeInfo := softwareupgrade.NewFailedUpgradeInfo()
			var resumeUpgrade bool
			defer func() {
				if failedUpgradeInfo.FailedNodeSoftware != nil {
					data, err := json.Marshal(failedUpgradeInfo)
					if err == nil {
						softwareupgrade.SaveDataToFile(FailedNodesFilename, data)
					}
				}
			}()

			// if a previous session exists...
			if softwareupgrade.FileExists(FailedNodesFilename) {
				data, err := softwareupgrade.ReadDataFromFile(FailedNodesFilename)
				if err == nil {
					err = json.Unmarshal(data, &failedUpgradeInfo.FailedNodeSoftware)
					resumeUpgrade = err == nil && len(failedUpgradeInfo.FailedNodeSoftware) > 0
				}
			}

			for _, softwareGroup := range SoftwareGroupNames {
				var doPause bool
				// Look up the software for each softwareGroup
				d.Printf("Performing upgrade for software group: %s\n", softwareGroup)
				groupSoftware := upgradeconfig.GetGroupSoftware(softwareGroup)

				// Get the nodes for this group
				groupNodes := upgradeconfig.GetGroupNodes(softwareGroup)
				for _, node := range groupNodes {
					if len(groupSoftware) == 0 {
						continue
					}
					doPause = true
					for _, software := range groupSoftware {
						nodeInfo := upgradeconfig.GetNodeUpgradeInfo(node, software)
						nodeDNS := node

						// If this is a resume operation, and the node and software doesn't
						// exist in the failedUpgradeInfo then skip the current node and software.
						if resumeUpgrade {
							if !failedUpgradeInfo.ExistsNodeSoftware(node, software) {
								d.Println("Skipping software %s for node %s", software, node)
								continue
							}
						}

						d.Print("Upgrading node: %s with software: %s\n", nodeDNS, software)
						sshConfig := softwareupgrade.NewSSHConfig(nodeInfo.SSHUserName, nodeInfo.SSHCert, nodeDNS)

						// Stop the running software, upgrade it, then start the software
						StopCmd := nodeInfo.StopCmd
						StopResult, err := sshConfig.Run(StopCmd)
						if err != nil { // If stop failed, skip the upgrade!
							d.Printf(softwareupgrade.CNodeMsgSSS, nodeDNS, softwareupgrade.CStop, err)
							failedUpgradeInfo.AddNodeSoftware(nodeDNS, software)
							continue
						}
						d.Printf(softwareupgrade.CNodeMsgSSS, nodeDNS, softwareupgrade.CStop, StopResult)

						if !dryRun {
							err := nodeInfo.RunUpgrade(sshConfig) // the upgrade needs to either move or overwrite the older version
							if err != nil {
								d.Println("Error during RunUpgrade: %v", err)
								// keep track of nodes and the software that failed to upgrade on these nodes
								failedUpgradeInfo.AddNodeSoftware(nodeDNS, software)
							} else {
								d.Println("Upgraded node: %s with software %s successfully!", nodeDNS, software)
							}
						}

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
					d.Println(" completed!")
				}
				d.Println("") // leave one line between one group and next group
			}

			appStatus = "completed"
			softwareupgrade.ClearSSHConfigCache()

		}()
	} else {
		d.Println(`Error reading from JSON configuration file: "%s", error: %v`, JSONFilename, err)
	}

}
