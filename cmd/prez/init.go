package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed templates static
var content embed.FS

const sampleSlide = `Title of Presentation
Subtitle or tagline
15:04 2 Jan 2006

Author Name
author@example.com

* Introduction

Welcome to your new presentation.

- Edit this file to add your content
- Use .image, .code, .link directives
- Press 'N' for presenter notes

: These are presenter notes (hidden by default).

* Next Steps

Add more slides with the * prefix.

  // Preformatted code blocks are indented
  fmt.Println("Hello, World!")

`

func initPresentation(dir string) error {
	err := fs.WalkDir(content, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || path == "." {
			return err
		}
		dest := filepath.Join(dir, path)
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, err := content.ReadFile(path)
		if err != nil {
			return err
		}
		fmt.Println(" ", path)
		return os.WriteFile(dest, data, 0644)
	})
	if err != nil {
		return err
	}

	slidePath := filepath.Join(dir, "sample.slide")
	if err := os.WriteFile(slidePath, []byte(sampleSlide), 0644); err != nil {
		return err
	}

	fmt.Println("Presentation initialized in:", dir)
	fmt.Printf("  Edit %s then run:\n", slidePath)
	fmt.Printf("  prez -base %s\n", dir)
	return nil
}
