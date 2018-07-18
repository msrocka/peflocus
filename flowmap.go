package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/msrocka/ilcd"
)

// MapKey creates a key <location-code>/<flow uuid>. The location code is
// optional and thus can be an empty string.
func MapKey(location, uuid string) string {
	key := strings.TrimSpace(location) + "/" + strings.TrimSpace(uuid)
	return strings.ToLower(key)
}

// FlowMapEntry represents a row in the mapping file
type FlowMapEntry struct {
	Location string
	OldID    string
	NewID    string
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

	// Contains the IDs of the flows that where used but not (un)mapped. These
	// flows should be copied into the target archive.
	untouchedUsed map[string]bool
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
		mappings:      make(map[string]*FlowMapEntry),
		unmappings:    make(map[string]*FlowMapEntry),
		used:          make(map[string]bool),
		untouchedUsed: make(map[string]bool)}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			log.Println("WARNING: invalid flow mapping in row", i)
			continue
		}
		e := FlowMapEntry{
			OldID:    strings.TrimSpace(row[0]),
			Location: strings.TrimSpace(row[1]),
			NewID:    strings.TrimSpace(row[2])}
		key := MapKey(e.Location, e.OldID)
		fm.mappings[key] = &e
		fm.unmappings[e.NewID] = &e
	}
	log.Println(" ... read", len(fm.mappings), "mappings")
	return &fm
}

// ResetStats clears the mapping statistics
func (m *FlowMap) ResetStats() {
	m.used = make(map[string]bool)
	m.untouchedUsed = make(map[string]bool)
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
		fn(factor)
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
	flowRef := e.FindElement("./referenceToFlowDataSet")
	if flowRef == nil {
		return
	}
	idAttr := flowRef.SelectAttr("refObjectId")
	if idAttr == nil {
		return
	}
	location := ""
	locElem := e.FindElement("./location")
	if locElem != nil {
		location = strings.TrimSpace(locElem.Text())
	}
	key := MapKey(location, idAttr.Value)
	mapping := m.mappings[key]
	if mapping == nil {
		m.untouchedUsed[idAttr.Value] = true
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
		m.untouchedUsed[idAttr.Value] = true
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

	nameElem := e.FindElement("./referenceToFlowDataSet/shortDescription")
	if nameElem != nil {
		name := nameElem.Text()
		name = strings.TrimSuffix(name, " - "+unmapping.Location)
		nameElem.SetText(name)
	}

	m.used[unmapping.NewID] = true
}
