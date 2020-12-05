// Package editor lets the user edit either arbitrary text or a JSON representation of an arbitrary object
// in their text editor of choice (default is vi).
package editor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/google/uuid"
)

// Edit runs the user's text editor with arbitrary initial content,
// then it lets the user edit it and returns the edited result.
func Edit(content string) (string, error) {
	filename := fmt.Sprintf("/tmp/tmpeditor%s", uuid.Must(uuid.NewRandom()).String())

	err := ioutil.WriteFile(filename, []byte(content), 0600)
	if err != nil {
		return "", err
	}

	defer os.Remove(filename)

	editorvar := os.Getenv("EDITOR")
	if editorvar == "" {
		editorvar = "vi"
	}

	editorcmd := exec.Command(editorvar, filename)
	editorcmd.Stdin = os.Stdin
	editorcmd.Stdout = os.Stdout

	err = editorcmd.Run()
	if err != nil {
		return "", err
	}

	edited, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(edited), nil
}

// EditJSON runs the user's text editor with a JSON representation of input parameter i,
// then it lets the user edit it and unmarshals the edited result back into i.
func EditJSON(i interface{}) error {
	return EditJSONTarget(i, i)
}

// EditJSONTarget is similar to EditJSON but lets you specify a different target for unmarshaling.
func EditJSONTarget(i interface{}, target interface{}) error {
	j, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return err
	}
	var correctContent bool
	for !correctContent {
		edited, err := Edit(string(j))
		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(edited), target)
		if err == nil {
			correctContent = true
			break
		}

		j = []byte(edited)
		fmt.Printf("failed to unmarshal edited content: %s.\nPress enter to edit", err)
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
	return nil
}
