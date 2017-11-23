package main

import (
	"log"
	"path/filepath"

	"github.com/msrocka/ilcd"
)

// CalcImpact contains the information of an LCIA category for the calculation.
type CalcImpact struct {
	UUID    string
	Name    string
	Method  string
	Factors map[string]float64
}

// CalcProcess contains the process information that are used for the
// calculation.
type CalcProcess struct {
	UUID     string
	Name     string
	Location string
	RefFlow  *CalcExchange
	Results  map[string]float64
	Links    []CalcExchange
}

// CalcFlow contains the information of a flow for the calculation.
type CalcFlow struct {
	UUID string
	Name string
	Type ilcd.FlowType
}

// CalcExchange contains information of a process input or output.
type CalcExchange struct {
	FlowID string
	Amount float64
}

// NewCalcProcess initializes a process with a direct LCIA result.
func NewCalcProcess(p *ilcd.Process, impacts []*CalcImpact) *CalcProcess {
	if p == nil || p.Info == nil {
		return nil
	}
	proc := &CalcProcess{
		UUID:    p.UUID(),
		Name:    p.FullName("en"),
		Results: make(map[string]float64)}
	if p.Location != nil {
		proc.Location = p.Location.Code
	}
	proc.addDirectResults(p, impacts)
	return proc
}

func (proc *CalcProcess) addDirectResults(p *ilcd.Process, impacts []*CalcImpact) {
	log.Println("INFO: Calculate direct result for process", proc.UUID)
	for _, impact := range impacts {
		result := 0.0
		for _, e := range p.Exchanges {
			if e.Flow == nil {
				continue
			}
			value := e.ResultingAmount
			if e.Direction == "Input" {
				value = -value
			}
			flowID := NormKey(e.Flow.UUID)
			location := NormKey(e.Location)
			if location == "" {
				factor := impact.Factors[flowID]
				if factor != 0.0 {
					result += (factor * value)
				}
			} else {
				factor, found := impact.Factors[location+"/"+flowID]
				if !found {
					factor, found = impact.Factors[flowID]
					if found {
						log.Println(" ... WARNING: use generic LCIA factor for",
							"regionalized exchange, flow:", flowID, "@", location)
					}
				}
				if factor != 0.0 {
					result += (factor * value)
				}
			}
		}
		proc.Results[impact.UUID] = result
	}
}

// NewCalcImpact initializes a LCIA category for the calculation. It returns nil
// if something went wrong.
func NewCalcImpact(m *ilcd.Method) *CalcImpact {
	if m == nil || m.Info == nil {
		return nil
	}
	name := ""
	if m.Info.Name != nil {
		name = m.Info.Name.Get("en")
	}
	if len(m.Factors) == 0 {
		log.Println("ERROR: No LCIA factors in method", name)
		return nil
	}
	impact := &CalcImpact{
		UUID:    m.Info.UUID,
		Name:    name,
		Method:  m.Info.Methodology,
		Factors: make(map[string]float64)}
	for _, f := range m.Factors {
		impact.addFactor(&f)
	}
	log.Println("INFO: Loaded method", name, "with", len(impact.Factors), "factors")
	return impact
}

func (impact *CalcImpact) addFactor(f *ilcd.ImpactFactor) {
	if impact == nil || f == nil || f.Flow == nil || f.MeanValue == 0.0 {
		return
	}
	value := f.MeanValue
	if f.Direction == "Input" {
		value = -f.MeanValue
	}
	flowID := NormKey(f.Flow.UUID)
	location := NormKey(f.Location)
	if location == "" {
		impact.addFactorValue(flowID, value)
	} else {
		impact.addFactorValue(location+"/"+flowID, value)
	}
}

func (impact *CalcImpact) addFactorValue(key string, value float64) {
	oldVal, contains := impact.Factors[key]
	if contains {
		log.Println(" ... ERROR: multiple LCIA factors for same flow",
			key, ": ", oldVal, value)
		return
	}
	impact.Factors[key] = value
}

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
		impacts := c.impacts(reader)
		if len(impacts) == 0 {
			log.Println("INFO: No LCIA methods found in package",
				name, "nothing to caluclate")
			continue
		}
		procs := c.processes(reader, impacts)
		if len(procs) == 0 {
			log.Println("INFO: No processes found in package",
				name, "nothing to caluclate")
			continue
		}
	}
}

func (c *Calculator) impacts(r *ilcd.ZipReader) []*CalcImpact {
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

func (c *Calculator) processes(r *ilcd.ZipReader,
	impacts []*CalcImpact) []*CalcProcess {
	var processes []*CalcProcess
	err := r.EachProcess(func(p *ilcd.Process) bool {
		proc := NewCalcProcess(p, impacts)
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
