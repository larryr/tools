// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prez

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/larryr/tools/prez/present"
	"golang.org/x/tools/godoc/static"
	"golang.org/x/tools/playground/socket"
)

var scripts = []string{"jquery.js", "jquery-ui.js", "playground.js", "play.js"}

// playScript registers an HTTP handler at /play.js that serves all the
// scripts specified by the variable above, and appends a line that
// initializes the playground with the specified transport.
func playScript(root, transport string) {
	modTime := time.Now()
	var buf bytes.Buffer
	for _, p := range scripts {
		if s, ok := static.Files[p]; ok {
			buf.WriteString(s)
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(root, "static", p))
		if err != nil {
			panic(err)
		}
		buf.Write(b)
	}
	fmt.Fprintf(&buf, "\ninitPlayground(new %v());\n", transport)
	b := buf.Bytes()
	http.HandleFunc("/play.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/javascript")
		http.ServeContent(w, r, "", modTime, bytes.NewReader(b))
	})
}

func InitPlayground(basepath string, origin *url.URL, nativeClient bool) {
	if present.PlayEnabled {
		if nativeClient {
			socket.RunScripts = false
			socket.Environ = func() []string {
				if runtime.GOARCH == "amd64" {
					return environ("GOOS=nacl", "GOARCH=amd64p32")
				}
				return environ("GOOS=nacl")
			}
		}
		playScript(basepath, "SocketTransport")
		http.Handle("/socket", socket.NewHandler(origin))
	}
}

func playable(c present.Code) bool {
	return present.PlayEnabled && c.Play
}

func environ(vars ...string) []string {
	env := os.Environ()
	for _, r := range vars {
		k := strings.SplitAfter(r, "=")[0]
		var found bool
		for i, v := range env {
			if strings.HasPrefix(v, k) {
				env[i] = r
				found = true
			}
		}
		if !found {
			env = append(env, r)
		}
	}
	return env
}
