package main

import (
	"log"
	"os"
	"strings"
)

// Args contains the command line arguments of application.
type Args struct {
	Command  string
	WorkDir  string
	MapFile  string
	SkipDocs string
}

// ReadArgs reads the command line arguments.
func ReadArgs() *Args {

	if len(os.Args) < 2 {
		log.Fatalln("ERROR: No command given. (Usage: peflocus <command> <option>)")
	}

	args := Args{
		Command: os.Args[1],
		WorkDir: "zips",
		MapFile: "flow_mapping.csv"}

	flag := ""
	for i, val := range os.Args {
		if i < 2 {
			continue
		}
		if strings.HasPrefix(val, "-") {
			flag = val
			continue
		}
		if flag == "" {
			continue
		}
		switch flag {
		case "-workdir":
			args.WorkDir = val
		case "-mapfile":
			args.MapFile = val
		case "-skipdocs":
			args.SkipDocs = val
		}
		flag = ""
	}

	return &args
}
