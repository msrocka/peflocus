package main

import (
	"log"

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
	RefFlows []CalcExchange
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

// CalcData contains all the calculation data of a package
type CalcData struct {
	impacts   []*CalcImpact
	processes []*CalcProcess
	flows     map[string]*CalcFlow
}

func (data *CalcData) isRefFlow(p *ilcd.Process, e *ilcd.Exchange) bool {
	flow := data.flows[e.Flow.UUID]
	if flow == nil || flow.Type == ilcd.ElementaryFlow {
		return false
	}
	isRef := false
	for _, ref := range p.QRefs {
		if e.InternalID == ref {
			isRef = true
			break
		}
	}
	if !isRef {
		return false
	}
	if e.Direction == "Output" && flow.Type == ilcd.ProductFlow {
		return true
	}
	if e.Direction == "Input" && flow.Type == ilcd.WasteFlow {
		return true
	}
	return false
}

func (data *CalcData) isLinkFlow(e *ilcd.Exchange) bool {
	flow := data.flows[e.Flow.UUID]
	if flow == nil || flow.Type == ilcd.ElementaryFlow {
		return false
	}
	if e.Direction == "Input" && flow.Type == ilcd.ProductFlow {
		return true
	}
	if e.Direction == "Output" && flow.Type == ilcd.WasteFlow {
		return true
	}
	return false
}

// NewCalcProcess initializes a process with a direct LCIA result.
func NewCalcProcess(p *ilcd.Process, data *CalcData) *CalcProcess {
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
	for _, e := range p.Exchanges {
		if e.Flow == nil || e.ResultingAmount == 0.0 {
			continue
		}
		if data.isRefFlow(p, &e) {
			proc.RefFlows = append(proc.RefFlows,
				CalcExchange{FlowID: e.Flow.UUID, Amount: e.ResultingAmount})
		} else if data.isLinkFlow(&e) {
			proc.Links = append(proc.Links,
				CalcExchange{FlowID: e.Flow.UUID, Amount: e.ResultingAmount})
		}
	}

	if len(proc.RefFlows) > 1 {
		log.Println("WARNING: Process", proc.UUID, "has multiple reference flows")
	}
	if len(proc.RefFlows) == 0 {
		log.Println("WARNING: Process", proc.UUID, "has no reference flows")
	}

	proc.addDirectResults(p, data.impacts)
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
