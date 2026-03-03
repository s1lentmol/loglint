package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/s1lentmol/loglint/internal/analyzer"
	"github.com/s1lentmol/loglint/internal/config"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var configPath = flag.String("config", "", "path to loglint config file")

func main() {
	path := detectConfigPath(os.Args[1:])
	if path == "" && configPath != nil {
		path = *configPath
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	an, err := analyzer.New(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	singlechecker.Main(an)
}

func detectConfigPath(args []string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-config" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, "-config=") {
			return strings.TrimPrefix(a, "-config=")
		}
	}

	return ""
}
