package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	mdserve "mdserve"
	"mdserve/internal/server"
)

func main() {
	port := flag.Int("port", 3333, "listen port")
	noWatch := flag.Bool("no-watch", false, "disable file watching and live reload")
	flag.Parse()

	dir := flag.Arg(0)
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot get current directory: %v\n", err)
			os.Exit(1)
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot resolve path: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: directory does not exist: %s\n", absDir)
		os.Exit(1)
	}

	cfg := server.Config{
		DocRoot:  absDir,
		Port:     *port,
		NoWatch:  *noWatch,
		AssetsFS: mdserve.Assets,
	}

	if err := server.New(cfg).Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
