package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/msrocka/ilcd"
)

func getKey(location, uuid string) string {
	key := strings.TrimSpace(location) + "/" + strings.TrimSpace(uuid)
	return strings.ToLower(key)
}

// FlowMapEntry represents a row in the mapping file
type FlowMapEntry struct {
	Location string
	OldID    string
	NewID    string
}

func (e *FlowMapEntry) mapKey() string {
	return getKey(e.Location, e.OldID)
}

// FlowMap contains the flow mappings.
type FlowMap struct {

	// (location/OldID) -> map entry
	mappings map[string]*FlowMapEntry

	// NewID -> map entry
	unmappings map[string]*FlowMapEntry

	// When running in map-mode: contains (location/OldID) -> bool
	// When running in unmap-mode: contains NewID -> true
	used map[string]bool
}

// ReadFlowMap reads the flow mappings from the given file.
func ReadFlowMap(file string) *FlowMap {
	log.Println("INFO: Read flow mappings from", file)
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln("ERROR: Failed to read mapping file", file, err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalln("ERROR: Failed to read mapping file", file, err)
	}
	fm := FlowMap{
		mappings:   make(map[string]*FlowMapEntry),
		unmappings: make(map[string]*FlowMapEntry),
		used:       make(map[string]bool)}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		e := FlowMapEntry{OldID: row[0], Location: row[1], NewID: row[2]}
		fm.mappings[e.mapKey()] = &e
		fm.unmappings[e.NewID] = &e
	}
	log.Println(" ... read", len(fm.mappings), "mappings")
	return &fm
}

// ResetStats clears the mapping statistics
func (m *FlowMap) ResetStats() {
	m.used = make(map[string]bool)
}

// MapFlows maps the flows in the given data set if it is an LCIA method or
// process
func (m *FlowMap) MapFlows(zipEntry string, data []byte) ([]byte, error) {
	if ilcd.IsMethodPath(zipEntry) {
		return m.forMethod(data, m.mapFlow)
	}
	if ilcd.IsProcessPath(zipEntry) {
		return m.forProcess(data, m.mapFlow)
	}
	return data, nil
}

// UnmapFlows applies a reverse mapping: reasigning the old flow UUIDs.
func (m *FlowMap) UnmapFlows(zipEntry string, data []byte) ([]byte, error) {
	if ilcd.IsMethodPath(zipEntry) {
		return m.forMethod(data, m.unmapFlow)
	}
	if ilcd.IsProcessPath(zipEntry) {
		return m.forProcess(data, m.unmapFlow)
	}
	return data, nil
}

func (m *FlowMap) forMethod(data []byte, fn func(e *etree.Element)) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}
	uuidElem := doc.FindElement("./LCIAMethodDataSet/LCIAMethodInformation/dataSetInformation/UUID")
	if uuidElem != nil {
		log.Println("Replace flows in LCIA method", uuidElem.Text())
	}
	factors := doc.FindElements("./LCIAMethodDataSet/characterisationFactors/factor")
	log.Println(" ... check", len(factors), "factors")
	for _, factor := range factors {
		m.mapFlow(factor)
	}
	return doc.WriteToBytes()
}

func (m *FlowMap) forProcess(data []byte, fn func(e *etree.Element)) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}
	uuidElem := doc.FindElement("./processDataSet/processInformation/dataSetInformation/UUID")
	if uuidElem != nil {
		log.Println("Replace flows in process", uuidElem.Text())
	}
	exchanges := doc.FindElements("./processDataSet/exchanges/exchange")
	log.Println(" ... check", len(exchanges), "exchanges")
	for _, e := range exchanges {
		fn(e)
	}
	return doc.WriteToBytes()
}

// mapFlow assigns the new flow UUIDs from the mapping to exchanges or LCIA
// factors with a matching pair of old flow UUID and location.
func (m *FlowMap) mapFlow(e *etree.Element) {
	locElem := e.FindElement("./location")
	if locElem == nil {
		return
	}
	location := strings.TrimSpace(locElem.Text())
	if location == "" {
		return
	}
	flowRef := e.FindElement("./referenceToFlowDataSet")
	if flowRef == nil {
		log.Println(" ... ERROR: no flow reference found")
		return
	}
	idAttr := flowRef.SelectAttr("refObjectId")
	if idAttr == nil {
		log.Println(" ... ERROR: no flow reference found")
		return
	}
	key := getKey(location, idAttr.Value)
	mapping := m.mappings[key]
	if mapping == nil {
		log.Println(" ... ERROR: Missing flow mapping for",
			idAttr.Value, " -> ", location)
		return
	}
	idAttr.Value = mapping.NewID
	uriAttr := flowRef.SelectAttr("uri")
	if uriAttr != nil {
		uriAttr.Value = "../flows/" + mapping.NewID + ".xml"
	}
	m.used[key] = true
}

// unmapFlow assigns back the old flow UUID to exchanges and LCIA factors
// that have a new flow UUID.
func (m *FlowMap) unmapFlow(e *etree.Element) {
	flowRef := e.FindElement("./referenceToFlowDataSet")
	if flowRef == nil {
		log.Println(" ... ERROR: no flow reference found")
		return
	}
	idAttr := flowRef.SelectAttr("refObjectId")
	if idAttr == nil {
		log.Println(" ... ERROR: no flow reference found")
		return
	}
	unmapping := m.unmappings[idAttr.Value]
	if unmapping == nil {
		return
	}
	idAttr.Value = unmapping.OldID
	uriAttr := flowRef.SelectAttr("uri")
	if uriAttr != nil {
		uriAttr.Value = "../flows/" + unmapping.OldID + ".xml"
	}

	locElem := e.FindElement("./location")
	if locElem == nil {
		locElem = etree.NewElement("location")
		e.InsertChild(e.FindElement("./exchangeDirection"), locElem)
	}
	locElem.SetText(unmapping.Location)

	m.used[unmapping.NewID] = true
}
