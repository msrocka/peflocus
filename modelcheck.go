package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/msrocka/ilcd"
)

func modelCheck(args *Args) {
	for _, zipName := range GetZipNames(args.WorkDir) {
		zipPath := filepath.Join(args.WorkDir, zipName)
		reader, err := ilcd.NewZipReader(zipPath)
		if err != nil {
			log.Println("ERROR: Could not read ILCD package", zipPath)
			continue
		}
		fmt.Println("\nCheck models in", zipPath)
		reader.EachModel(func(model *ilcd.Model) bool {
			printModelReport(model, reader)
			printGraph(model)
			return true
		})
	}

}

func printModelReport(model *ilcd.Model, reader *ilcd.ZipReader) {
	fmt.Println("\nCheck model", model.FullName("en"))

	if model.RefProcess() == nil {
		fmt.Println("  .. error: the reference process does not exist")
	}

	// read and check the processes
	processes := make(map[int]*ilcd.Process)
	for _, pi := range model.Processes {
		if pi.Process == nil {
			fmt.Println("  .. error: no process ref. in", pi.InternalID)
			continue
		}
		zfile := reader.FindDataSet(ilcd.ProcessDataSet, pi.Process.UUID)
		if zfile == nil {
			fmt.Println("  .. error: process with ID=",
				pi.Process.UUID, "does not exits")
			continue
		}
		process, err := zfile.ReadProcess()
		if err != nil {
			fmt.Println("  .. error: failed to read process ID=", pi.Process.UUID)
			continue
		}
		processes[pi.InternalID] = process
	}

	// check the connections
	for _, pi := range model.Processes {
		provider := processes[pi.InternalID]
		if provider == nil {
			continue
		}
		fmt.Println("  .. info: check process internalID=",
			pi.InternalID, "UUID=", provider.UUID())
		for _, con := range pi.Connections {
			output := findExchange(provider, con.OutputFlow, "Output")
			if output == nil {
				fmt.Println("  .. error: process ID=", pi.Process.UUID,
					"has no output with flow", con.OutputFlow)
				continue
			}
			for _, link := range con.Links {
				recipient := processes[link.ProcessID]
				if recipient == nil {
					fmt.Println("  .. error: process with internalID=",
						link.ProcessID, "does not exist")
					continue
				}
				input := findExchange(recipient, link.InputFlow, "Input")
				if input == nil {
					fmt.Println("  .. error: process ID=", recipient.UUID(),
						" has no input of flow", link.InputFlow)
				}
			}
		}
	}
}

func findExchange(process *ilcd.Process,
	flowID, direction string) *ilcd.Exchange {
	for i := range process.Exchanges {
		e := &process.Exchanges[i]
		if direction != e.Direction {
			continue
		}
		if e.Flow != nil && e.Flow.UUID == flowID {
			return e
		}
	}
	return nil
}

func printGraph(model *ilcd.Model) {
	fmt.Println("\n  .. The model graph:")
	fmt.Println("\n  digraph G {")

	if ref := model.RefProcess(); ref != nil {
		fmt.Println("  ", ref.InternalID, "[fillcolor=pink style=filled]")
	}

	for _, pi := range model.Processes {
		for _, con := range pi.Connections {
			for _, link := range con.Links {
				fmt.Println("  ", pi.InternalID, "->", link.ProcessID)
			}
		}
	}
	fmt.Println("  }")
}
