package softwareupgrade

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// IntToString converts an integer to a string
func IntToStr(value int) string {
	return strconv.Itoa(value)
}

func expand(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
