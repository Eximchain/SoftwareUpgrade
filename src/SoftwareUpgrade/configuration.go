package softwareupgrade

import (
	"encoding/json"
	"errors"
	"log"
	"time"
)

type (
	CopyInfo struct {
		LocalFilename  string `json:"LocalFilename"`
		RemoteFilename string `json:"RemoteFilename"`
	}

	SSHInfo struct {
		SSHCert     string `json:"ssh_cert"`
		SSHUserName string `json:"ssh_username"`
	}

	UpgradeStruct struct {
		SourceFilePath string `json:"Local_Filename"`  // local file path
		DestFilePath   string `json:"Remote_Filename"` // remote file path
		Permissions    string `json:"Permissions"`     // permissions of the newly copied file
		VerifyCopy     string `json:"VerifyCopy"`      // command to run to verify copy is successful
	}

	UpgradeInfo struct {
		PostUpgrade []string `json:"postupgrade"`
		PreUpgrade  []string `json:"preupgrade"`
		StartCmd    string   `json:"start"`
		StopCmd     string   `json:"stop"`

		// The key string is actually integer, and the order of the
		// copy will be numeric order.
		Copy map[string]UpgradeStruct `json:"Copy"`
	}

	NodeInfoContainer struct {
		UpgradeInfo
		SSHInfo
	}

	// This specifies the upgrade configuration for each node,
	// if the node is not specified
	NodeUpgradeConfig struct {
		NodeUpgradeInfo map[string]UpgradeInfo
	}

	Duration struct {
		time.Duration
	}

	// UpgradeConfig contains the configuration for upgrading nodes
	UpgradeConfig struct {
		Common struct {
			SSHInfo                           // This specifies the general and common SSL configuration for common nodes
			SoftwareGroup map[string][]string `json:"software_group"` // This specifies the software type that's possible to run on a node, the start and stop command, the command used to upgrade the software
			GroupPause    Duration            `json:"group_pause_after_upgrade"`
		} `json:"common"`
		Nodes    map[string]NodeInfoContainer `json:"nodes"`    // This is a map with the key as the DNS hostnames of each node that participates in the network
		Software map[string]UpgradeInfo       `json:"software"` // this defines each individual piece of software
		// This specifies the combination of software on each node.
		// So, say, vault and quorum is one group 1.
		// blockmetrics, cloudwatchmetrics, constellation is group 2.
		// So node1 and node3 runs group 1.
		// node2 and node4 runs group 2.
		// node5, node6, node7 runs group 3.
		SoftwareGroupNodes map[string][]string `json:"groupnodes"`
	}
)

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func (nodeInfo *NodeInfoContainer) RunUpgrade(sshConfig *SSHConfig) {
	// Support i := 0 or i := 1 by checking for empty struct
	if len(nodeInfo.Copy) > 0 {
		for i := 0; i < len(nodeInfo.Copy)+1; i++ {
			index := IntToStr(i)
			upgradeStruct := nodeInfo.Copy[index]
			if (UpgradeStruct{}) == upgradeStruct || upgradeStruct.SourceFilePath == "" { // skip empty struct, or empty source
				continue
			}
			err := sshConfig.CopyLocalFileToRemoteFile(
				upgradeStruct.SourceFilePath,
				upgradeStruct.DestFilePath, upgradeStruct.Permissions)
			if err != nil {
				log.Printf("Error during RunUpgrade: %s", err)
			}
		}
	}
}

func (config *UpgradeConfig) GetGroupNames() (result []string) {
	for groupKey, _ := range config.SoftwareGroupNodes {
		result = append(result, groupKey)
	}
	return
}

func (config *UpgradeConfig) GetGroupNodes(groupName string) (result []string) {
	result = config.SoftwareGroupNodes[groupName]
	return
}

func (config *UpgradeConfig) GetGroupSoftware(groupName string) (result []string) {
	result = config.Common.SoftwareGroup[groupName]
	return
}

func (config *UpgradeConfig) GetNodes() (result []string) {
	for _, groupNodesList := range config.SoftwareGroupNodes {
		for _, nodeDNS := range groupNodesList {
			result = append(result, nodeDNS)
		}
	}
	return
}

// getUpgradeInfo gets the specific upgrade information for a particular node's software.
func (config *UpgradeConfig) GetNodeUpgradeInfo(node, software string) (result *NodeInfoContainer) {
	result = &NodeInfoContainer{}
	nodeInfo := config.Nodes[node]
	if len(nodeInfo.PostUpgrade) > 0 {
		result.PostUpgrade = nodeInfo.PostUpgrade
	} else {
		result.PostUpgrade = config.Software[software].PostUpgrade
	}
	if len(nodeInfo.PreUpgrade) > 0 {
		result.PreUpgrade = nodeInfo.PreUpgrade
	} else {
		result.PreUpgrade = config.Software[software].PreUpgrade
	}
	if nodeInfo.StartCmd != "" {
		result.StartCmd = nodeInfo.StartCmd
	} else {
		result.StartCmd = config.Software[software].StartCmd
	}
	if nodeInfo.StopCmd != "" {
		result.StopCmd = nodeInfo.StopCmd
	} else {
		result.StopCmd = config.Software[software].StopCmd
	}
	if nodeInfo.SSHUserName != "" {
		result.SSHUserName = nodeInfo.SSHUserName
	} else {
		result.SSHUserName = config.Common.SSHUserName
	}
	if nodeInfo.SSHCert != "" {
		result.SSHCert = nodeInfo.SSHCert
	} else {
		result.SSHCert = config.Common.SSHCert
	}
	if len(nodeInfo.Copy) > 0 {
		result.Copy = nodeInfo.Copy
	} else {
		result.Copy = config.Software[software].Copy
	}
	if (len(result.Copy) == 0) || (len(result.PreUpgrade) == 0) || (len(result.PostUpgrade) == 0) ||
		(result.SSHCert == "") || (result.SSHUserName == "") || (result.StartCmd == "") || (result.StopCmd == "") {
		// panic(fmt.Sprintf("One of the fields is empty! %+v", result))
	}
	return
}
