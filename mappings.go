package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
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

func (e *FlowMapEntry) key() string {
	return getKey(e.Location, e.OldID)
}

// FlowMap contains the flow mappings.
type FlowMap struct {
	mappings map[string]*FlowMapEntry
}

func readMappings() *FlowMap {
	log.Println("Read flow mappings from flow_mappings.csv")
	file, err := os.Open("flow_mappings.csv")
	if err != nil {
		log.Fatalln("Failed to read mapping file flow_mappings.csv", err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		log.Fatalln("Failed to read mapping file flow_mappings.csv", err)
	}
	fm := FlowMap{mappings: make(map[string]*FlowMapEntry)}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		e := FlowMapEntry{OldID: row[0], Location: row[1], NewID: row[2]}
		fm.mappings[e.key()] = &e
	}
	return &fm
}
