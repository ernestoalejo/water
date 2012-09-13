package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if err := testFiles(); err != nil {
		fmt.Println(err)
	}
}

func testFiles() error {
	files, err := ioutil.ReadDir("test")
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := testFile(file.Name()); err != nil {
			return err
		}
	}

	return nil
}

func testFile(file string) error {
	log.Println("Running", file)

	f, err := os.Open("test/" + file)
	if err != nil {
		return err
	}
	defer f.Close()

	code, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	app := string(code)
	parts := strings.Split(app, "\n###########################################################\n\n")
	if len(parts) != 2 {
		return fmt.Errorf("file doesn't have the test section: %s", file)
	}

	cmd := exec.Command("water")

	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	io.Copy(in, bytes.NewBufferString(parts[0]))

	in.Close()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	if string(output) != parts[1] {
		return fmt.Errorf("bad output in the %s program.\n\nOUTPUT:\n%s\n\nEXPECTED:\n%s",
			file, output, parts[1])
	}

	return nil
}
