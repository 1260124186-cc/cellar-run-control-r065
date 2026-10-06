package domain

import "time"

// Occupancy records which formulas and vessels are still held by batches that
// have not reached a terminal state. It is maintained alongside every batch
// write so retirement and reuse decisions do not have to rescan every run.
type Occupancy struct {
	FormulaOpen map[string]int64 `json:"formula_open"`
	VesselOpen  map[string]int64 `json:"vessel_open"`
	Generation  int64            `json:"generation"`
}

func EmptyOccupancy() Occupancy {
	return Occupancy{
		FormulaOpen: map[string]int64{},
		VesselOpen:  map[string]int64{},
	}
}

func (o Occupancy) FormulaInUse(formulaID string) bool {
	return o.FormulaOpen[formulaID] > 0
}

func (o Occupancy) VesselInUse(vesselID string) bool {
	return o.VesselOpen[vesselID] > 0
}

func (o *Occupancy) TrackFormula(formulaID string) {
	if o.FormulaOpen == nil {
		o.FormulaOpen = map[string]int64{}
	}
	o.FormulaOpen[formulaID]++
	o.Generation++
}

func (o *Occupancy) ReleaseFormula(formulaID string) {
	if o.FormulaOpen[formulaID] > 0 {
		o.FormulaOpen[formulaID]--
	}
	o.Generation++
}

func (o *Occupancy) TrackVessel(vesselID string) {
	if o.VesselOpen == nil {
		o.VesselOpen = map[string]int64{}
	}
	o.VesselOpen[vesselID]++
	o.Generation++
}

func (o *Occupancy) ReleaseVessel(vesselID string) {
	if o.VesselOpen[vesselID] > 0 {
		o.VesselOpen[vesselID]--
	}
	o.Generation++
}

func (o Occupancy) Clone() Occupancy {
	cloned := Occupancy{Generation: o.Generation}
	if o.FormulaOpen != nil {
		cloned.FormulaOpen = make(map[string]int64, len(o.FormulaOpen))
		for key, value := range o.FormulaOpen {
			cloned.FormulaOpen[key] = value
		}
	}
	if o.VesselOpen != nil {
		cloned.VesselOpen = make(map[string]int64, len(o.VesselOpen))
		for key, value := range o.VesselOpen {
			cloned.VesselOpen[key] = value
		}
	}
	return cloned
}

type Snapshot struct {
	Revision    int64                      `json:"revision"`
	UpdatedAt   time.Time                  `json:"updated_at"`
	Occupancy   Occupancy                  `json:"occupancy"`
	Formulas    map[string]Formula         `json:"formulas"`
	Vessels     map[string]Vessel          `json:"vessels"`
	Runs        map[string]FermentationRun `json:"runs"`
	NameIndex   map[string]string          `json:"formula_name_index"`
	VesselIndex map[string]string          `json:"vessel_code_index"`
	RunIndex    map[string]string          `json:"run_code_index"`
}

func EmptySnapshot() Snapshot {
	return Snapshot{
		Revision:    0,
		Occupancy:   EmptyOccupancy(),
		Formulas:    map[string]Formula{},
		Vessels:     map[string]Vessel{},
		Runs:        map[string]FermentationRun{},
		NameIndex:   map[string]string{},
		VesselIndex: map[string]string{},
		RunIndex:    map[string]string{},
	}
}

func (s Snapshot) Clone() Snapshot {
	cloned := EmptySnapshot()
	cloned.Revision = s.Revision
	cloned.UpdatedAt = s.UpdatedAt
	cloned.Occupancy = s.Occupancy.Clone()
	for key, value := range s.Formulas {
		cloned.Formulas[key] = value.Clone()
	}
	for key, value := range s.Vessels {
		cloned.Vessels[key] = value
		if value.ActiveRunID != nil {
			runID := *value.ActiveRunID
			cloned.Vessels[key] = value
			cloned.Vessels[key] = withActiveRun(value, &runID)
		}
	}
	for key, value := range s.Runs {
		cloned.Runs[key] = value.Clone()
	}
	for key, value := range s.NameIndex {
		cloned.NameIndex[key] = value
	}
	for key, value := range s.VesselIndex {
		cloned.VesselIndex[key] = value
	}
	for key, value := range s.RunIndex {
		cloned.RunIndex[key] = value
	}
	return cloned
}

func withActiveRun(v Vessel, runID *string) Vessel {
	value := v
	if runID == nil {
		value.ActiveRunID = nil
		return value
	}
	copied := *runID
	value.ActiveRunID = &copied
	return value
}
