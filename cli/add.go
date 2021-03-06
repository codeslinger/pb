package cli

import "bytes"
import "fmt"
import "io"
import "io/ioutil"
import "log"
import "os"
import "strings"

import "github.com/lead-scm/pb/lib"
import "github.com/mitchellh/cli"

// AddCommand is the controller for staging files to the working area. Right now
// you have to add one file at a time, but in the future (TODO) you can specify
// whole directories.
type AddCommand struct {
	UI cli.Ui
}

// Help displays explanitory text for the AddCommand.
func (c *AddCommand) Help() string {
	return "Add a file to the working index"
}

// Synopsis is aliased to Help.
func (c *AddCommand) Synopsis() string {
	return c.Help()
}

// Run stages a given file to the repository.
func (c *AddCommand) Run(args []string) int {
	path := args[0]
	Add(path)
	return 0
}

// Add a file to the working index. If the file location is already in the
// working index, update the blob reference for that location.
func Add(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, f)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	// Persist file as blob in object store
	b := &lib.Blob{Contents: string(buf.Bytes())}
	err = b.Put()
	if err != nil {
		log.Fatal(err)
	}

	index, err := os.Open("./.pb/index")
	if err != nil {
		log.Fatal(err)
	}

	indexBuf := bytes.NewBuffer(nil)
	_, err = io.Copy(indexBuf, index)
	if err != nil {
		log.Fatal(err)
	}
	indexContents := string(indexBuf.Bytes())
	index.Close()

	lines := strings.Split(indexContents, "\n")
	itemFound := false
	newLines := make([]string, 0, len(lines))

	for _, line := range lines {
		tokens := strings.Split(line, " ")

		if tokens[0] == path {
			newLines = append(newLines, formatIndexLine(path, b))
			itemFound = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !itemFound {
		newLines = append(newLines, formatIndexLine(path, b))
	}

	newContents := strings.TrimLeft(strings.Join(newLines, "\n"), "\n")
	ioutil.WriteFile("./.pb/index", []byte(newContents), 0666)
}

func formatIndexLine(path string, b *lib.Blob) string {
	return fmt.Sprintf("%s %s", path, b.Hash())
}
