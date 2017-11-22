package main

import (
	"log"
	"os"
)

// Args contains the command line arguments of application.
type Args struct {
	Command string
	WorkDir string
	MapFile string
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
		if flag == "" {
			flag = val
			continue
		}
		switch flag {
		case "-workdir":
			args.WorkDir = val
		case "-mapfile":
			args.MapFile = val
		default:
			flag = val
		}
	}

	return &args
}
