package main

import (
	"log"
	"path/filepath"

	"github.com/msrocka/ilcd"
)

// Calculator calculates the LCIA results of a set of ILCD packages.
type Calculator struct {
	workdir string
}

// NewCalculator initializes a new calculator from the given args
func NewCalculator(args *Args) *Calculator {
	return &Calculator{workdir: args.WorkDir}
}

// Run executes the calculation.
func (c *Calculator) Run() {
	zips := GetZipNames(c.workdir)
	for _, name := range zips {
		log.Println("INFO: Calculate the results of", name)
		path := filepath.Join(c.workdir, name)
		reader, err := ilcd.NewZipReader(path)
		if err != nil {
			log.Println("ERROR: Failed to read zip", name, err)
			continue
		}
		data := CalcData{}
		data.impacts = c.readImpacts(reader)
		if len(data.impacts) == 0 {
			log.Println("INFO: No LCIA methods found in package",
				name, "nothing to caluclate")
			continue
		}
		data.flows = c.readFlows(reader)
		data.processes = c.readProcesses(reader, &data)
		if len(data.processes) == 0 {
			log.Println("INFO: No processes found in package",
				name, "nothing to caluclate")
			continue
		}
	}
}

func (c *Calculator) readImpacts(r *ilcd.ZipReader) []*CalcImpact {
	var impacts []*CalcImpact
	err := r.EachMethod(func(method *ilcd.Method) bool {
		impact := NewCalcImpact(method)
		if impact != nil {
			impacts = append(impacts, impact)
		}
		return true
	})
	if err != nil {
		log.Println("ERROR: failed to read some LCIA methods", err)
	}
	return impacts
}

func (c *Calculator) readProcesses(r *ilcd.ZipReader,
	data *CalcData) []*CalcProcess {
	var processes []*CalcProcess
	err := r.EachProcess(func(p *ilcd.Process) bool {
		proc := NewCalcProcess(p, data)
		if proc != nil {
			processes = append(processes, proc)
		}
		return true
	})
	if err != nil {
		log.Println("ERROR: failed to read some processes", err)
	}
	return processes
}

func (c *Calculator) readFlows(r *ilcd.ZipReader) map[string]*CalcFlow {
	m := make(map[string]*CalcFlow)
	err := r.EachFlow(func(flow *ilcd.Flow) bool {
		f := &CalcFlow{Type: flow.FlowType(), UUID: flow.UUID()}
		if flow.Info != nil && flow.Info.Name != nil {
			f.Name = flow.Info.Name.BaseName.Get("en")
		}
		m[f.UUID] = f
		return true
	})
	if err != nil {
		log.Println("ERROR: failed to read some flows", err)
	}
	return m
}
