// lkan is a small kanban board with a YAML source of truth and a local
// web UI.
//
// Usage:
//
//	lkan init               Create a starter board.yaml in the current dir.
//	lkan serve              Start the web UI (default if no subcommand).
//	lkan -f path -http addr Override file path and listen address.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/larryr/tools/lkan"
)

var (
	file     = flag.String("f", "board.yaml", "path to board YAML file")
	httpAddr = flag.String("http", "127.0.0.1:8080", "HTTP listen address")
	watch    = flag.Bool("watch", false, "watch -f for changes, auto-reload browser (serve only)")
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: lkan [flags] <subcommand>

subcommands:
  serve   start the web UI (default)
  init    write a starter board.yaml to -f path
  usage   print full adoption guide and board.yaml schema

flags:
`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := "serve"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	switch cmd {
	case "init":
		if err := lkan.WriteStarter(*file); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("wrote %s\n", *file)
	case "serve":
		if err := serve(*file, *httpAddr, *watch); err != nil {
			log.Fatal(err)
		}
	case "usage":
		if err := lkan.Usage(os.Stdout); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func serve(path, addr string, watch bool) error {
	store, err := lkan.NewStore(path)
	if err != nil {
		return err
	}
	var opts []lkan.Option
	if watch {
		wr, err := lkan.NewWatcher(path)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			if err := wr.Run(ctx); err != nil {
				log.Printf("lkan watch stopped: %v", err)
			}
		}()
		opts = append(opts, lkan.WithWatcher(wr))
		log.Printf("lkan: watching %s for changes", path)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("lkan: serving %s — open http://%s", path, ln.Addr())
	return http.Serve(ln, lkan.NewServer(store, opts...))
}
