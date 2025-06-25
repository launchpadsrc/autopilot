package poppler

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/samber/lo"
)

func ToHTML(r io.Reader) (string, error) {
	path := filepath.Join(os.TempDir(), lo.RandomString(16, lo.AlphanumericCharset))

	// Poppler generates two files: `*s.html` and `*-html.html`, where `*` is a given file name.
	// The actual HTML content is in the `*-html.html` file.
	inputPath := path + ".html"
	actualPath := path + "-html.html"
	sPath := path + "s.html"
	defer os.Remove(actualPath)
	defer os.Remove(sPath)

	cmd := exec.Command(
		"pdftohtml",
		"-i",       // ignore images
		"-s",       // generate single document that includes all pages
		"-nomerge", // do not merge paragraphs
		"-",        // read from stdin
		inputPath,
	)

	cmd.Stdin = r
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	if stderr.Len() > 0 {
		return "", errors.New(stderr.String())
	}

	data, err := os.ReadFile(actualPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
