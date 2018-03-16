package main

import (
	"encoding/xml"
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
	flowPrefix := ""
	reader.Map(writer, func(zipFile *ilcd.ZipFile) (string, []byte) {
		path := zipFile.Path()
		if flowPrefix == "" && ilcd.IsFlowPath(path) {
			flowPrefix = strings.Split(path, "flows")[0] + "flows/"
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

	gen := FlowGen{
		flowMap: u.flowMap,
		prefix:  flowPrefix,
		reader:  reader,
		writer:  writer}
	gen.Generate(false)

	// copy the flows that were not mapped
	reader.Map(writer, func(zipFile *ilcd.ZipFile) (string, []byte) {
		if zipFile.Type() != ilcd.FlowDataSet {
			return "", nil
		}
		data, err := zipFile.Read()
		if err != nil {
			log.Println("ERROR: Failed to read flow", zipFile.Path())
			return "", nil
		}
		flow := &ilcd.Flow{}
		if err = xml.Unmarshal(data, flow); err != nil {
			log.Println("ERROR: Failed to read flow", zipFile.Path())
			return "", nil
		}
		uuid := flow.UUID()
		if u.flowMap.used[uuid] {
			// skip unmapped flows
			return "", nil
		}
		path := flowPrefix + uuid + "_" + flow.Version() + ".xml"
		return path, data
	})

	u.flowMap.ResetStats()
}
