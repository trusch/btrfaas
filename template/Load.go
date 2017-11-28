package template

import (
	"bytes"
	"fmt"
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
	if err = copyDir(filepath.Join(tmpDir, path), targetFolder); err != nil {
		return err
	}
	if err = replaceFunctionName(filepath.Join(targetFolder, "function.yaml"), filepath.Base(targetFolder)); err != nil {
		return err
	}
	return replaceFunctionName(filepath.Join(targetFolder, "Dockerfile"), filepath.Base(targetFolder))
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

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
