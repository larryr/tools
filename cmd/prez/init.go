package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// De-branded templates from this repo.
	templateBase = "https://raw.githubusercontent.com/larryr/tools/main/cmd/prez/templates/"
	// Static assets (JS/CSS) from the upstream present tool.
	staticBase = "https://raw.githubusercontent.com/golang/tools/master/cmd/present/static/"
)

var templateFiles = []string{"action.tmpl", "slides.tmpl", "article.tmpl", "dir.tmpl"}

var staticFiles = []string{
	"styles.css", "dir.css", "article.css", "notes.css",
	"slides.js", "dir.js", "notes.js", "play.js",
	"playground.js", "jquery.js", "jquery-ui.js",
}

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
	for _, d := range []string{"templates", "static"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			return err
		}
	}

	fmt.Println("Fetching templates...")
	for _, f := range templateFiles {
		dest := filepath.Join(dir, "templates", f)
		if err := download(templateBase+f, dest); err != nil {
			return fmt.Errorf("template %s: %v", f, err)
		}
	}

	fmt.Println("Fetching static assets...")
	for _, f := range staticFiles {
		dest := filepath.Join(dir, "static", f)
		if err := download(staticBase+f, dest); err != nil {
			return fmt.Errorf("static %s: %v", f, err)
		}
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

func download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
