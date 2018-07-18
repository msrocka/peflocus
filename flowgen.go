package main

import (
	"errors"
	"log"
	"strings"

	"github.com/beevik/etree"
	"github.com/msrocka/ilcd"
)

// FlowGenerator generates the new flow data sets after a mapping / unmapping.
type FlowGenerator struct {
	// the flow folder under which the generated flow data sets should be
	// stored.
	folder string

	// the zip reader from where this generator reads the original data
	reader *ilcd.ZipReader

	// the zip writer where this generator generates the new flows.
	writer *ilcd.ZipWriter

	// the flow map which the generator uses
	flowMap *FlowMap

	// indicates whether this generator should generate the mapped or unmapped flows.
	forMapped bool
}

type genFlowInfo struct {
	sourceID string
	targetID string
	location string
}

// Generate creates the mapped flow in the target package.
func (gen *FlowGenerator) Generate() {
	log.Println("INFO: Generate new flows")
	generated := make(map[string]bool)
	for usedKey := range gen.flowMap.used {

		genInfo := gen.genInfo(usedKey)
		if genInfo == nil {
			log.Println(" ... WARNING: did not find a mapping for", usedKey)
			continue
		}
		if generated[genInfo.targetID] {
			continue
		}

		flowEntry := gen.reader.FindDataSet(ilcd.FlowDataSet, genInfo.sourceID)
		if flowEntry == nil {
			log.Println(" ... ERROR: flow", genInfo.sourceID,
				"mapped and used but could not find it in package")
			continue
		}
		data, err := flowEntry.Read()
		if err != nil {
			log.Println(" ... ERROR: Failed to read flow", flowEntry.Path(), err)
			continue
		}

		data, err = gen.doIt(genInfo, data)
		if err != nil {
			log.Println(" ... ERROR: Failed to create flow", genInfo.targetID,
				"from", genInfo.sourceID, err)
			continue
		}

		newEntry := gen.folder + genInfo.targetID + ".xml"
		if err = gen.writer.Write(newEntry, data); err != nil {
			log.Println(" ... ERROR: Failed to write new flow", newEntry, err)
			continue
		}
		generated[genInfo.targetID] = true

	}
	log.Println(" ... generated", len(generated), "new flows")
}

func (gen *FlowGenerator) genInfo(usedKey string) *genFlowInfo {
	if gen.flowMap == nil {
		return nil
	}
	var entry *FlowMapEntry
	if gen.forMapped {
		entry = gen.flowMap.mappings[usedKey]
	} else {
		entry = gen.flowMap.unmappings[usedKey]
	}
	if entry == nil {
		return nil
	}
	if gen.forMapped {
		return &genFlowInfo{
			sourceID: entry.OldID,
			targetID: entry.NewID,
			location: entry.Location}
	}
	return &genFlowInfo{
		sourceID: entry.NewID,
		targetID: entry.OldID}

}

// if location="" it is expected that this function runs in `unmapped` mode,
// which will remove flow locations from the name; otherwise the location code
// is added to the name -> this is a quickly hack to get this done o_O
func (gen *FlowGenerator) doIt(genInfo *genFlowInfo, data []byte) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}
	uuid := doc.FindElement("./flowDataSet/flowInformation/dataSetInformation/UUID")
	if uuid == nil {
		return nil, errors.New("No UUID element")
	}
	uuid.SetText(genInfo.targetID)
	npath := "./flowDataSet/flowInformation/dataSetInformation/name/baseName"
	for _, elem := range doc.FindElements(npath) {
		name := elem.Text()
		if gen.forMapped && genInfo.location != "" {
			elem.SetText(name + " - " + genInfo.location)
		} else if !gen.forMapped {
			locIdx := strings.LastIndex(name, " - ")
			if locIdx > 0 {
				name = name[:locIdx]
				elem.SetText(name)
			}
		}
	}
	return doc.WriteToBytes()
}
