package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/msrocka/ilcd"
)

// FlowUnmapper reverses an applied mapping
type FlowUnmapper struct {
	workdir string
	mapfile string

	flowMap *FlowMap
}

// NewFlowUnmapper initializes a new flow unmapper from the given arguments.
func NewFlowUnmapper(args *Args) *FlowUnmapper {
	return &FlowUnmapper{workdir: args.WorkDir, mapfile: args.MapFile}
}

// Run executes the flow un-mapping.
func (u *FlowUnmapper) Run() {
	u.flowMap = ReadFlowMap(u.mapfile)
	zips := GetZipNames(u.workdir)
	for _, name := range zips {
		sourcePath := filepath.Join(u.workdir, name)
		targetPath := filepath.Join(u.workdir, "peflocus_unmapped_"+name)
		DeleteExisting(targetPath)
		log.Println("INFO: map flows in", sourcePath, "to", targetPath)
		u.doIt(sourcePath, targetPath)
	}
}

func (u *FlowUnmapper) doIt(sourcePath, targetPath string) {

	// create the reader and writer
	reader, err := ilcd.NewZipReader(sourcePath)
	if err != nil {
		log.Println("ERROR: Failed to read zip", sourcePath, ":", err)
		return
	}
	defer reader.Close()
	writer, err := ilcd.NewZipWriter(targetPath)
	if err != nil {
		log.Println("ERROR: Failed to create zip writer for",
			targetPath, ":", err)
		return
	}
	defer writer.Close()

	// unmap the flows in the data sets
	flowFolder := ""
	reader.Map(writer, func(zipFile *ilcd.ZipFile) (string, []byte) {
		path := zipFile.Path()
		if flowFolder == "" && ilcd.IsFlowPath(path) {
			flowFolder = strings.Split(path, "flows")[0] + "flows/"
		}
		if zipFile.Type() == ilcd.FlowDataSet {
			return "", nil // flows are filtered & written later
		}
		data, err := zipFile.Read()
		if err != nil {
			log.Println("ERROR: Failed to read entry", path, err)
			return "", nil
		}
		converted, err := u.flowMap.UnmapFlows(path, data)
		if err != nil {
			log.Println("ERROR: Failed to umap flows in", path, err)
			return "", nil
		}
		return path, converted
	})

	gen := FlowGenerator{
		flowMap:   u.flowMap,
		folder:    flowFolder,
		reader:    reader,
		writer:    writer,
		forMapped: false}
	gen.Generate()

	// copy the flows that were not mapped but are used
	log.Println("INFO: Copy untouched but used flows")
	count := 0
	reader.Map(writer, func(zipFile *ilcd.ZipFile) (string, []byte) {
		if zipFile.Type() != ilcd.FlowDataSet {
			return "", nil
		}
		data, err := zipFile.Read()
		if err != nil {
			log.Println("ERROR: Failed to read flow", zipFile.Path())
			return "", nil
		}
		flow, err := zipFile.ReadFlow()
		if err != nil {
			log.Println("ERROR: Failed to read flow", zipFile.Path())
			return "", nil
		}
		uuid := flow.UUID()
		if !u.flowMap.untouchedUsed[uuid] {
			// skip all flows that where mapped or that are not used
			return "", nil
		}
		path := flowFolder + uuid + "_" + flow.Version() + ".xml"
		count++
		return path, data
	})
	log.Println(" ... copied", count, "flows")

	u.flowMap.ResetStats()
}
