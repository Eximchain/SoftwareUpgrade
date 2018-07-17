package softwareupgrade

type (
	// SSHConfigLocalHost ...
	LocalHostHasher struct {
	}

	Hasher interface {
		Md5sum(path string) (result string, err error)
		Sha256sum(path string) (result string, err error)
	}
)

// NewSSHConfigLocalHost returns an empty structure
func NewLocalHostHasher() (result *LocalHostHasher) {
	return &LocalHostHasher{}
}

// Hash implicitly implements HashInterface
func (config *LocalHostHasher) Sha256sum(path string) (result string, err error) {
	return localSHA256(path)
}

func (hasher *LocalHostHasher) Md5sum(path string) (result string, err error) {
	return
}
