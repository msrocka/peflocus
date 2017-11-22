package main

import "log"

// FlowMapper applies a flow mapping to the ILCD packages in a working directory.
type FlowMapper struct {
	workdir string
	mapfile string

	flowMap *FlowMap
}

// NewFlowMapper initializes a new flow mapper from the given arguments.
func NewFlowMapper(args *Args) *FlowMapper {
	return &FlowMapper{workdir: args.WorkDir, mapfile: args.MapFile}
}

// Run executes the flow mapping.
func (m *FlowMapper) Run() {
	flowMap := ReadFlowMap()

	zips := getZips()
	if len(zips) == 0 {
		log.Println("No zip files found.")
		return
	}

	log.Println("Found", len(zips), "zip files for conversion")
	for _, name := range zips {
		log.Println("Convert zip file", name)
		runConversion(name, flowMap)
	}
}
