package template

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/trusch/btrfaas/faas"

	yaml "gopkg.in/yaml.v2"
)

// Build builds a template by:
//  * asserting that function.yaml exists
//  * check for Makefile and execute if found
//  * check for a Dockerfile and build it with the image name from function.yaml
func Build(directory string) error {
	infos, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	if !checkForFunctionYaml(infos) {
		return errors.New("no function.yaml in template")
	}

	if checkForMakefile(infos) {
		script := "cd %v; make"
		script = fmt.Sprintf(script, directory)
		cmd := exec.Command("bash", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return err
		}
	}

	if checkForDockerfile(infos) {
		imageName, err := getImageNameFromFunctionYaml(directory)
		if err != nil {
			return err
		}
		script := "cd %v; docker build -t %v ."
		script = fmt.Sprintf(script, directory, imageName)
		cmd := exec.Command("bash", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func checkForFunctionYaml(infos []os.FileInfo) bool {
	for _, info := range infos {
		if info.Name() == "function.yaml" {
			return true
		}
	}
	return false
}

func checkForDockerfile(infos []os.FileInfo) bool {
	for _, info := range infos {
		if info.Name() == "Dockerfile" {
			return true
		}
	}
	return false
}

func checkForMakefile(infos []os.FileInfo) bool {
	for _, info := range infos {
		if info.Name() == "Makefile" {
			return true
		}
	}
	return false
}

func getImageNameFromFunctionYaml(directory string) (string, error) {
	bs, err := ioutil.ReadFile(filepath.Join(directory, "function.yaml"))
	if err != nil {
		return "", err
	}
	spec := &faas.DeployFunctionOptions{}
	if err = yaml.Unmarshal(bs, spec); err != nil {
		return "", err
	}
	return spec.Image, nil
}
