// Package filemgr provides provides file read/write functions using external binaries.
// This is especially useful if you want to use e.g. gpg, while letting the user override it
// if necessary.
package filemgr

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	// Default to using gpg
	// use the first private key we find for encryption
	writeDefault = `gpg --yes --encrypt --recipient $(gpg --list-secret-keys --with-colons | head -1 | cut -d: -f5) -o`
	readDefault  = `gpg -d -q`

	// Override default r/w cmd through env
	// e.g.
	// FILE_READER_CMD="gpg -d -q --passphrase XXX" for scripting + gpg
	// FILE_WRITER_CMD="cat >" for plain
	// ...
	writeCmd = writeDefault
	readCmd  = readDefault

	lock sync.Mutex
)

func init() {
	SetWriteCmd(os.Getenv("FILE_WRITER_CMD"))
	SetReadCmd(os.Getenv("FILE_READER_CMD"))
}

// SetWriteCmd overrides the cmd used for writing to files. Lower priority than env FILE_WRITER_CMD (noop if already changed).
func SetWriteCmd(cmd string) {
	lock.Lock()
	if strings.TrimSpace(cmd) != "" && writeCmd == writeDefault {
		writeCmd = cmd
	}
	lock.Unlock()
}

// SetReadCmd overrides the cmd used for reading files. Lower priority than env FILE_READER_CMD (noop if already changed).
func SetReadCmd(cmd string) {
	lock.Lock()
	if strings.TrimSpace(cmd) != "" && readCmd == readDefault {
		readCmd = cmd
	}
	lock.Unlock()
}

// ReadFile reads target file using the configured file reader bin
func ReadFile(path string) (string, error) {

	lock.Lock()
	cmdArgs := []string{"-c", readCmd + " " + path}
	lock.Unlock()

	cmd := exec.Command("sh", cmdArgs...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	defer stderr.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	errout, _ := ioutil.ReadAll(stderr)

	out, _ := ioutil.ReadAll(stdout)

	err = cmd.Wait()

	if err != nil {
		return "", fmt.Errorf("%s: %s", string(errout), err)
	}

	return string(out), nil
}

// WriteFile writes target file (passing content through stdin) using the configured filr writer bin
func WriteFile(path, content string) error {

	lock.Lock()
	cmdArgs := []string{"-c", writeCmd + " " + path}
	lock.Unlock()

	cmd := exec.Command("sh", cmdArgs...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, content)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", string(out), err)
	}

	return nil
}
