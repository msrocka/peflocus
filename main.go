package main

import (
	"log"
)

func main() {
	log.SetFlags(0)
	args := ReadArgs()
	switch args.Command {
	case "map":
		NewFlowMapper(args).Run()
	case "merge":
		NewMerger(args).Run()
	case "calc":
		NewCalculator(args).Run()
	default:
		log.Fatalln("ERROR: Unknown command", args.Command)
	}
}
