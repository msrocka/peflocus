package main

import (
	"encoding/xml"
	"log"
	"path/filepath"
	"strings"

	"github.com/msrocka/ilcd"
)

// Merger merges a set of ILCD zip packages into a single file.
type Merger struct {
	workdir  string
	skipDocs bool
	content  map[string]bool
}

// NewMerger initializes a new merger from the given args.
func NewMerger(args *Args) *Merger {
	m := &Merger{workdir: args.WorkDir, content: make(map[string]bool)}
	if args.SkipDocs == "true" || args.SkipDocs == "1" {
		m.skipDocs = true
	}
	return m
}

// Run executes the package merging
func (m *Merger) Run() {
	destPath := filepath.Join(m.workdir, "peflocus_merged.zip")
	DeleteExisting(destPath)
	writer, err := ilcd.NewZipWriter(destPath)
	if err != nil {
		log.Fatalln("ERROR: cannot write to zip file", destPath, ": ", err)
	}
	defer writer.Close()
	zips := GetZipNames(m.workdir)
	log.Println("Merge", len(zips), "zip files into", destPath)
	for _, name := range zips {
		reader, err := ilcd.NewZipReader(filepath.Join(m.workdir, name))
		if err != nil {
			log.Println("ERROR: failed to read zip", name, ": ", err)
			continue
		}
		log.Println("INFO: add zip", name)
		m.doIt(reader, writer)
		if err = reader.Close(); err != nil {
			log.Println("ERROR: failed to close zip", name, ": ", err)
		}
	}
	log.Println("INFO: merged", len(m.content), "entries into a single file")
}

func (m *Merger) doIt(reader *ilcd.ZipReader, writer *ilcd.ZipWriter) {
	reader.EachEntry(func(path string, data []byte) error {
		t := GetPathType(path)
		if t == -1 {
			log.Println("INFO: ignore", path)
			return nil
		}
		if t == ilcd.ExternalDoc {
			if m.skipDocs {
				return nil
			}
			m.addExternalDoc(writer, path, data)
			return nil
		}
		ds := m.init(t)
		err := xml.Unmarshal(data, ds)
		if err != nil || ds == nil {
			log.Println("ERROR: could not load data set for", path)
			return nil
		}
		p := "ILCD/" + t.Folder() + "/" + ds.UUID() + ".xml"
		if !m.content[p] {
			m.content[p] = true
			err := writer.WriteEntry(p, data)
			if err != nil {
				log.Println("ERROR: failed to add data set", path, err)
			} else {
				log.Println("INFO: added data set", p)
			}
		}
		return nil
	})
}

func (m *Merger) addExternalDoc(writer *ilcd.ZipWriter, path string, data []byte) {
	doc := m.docName(path)
	p := "ILCD/" + ilcd.ExternalDoc.Folder() + "/" + doc
	if doc == "" || m.content[p] {
		return
	}
	m.content[p] = true
	err := writer.WriteEntry(p, data)
	if err != nil {
		log.Println("ERROR: failed to add external doc", doc, err)
	} else {
		log.Println("INFO: added external doc", doc)
	}
}

func (m *Merger) docName(path string) string {
	parts := strings.Split(path, ilcd.ExternalDoc.Folder())
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimLeft(parts[1], "/\\")
}

func (m *Merger) init(t ilcd.DataSetType) ilcd.DataSet {
	switch t {
	case ilcd.ContactDataSet:
		return &ilcd.Contact{}
	case ilcd.SourceDataSet:
		return &ilcd.Source{}
	case ilcd.UnitGroupDataSet:
		return &ilcd.UnitGroup{}
	case ilcd.FlowPropertyDataSet:
		return &ilcd.FlowProperty{}
	case ilcd.FlowDataSet:
		return &ilcd.Flow{}
	case ilcd.ProcessDataSet:
		return &ilcd.Process{}
	case ilcd.MethodDataSet:
		return &ilcd.Method{}
	default:
		return nil
	}
}
