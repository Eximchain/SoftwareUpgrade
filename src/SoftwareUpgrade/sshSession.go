package softwareupgrade

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"golang.org/x/crypto/ssh"
)

// The SSHConfig-related functions rely on the presence of pkill and pgrep on the target system
// pgrep and pkill is assumed to be located in the environmental PATH

type (
	// SSHConfig is used to carry the username, privatekey and the host to connect to.
	SSHConfig struct {
		user            string
		privateKey      string
		HostIPOrAddr    string
		RemoteOS        string
		session         *ssh.Session
		client          *ssh.Client
		autoOpenSession bool
	}

	// ResProcessStatus provides the status
	ResProcessStatus struct {
		Exists bool
		err    error
	}
)

// NewSSHConfig initializes a SSHConfig structure for executing a Run or Copy* command.
func NewSSHConfig(user, KeyFilename, HostIPOrAddr string) (result *SSHConfig) {
	privateKey := []byte{}
	var err error
	privateKey, err = ReadDataFromFile(KeyFilename)
	result = &SSHConfig{
		user:         user,
		HostIPOrAddr: HostIPOrAddr,
	}
	if err == nil {
		result.privateKey = string(privateKey)
	}
	result.EnableAutoOpen()
	return
}

// Clear clears the privateKey, user and host stored in the configuration.
func (sshConfig *SSHConfig) Clear() {
	sshConfig.privateKey = ""
	sshConfig.user = ""
	sshConfig.HostIPOrAddr = ""
}

// Close closes both the session and the connection to the client.
func (sshConfig *SSHConfig) Close() {
	sshConfig.CloseSession()
	sshConfig.CloseClient()
}

// CloseClient closes the client that was opened implicitly during OpenSession.
func (sshConfig *SSHConfig) CloseClient() {
	if sshConfig.client != nil {
		sshConfig.client.Close()
		sshConfig.client = nil
	}
}

// CloseSession closes the session that was opened using OpenSession
func (sshConfig *SSHConfig) CloseSession() {
	if sshConfig.session != nil {
		sshConfig.session.Close()
		sshConfig.session = nil
	}
}

// Connect connects to the given host specified in the configuration
func (sshConfig *SSHConfig) Connect() error {
	clientConfig, err := sshConfig.getClientConfig()
	if err != nil {
		return err
	}

	sshConfig.CloseSession()

	if sshConfig.client == nil {
		sshConfig.client, err = ssh.Dial("tcp", sshConfig.HostIPOrAddr+":22", clientConfig)
		if err != nil {
			return err
		}
	}

	sshConfig.session, err = sshConfig.client.NewSession()
	return err
}

// Copy copies the contents of the specified io.Reader to the given remote location.
// Requires a session to be opened already, unless autoOpenSession is set in the SSHConfig, in which case, Copy connects to the specified host given in the SSHConfig.
// permissions is a string, like 0644, or 0700, etc.
func (sshConfig *SSHConfig) Copy(reader io.Reader, remotePath string, permissions string, size int64) error {
	if sshConfig.session == nil {
		if !sshConfig.autoOpenSession {
			panic("No SSH session opened.")
		}
		err := sshConfig.Connect()
		if err != nil { // Failure to connect. Could be due to invalid host name, or host that cannot be reached.
			return err
		}
	}
	if len(permissions) != 4 {
		return errors.New("permissions need to be 4 characters")
	}

	filename := path.Base(remotePath)
	if filename == "" {
		return errors.New("Remote filename is empty")
	}

	directory := path.Dir(remotePath)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w, _ := sshConfig.session.StdinPipe()
		defer w.Close()
		fmt.Fprintln(w, "C"+permissions, size, filename)
		io.Copy(w, reader)
		fmt.Fprintln(w, "\x00")  // Send 0 byte to indicate EOF
		sshConfig.CloseSession() // A session only accepts one call to Run/Shell, etc, so close the session
	}()

	sshConfig.session.Run("/usr/bin/scp -t " + directory)
	wg.Wait() // waits for the coroutine to complete
	return nil
}

