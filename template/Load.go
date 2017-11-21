package template

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

// Load clones a template into a directory
func Load(uri, targetFolder string) error {
	if !strings.HasPrefix(uri, "http") && !strings.HasPrefix(uri, "git") {
		// no valid url, try prefixing with default btrfaas template url
		uri = "https://github.com/trusch/btrfaas.git/templates/" + uri
	}
	repo := uri
	path := "."
	if gitIdx := strings.Index(uri, ".git"); gitIdx != -1 {
		repo = uri[:gitIdx+4]
		path = uri[gitIdx+4:]
	}
	tmpDir := filepath.Join(os.TempDir(), "btrfaas-work")
	defer os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return err
	}
	_, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:      repo,
		Depth:    1,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}
	if err = os.Rename(filepath.Join(tmpDir, path), targetFolder); err != nil {
		return err
	}
	if err = replaceFunctionName(filepath.Join(targetFolder, "function.yaml"), filepath.Base(targetFolder)); err != nil {
		return err
	}
	if err = replaceFunctionName(filepath.Join(targetFolder, "Dockerfile"), filepath.Base(targetFolder)); err != nil {
		return err
	}
	return nil
}

func replaceFunctionName(file, functionName string) error {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	tmpl, err := template.New(file).Parse(string(bs))
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err = tmpl.Execute(buf, &struct{ FunctionName string }{functionName}); err != nil {
		return err
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, buf)
	return err
}
