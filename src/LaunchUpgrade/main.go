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

type (
	tAction int
)

const (
	appActionUnknown tAction = iota
	appActionUpgrade
	appActionRollback
)

func isValidAction(action tAction) bool {
	return action >= appActionUpgrade && action <= appActionRollback
}

var (
	d                                                        softwareupgrade.DebugLog
	commonSSLcertContent                                     []byte
	userSSLcertContent                                       []byte
	appStatus                                                string
	debugLogFilename, failedNodesFilename                    string
	rollbackInfoFilename                                     string
	jsonFilename                                             string
	disableNodeVerification, disableFileVerification, dryRun bool
	disableTargetDirVerification                             bool
	jsonContents                                             []byte
	mode, rollbackSuffix                                     string
	action                                                   tAction
)

func upgradeOrRollback() {
	var upgradeconfig softwareupgrade.UpgradeConfig
	// Parse the JSON
	json.Unmarshal(jsonContents, &upgradeconfig)

	if upgradeconfig.Common.SSHTimeout != "" {
		parsedTimeout, err := time.ParseDuration(upgradeconfig.Common.SSHTimeout)
		if err == nil {
			softwareupgrade.SetSSHTimeout(parsedTimeout)
		} else {
			softwareupgrade.SetSSHTimeout(5 * time.Second)
		}
	} else {
		softwareupgrade.SetSSHTimeout(5 * time.Second)
	}

	d.Println("This session PID: %d rollback file: %s", os.Getpid(), rollbackInfoFilename)

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
					sshConfig := softwareupgrade.NewSSHConfig(nodeInfo.SSHUserName, nodeInfo.SSHCert, node)
					for _, dirInfo := range nodeInfo.Copy {
						remoteDir := path.Dir(dirInfo.DestFilePath)
						hostDir := fmt.Sprintf("%s-%s", node, remoteDir)
						hostDirStruct := hostDirsCache[hostDir]
						if hostDirStruct == (DirExistStruct{}) {
							var err error
							hostDirStruct.dir = remoteDir
							if hostDirStruct.exist, err = sshConfig.DirectoryExists(remoteDir); err == nil {
								hostDirsCache[hostDir] = hostDirStruct
								if !hostDirStruct.exist {
									msg = fmt.Sprintf("%sRemote directory: %s doesn't exist on node: %s\n",
										msg, remoteDir, node)
								}
							} else {
								errmsg := fmt.Sprintf("Node: %s error: %v", node, err)
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
	rollbackInfo := softwareupgrade.NewRollbackSession(rollbackSuffix)

	var resumeUpgrade bool
	defer func() {

		// If the failedUpgradeInfo structure isn't empty, there's a failure in upgrading
		// so save the information.
		if !failedUpgradeInfo.Empty() {
			data, err := json.Marshal(failedUpgradeInfo)
			if err == nil {
				softwareupgrade.SaveDataToFile(failedNodesFilename, data)
			} else {
				d.Println("Unable to marshal the failed upgrade information.")
			}
		}

		if !rollbackInfo.RollbackInfo.Empty() {
			data, err := json.Marshal(rollbackInfo)
			if err == nil {
				softwareupgrade.SaveDataToFile(rollbackInfoFilename, data)
			} else {
				d.Println("Unable to marshal the failed upgrade information.")
			}
		}

	}()

	switch action {
	case appActionRollback:
		{
			if softwareupgrade.FileExists(rollbackInfoFilename) {
				data, err := softwareupgrade.ReadDataFromFile(rollbackInfoFilename)
				if err == nil {
					err = json.Unmarshal(data, &rollbackInfo)
					if err != nil {
						// Clear the data so that it's not persisted again
						rollbackInfo.RollbackInfo.Clear()
						failedUpgradeInfo.Clear()
						return
					}
					rollbackSuffix = rollbackInfo.SessionSuffix
				}
			}
		}
	case appActionUpgrade:
		{
			// if a previous session exists...
			if softwareupgrade.FileExists(failedNodesFilename) {
				data, err := softwareupgrade.ReadDataFromFile(failedNodesFilename)
				if err == nil {
					err = json.Unmarshal(data, &failedUpgradeInfo.FailedNodeSoftware)
					resumeUpgrade = err == nil && len(failedUpgradeInfo.FailedNodeSoftware) > 0
				} else {
					d.Printf("Unable to read data from the failed nodes session due to error: %v", err)
				}
			} else {
				// Build the failedNodeSoftware list since this is a new session
				for _, softwareGroup := range SoftwareGroupNames {
					groupSoftware := upgradeconfig.GetGroupSoftware(softwareGroup)
					groupNodes := upgradeconfig.GetGroupNodes(softwareGroup)
					for _, node := range groupNodes {
						for _, software := range groupSoftware {
							failedUpgradeInfo.AddNodeSoftware(node, software)
						}
					}
				}
			}
		}
	}
	defer func() {
		// shows upgrade/rollback aborted/completed on app completion
		d.Println("%s %s", mode, appStatus)
	}()

	appStatus = "aborted"

	if Terminated() {
		return
	}

	for _, softwareGroup := range SoftwareGroupNames {
		var doPause bool
		// Look up the software for each softwareGroup
		d.Printf("Performing %s for software group: %s\n", mode, softwareGroup)
		groupSoftware := upgradeconfig.GetGroupSoftware(softwareGroup)

		// Get the nodes for this group
		groupNodes := upgradeconfig.GetGroupNodes(softwareGroup)
		for _, node := range groupNodes {
			if len(groupSoftware) == 0 {
				continue
			}
			if Terminated() {
				break
			}
			doPause = true
			for _, software := range groupSoftware {
				if Terminated() {
					break
				}

				// If this is a rollback, and the node and software doesn't existn
				// in the rollback data, then skip to the next one
				if action == appActionRollback {
					if !rollbackInfo.RollbackInfo.ExistsNodeSoftware(node, software) {
						continue
					}
				}

				nodeInfo := upgradeconfig.GetNodeUpgradeInfo(node, software)

				// If this is a resume operation, and the node and software doesn't
				// exist in the failedUpgradeInfo then skip the current node and software.
				if resumeUpgrade {
					if !failedUpgradeInfo.ExistsNodeSoftware(node, software) {
						d.Println("Skipping software %s for node %s", software, node)
						continue
					}
				}

				d.Print("Upgrading node: %s with software: %s\n", node, software)
				sshConfig := softwareupgrade.NewSSHConfig(nodeInfo.SSHUserName, nodeInfo.SSHCert, node)

				// Stop the running software, upgrade it, then start the software
				StopCmd := nodeInfo.StopCmd
				StopResult, err := sshConfig.Run(StopCmd)
				if err != nil { // If stop failed, skip the upgrade!
					d.Printf(softwareupgrade.CNodeMsgSSS, node, softwareupgrade.CStop, err)
					continue
				}
				d.Printf(softwareupgrade.CNodeMsgSSS, node, softwareupgrade.CStop, StopResult)

				if !dryRun {
					switch action {
					case appActionRollback:
						{

							err := nodeInfo.RunRollback(sshConfig, rollbackSuffix)
							if err != nil {
								d.Println("Error during rollback: %v", err)
							} else {
								d.Println("Rollback node: %s with software %s successfully", node, software)
								rollbackInfo.RollbackInfo.RemoveNodeSoftware(node, software)
							}
						}
					case appActionUpgrade:
						{
							err := nodeInfo.RunUpgrade(sshConfig) // the upgrade needs to either move or overwrite the older version
							if err != nil {
								d.Println("Error during RunUpgrade: %v", err)
							} else {
								d.Println("Upgraded node: %s with software %s successfully!", node, software)
								failedUpgradeInfo.RemoveNodeSoftware(node, software)
								rollbackInfo.RollbackInfo.AddNodeSoftware(node, software)
							}
						}
					}
				}

				StartCmd := nodeInfo.StartCmd
				StartResult, err := sshConfig.Run(StartCmd)
				if err != nil {
					d.Printf(softwareupgrade.CNodeMsgSSS, node, softwareupgrade.CStart, err)
					continue
				}
				d.Printf(softwareupgrade.CNodeMsgSSS, node, softwareupgrade.CStart, StartResult)
			}
		}
		if Terminated() {
			break
		}
		if doPause { // pause only if upgrade has been run
			d.Printf("Pausing for %s...", upgradeconfig.Common.GroupPause)
			time.Sleep(upgradeconfig.Common.GroupPause.Duration)
			d.Println(" completed!")
		}
		d.Println("") // leave one line between one group and next group
	}

	if !Terminated() {
		appStatus = "completed"
	}
	softwareupgrade.ClearSSHConfigCache()

}

func main() {
	fmt.Println(softwareupgrade.CEximchainUpgradeTitle)

	rollbackSuffix = softwareupgrade.GetBackupSuffix()
	defaultRollbackName := fmt.Sprintf("~/Upgrade-Rollback-%s.session", rollbackSuffix)

	flag.StringVar(&mode, "mode", "upgrade", "mode (upgrade|rollback)")
	flag.BoolVar(&d.PrintDebug, "debug", false, "Specifies debug mode")
	flag.StringVar(&debugLogFilename, "debug-log", `~/Upgrade-debug.log`, "Specifies the debug log filename where logs are written to")
	flag.StringVar(&jsonFilename, "json", "", "Specifies the JSON configuration file to load nodes from")
	flag.StringVar(&failedNodesFilename, "failed-nodes", "~/Upgrade-Failed.session", "Specifes the file to load/save nodes that failed to upgrade")
	flag.StringVar(&rollbackInfoFilename, "rollback-nodes", defaultRollbackName, "Specifies the rollback filename for this session")
	flag.BoolVar(&disableNodeVerification, "disable-node-verification", false, "Disables node IP resolution verification")
	flag.BoolVar(&disableFileVerification, "disable-file-verification", false, "Disables source file existence verification")
	flag.BoolVar(&disableTargetDirVerification, "disable-target-dir-verification", false, "Disables target directory existence verification")
	flag.BoolVar(&dryRun, "dry-run", true, "Enables testing mode, doesn't upgrade, but starts and stops the software running on remote nodes")
	flag.Parse()

	switch mode {
	case "rollback":
		{
			action = appActionRollback
		}
	case "upgrade":
		{
			action = appActionUpgrade
		}
	}

	// Ensures that JSONFilename is provided by user
	// and that mode must either be rollback or upgrade
	if len(os.Args) <= 1 || jsonFilename == "" || !isValidAction(action) {
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

	// Read JSON configuration file
	if expandedJSONFilename, err := softwareupgrade.Expand(jsonFilename); err == nil {
		jsonFilename = expandedJSONFilename
	} else {
		d.Println("Unable to interpret/parse %s due to %v", jsonFilename, err)
		return
	}
	jsonContents, err = softwareupgrade.ReadDataFromFile(jsonFilename)

	EnableSignalHandler()
	// For processing of the JSON configuration
	if err == nil {
		upgradeOrRollback()
	} else {
		d.Println(`Error reading from JSON configuration file: "%s", error: %v`, jsonFilename, err)
	}

}
