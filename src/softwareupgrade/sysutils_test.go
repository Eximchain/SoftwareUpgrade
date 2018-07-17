package softwareupgrade

import (
	"testing"
)

func TestSHA256(t *testing.T) {
	localSHA256("/bin/mv")
}
