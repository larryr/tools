// lkan is a small kanban board with a YAML source of truth and a local
// web UI.
//
// Usage:
//
//	lkan init                          Create a starter board.yaml in the current dir.
//	lkan serve                         Start the web UI (default if no subcommand).
//	lkan usage                         Print the agent-facing guide to stdout.
//	lkan usage --update-agents-md      Insert/replace the lkan block in ./AGENTS.md.
//	lkan -f path -http addr            Override file path and listen address.
package main

import (
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
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `usage: lkan [-f board.yaml] [-http 127.0.0.1:8080] <subcommand>

subcommands:
  serve   start the web UI (default)
  init    write a starter board.yaml to -f path
  usage   print the agent-facing guide to stdout
          --update-agents-md [path]  insert/replace the lkan block in AGENTS.md
                                     (default path: ./AGENTS.md)
`)
}

func main() {
	flag.Usage = printUsage
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
		if err := serve(*file, *httpAddr); err != nil {
			log.Fatal(err)
		}
	case "usage":
		runUsage(flag.Args()[1:])
	default:
		printUsage()
		os.Exit(2)
	}
}

func runUsage(args []string) {
	fs := flag.NewFlagSet("usage", flag.ExitOnError)
	update := fs.Bool("update-agents-md", false, "insert/replace the lkan block in an AGENTS.md file")
	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}
	if !*update {
		fmt.Print(lkan.AgentGuide)
		return
	}
	path := "AGENTS.md"
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}
	action, err := lkan.UpdateAgentsMD(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s\n", action, path)
}

func serve(path, addr string) error {
	store, err := lkan.NewStore(path)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("lkan: serving %s — open http://%s", path, ln.Addr())
	return http.Serve(ln, lkan.NewServer(store))
}
