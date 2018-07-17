package softwareupgrade

import (
	"log"
	"testing"
)

var (
	failedUpgradeInfo *FailedUpgradeInfo
)

func init() {
	failedUpgradeInfo = NewFailedUpgradeInfo()
	failedUpgradeInfo.AddNodeSoftware("node1", "soft1")
	failedUpgradeInfo.AddNodeSoftware("node1", "soft2")
	failedUpgradeInfo.AddNodeSoftware("node1", "soft3")
	failedUpgradeInfo.AddNodeSoftware("node1", "soft4")
	failedUpgradeInfo.RemoveNodeSoftware("node1", "soft3")
	if count := failedUpgradeInfo.GetNodeSoftwareCount("node1"); count != 3 {
		log.Printf("Node count should be 3, but is: %d\n", count)
	}
}

func TestFailedUpgradeInfo_RemoveNodeSoftware(t *testing.T) {
	failedUpgradeInfo = NewFailedUpgradeInfo()
	if count := failedUpgradeInfo.GetNodeSoftwareCount("node1"); count != 0 {
		t.Fatalf("Node count should be 0, but is: %d", count)
	}
}

func TestFailedUpgradeInfo_AddNodeSoftware(t *testing.T) {
	failedUpgradeInfo = NewFailedUpgradeInfo()
	failedUpgradeInfo.AddNodeSoftware("node1", "software")
	if count := failedUpgradeInfo.GetNodeSoftwareCount("node1"); count != 1 {
		t.Fatalf("Node count shoud be 1, but is: %d", count)
	}
}
