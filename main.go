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
	case "unmap":
		NewFlowUnmapper(args).Run()
	case "merge":
		NewMerger(args).Run()
	case "model-check":
		modelCheck(args)
	default:
		log.Fatalln("ERROR: Unknown command", args.Command)
	}
}
