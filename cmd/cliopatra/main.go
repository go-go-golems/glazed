package main

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

type Program struct {
	Name   string   `yaml:"name"`
	Args   []string `yaml:"args"`
	Stdin  string   `yaml:"stdin"`
	Stdout string   `yaml:"stdout"`
	Stderr string   `yaml:"stderr"`
}

func saveArgsAndFlagsToYAML(program Program, fileName string) error {
	data, err := yaml.Marshal(program)
	if err != nil {
		return err
	}
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a program name and its flags and arguments.")
		os.Exit(1)
	}
	programName := os.Args[1]
	args := os.Args[2:]

	path, err := exec.LookPath(programName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd := exec.Command(path, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_ = stderr

	eg := errgroup.Group{}
	// copy stdin to stdinpipe in a goroutine
	eg.Go(func() error {
		// TODO(manuel, 2023-02-12) Capture stdin when running cliopatra
		//
		// See https://github.com/go-go-golems/glazed/issues/131
		n, err := io.Copy(stdin, os.Stdin)
		fmt.Printf("Copied %d bytes to stdin\n", n)
		if err != nil {
			return err
		}

		return nil
	})

	// run the main goroutine that captures stdout and stderr and waits for the program
	// to finish

	inStr := make([]byte, 0)
	outStr := make([]byte, 0)
	errStr := make([]byte, 0)

	eg.Go(func() error {
		fmt.Println("Start cmd reading loop")

		for {
			tempOut := make([]byte, 100)
			fmt.Println("Start reading stdout and stderr")
			n, err := stdout.Read(tempOut)
			fmt.Printf("Read %d bytes from stdout: %s\n", n, string(tempOut))
			if err != nil || n == 0 {
				break
			}
			outStr = append(outStr, tempOut...)
			//tempErr := make([]byte, 100)
			//fmt.Println("Start reading stderr")
			//n, err = stderr.Read(tempErr)
			//fmt.Printf("Read %d bytes from stderr: %s\n", n, string(tempErr))
			//if err != nil || n == 0 {
			//	break
			//}
			//errStr = append(errStr, tempErr...)
		}

		fmt.Println("Finished copying stdout and stderr")

		return nil
	})

	eg.Go(func() error {
		if err := cmd.Start(); err != nil {
			return err
		}

		fmt.Println("Start cmd.Wait()")
		err := cmd.Wait()
		fmt.Printf("cmd.Wait() returned %v\n", err)
		return err
	})

	if err := eg.Wait(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	program := Program{
		Name:   programName,
		Args:   args,
		Stdin:  string(inStr),
		Stdout: string(outStr),
		Stderr: string(errStr),
	}

	fileName := strings.Join([]string{programName, ".yaml"}, "")
	if err := saveArgsAndFlagsToYAML(program, fileName); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("The program has been run successfully, and the results have been saved to the file", fileName)
}