// CopyFile copies the contents of an io.Reader to a remote location, the length is determined by reading the io.Reader until EOF is reached.
// if the file length is known in advance, use "Copy" instead.
func (sshConfig *SSHConfig) CopyFile(fileReader io.Reader, remotePath string, permissions string) error {
	contentBytes, _ := ioutil.ReadAll(fileReader)
	byteReader := bytes.NewReader(contentBytes)

	err := sshConfig.Copy(byteReader, remotePath, permissions, int64(len(contentBytes)))
	return err
}

// CopyFromFile copies the contents of an os.File to a remote location, it will get the length of the file by looking it up from the filesystem.
func (sshConfig *SSHConfig) CopyFromFile(file os.File, remotePath string, permissions string) error {
	stat, err := file.Stat()
	if err == nil {
		err = sshConfig.Copy(&file, remotePath, permissions, stat.Size())
	}
	return err
}

// CopyLocalFileToRemoteFile copies the given local filename to the remote filename with the given permissions
// localFilename must be the filename of a local file and remoteFilename must be the remote filename, not a directory.
func (sshConfig *SSHConfig) CopyLocalFileToRemoteFile(localFilename, remoteFilename, permissions string) error {
	file, err := os.Open(localFilename)
	if err != nil {
		return err
	}
	defer file.Close()
	err = sshConfig.CopyFromFile(*file, remoteFilename, permissions)
	return err
}

// Destroy closes the connection to the client and clears the privatKey, user and host stored in the configuration.
func (sshConfig *SSHConfig) Destroy() {
	sshConfig.Close()
	sshConfig.Clear()
}

func (sshConfig *SSHConfig) DisableAutoOpen() {
	sshConfig.autoOpenSession = false
}

func (sshConfig *SSHConfig) EnableAutoOpen() {
	sshConfig.autoOpenSession = true
}

func (sshConfig *SSHConfig) getClientConfig() (*ssh.ClientConfig, error) {
	key, err := ssh.ParsePrivateKey([]byte(sshConfig.privateKey))
	if err != nil {
		return nil, err
	}
	// Authentication
	config := &ssh.ClientConfig{
		User: sshConfig.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return config, nil
}

// InteractiveSession() must always be followed by a deferred call to sshConfig.Destroy() or
// sshConfig.Close()
func (sshConfig *SSHConfig) InteractiveSession() {
	sshConfig.OpenSession()
}

// Interrupt sends the interrupt signal to the given processName...
func (sshConfig *SSHConfig) Interrupt(processName string) (result string, err error) {
	result, err = sshConfig.Signal(processName, CInt)
	return
}

// OpenSession opens a SSH session
func (sshConfig *SSHConfig) OpenSession() (*ssh.Session, *ssh.Client, error) {
	err := sshConfig.Connect()
	if err != nil {
		return nil, nil, err
	}
	return sshConfig.session, sshConfig.client, nil
}

// ProcessStatus detects if a process is running in the environment specified in the SSHConfig.
func (sshConfig *SSHConfig) ProcessStatus(processName string) *ResProcessStatus {
	cmd := fmt.Sprintf("pgrep -l %s", processName)
	runResult, err := sshConfig.Run(cmd)
	Result := &ResProcessStatus{}
	if err == nil {
		Result.Exists = runResult != ""
	} else {
		Result.err = err
	}
	return Result
}

// Run runs a command on the given SSH environment, usage: output, err := Run("ls")
// Automatically closes the client and session
func (sshConfig *SSHConfig) Run(cmd string) (string, error) {
	session, client, err := sshConfig.OpenSession()
	if err != nil {
		return "", err
	}

	sshConfig.client = nil // remove copy of the client
	defer client.Close()

	sshConfig.session = nil // remove copy of the session
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b // get output
	err = session.Run(cmd)
	return b.String(), err
}

// Signal sends the specified signal to the given processName…
func (sshConfig *SSHConfig) Signal(processName, signal string) (result string, err error) {
	command := fmt.Sprintf("%s -%s %s", CPKill, signal, processName)
	result, err = sshConfig.Run(command)
	return
}
