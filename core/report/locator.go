package report

// Locator pinpoints where a piece of evidence came from. Exactly one group of
// fields (file, command, passage, or note) should be set; Shape reports which.
type Locator struct {
	File      string `json:"file,omitempty"`
	LineStart int    `json:"line_start,omitempty"`
	LineEnd   int    `json:"line_end,omitempty"`

	Command    string `json:"command,omitempty"`
	Cwd        string `json:"cwd,omitempty"`
	ExitStatus *int   `json:"exit_status,omitempty"`
	LogRef     string `json:"log_ref,omitempty"`

	Passage             string `json:"passage,omitempty"`
	SourceRef           string `json:"source_ref,omitempty"`
	SourceDateOrVersion string `json:"source_date_or_version,omitempty"`
	AccessDate          string `json:"access_date,omitempty"`

	Note string `json:"note,omitempty"` // non-qualifying pointer, e.g. "PR description, paragraph 2"
}

// LocatorShape names the one complete field group a Locator populates.
type LocatorShape string

const (
	ShapeFile    LocatorShape = "file"
	ShapeCommand LocatorShape = "command"
	ShapePassage LocatorShape = "passage"
	ShapeNote    LocatorShape = "note"
	ShapeNone    LocatorShape = ""
)

// shapeGroup describes one field group's participation in a Locator: whether
// any of its fields are set, and whether the full group is populated.
type shapeGroup struct {
	shape    LocatorShape
	isSet    bool
	complete bool
}

// fileGroup reports whether any file-group field is set, and whether the group is complete.
func (l Locator) fileGroup() shapeGroup {
	isSet := l.File != "" || l.LineStart != 0 || l.LineEnd != 0
	complete := l.File != "" && l.LineStart >= 1 && l.LineEnd >= l.LineStart
	return shapeGroup{ShapeFile, isSet, complete}
}

// commandGroup reports whether any command-group field is set, and whether the group is complete.
func (l Locator) commandGroup() shapeGroup {
	isSet := l.Command != "" || l.Cwd != "" || l.LogRef != "" || l.ExitStatus != nil
	complete := l.Command != "" && l.Cwd != "" && l.LogRef != "" && l.ExitStatus != nil
	return shapeGroup{ShapeCommand, isSet, complete}
}

// passageGroup reports whether any passage-group field is set, and whether the group is complete.
func (l Locator) passageGroup() shapeGroup {
	isSet := l.Passage != "" || l.SourceRef != "" || l.SourceDateOrVersion != "" || l.AccessDate != ""
	complete := l.Passage != "" && l.SourceRef != "" && l.SourceDateOrVersion != "" && l.AccessDate != ""
	return shapeGroup{ShapePassage, isSet, complete}
}

// noteGroup reports whether the note-group field is set; the group is complete whenever it is set.
func (l Locator) noteGroup() shapeGroup {
	isSet := l.Note != ""
	return shapeGroup{ShapeNote, isSet, isSet}
}

// Shape returns the one complete shape; partial or mixed fields yield ShapeNone.
func (l Locator) Shape() LocatorShape {
	groups := []shapeGroup{l.fileGroup(), l.commandGroup(), l.passageGroup(), l.noteGroup()}

	var found shapeGroup
	setCount := 0
	for _, g := range groups {
		if g.isSet {
			setCount++
			found = g
		}
	}
	if setCount != 1 || !found.complete {
		return ShapeNone
	}
	return found.shape
}

// Qualifies is true for file, command, and passage: what the evaluator opened or ran.
func (l Locator) Qualifies() bool {
	switch l.Shape() {
	case ShapeFile, ShapeCommand, ShapePassage:
		return true
	default:
		return false
	}
}

// Evidence is a single piece of support cited for a check's verdict.
type Evidence struct {
	Kind          Kind               `json:"kind"`
	Locator       Locator            `json:"locator"`
	Observation   string             `json:"observation"`
	Provenance    EvidenceProvenance `json:"provenance"`
	Isolation     Isolation          `json:"isolation"`
	Origin        Origin             `json:"origin"`
	RubricVersion string             `json:"rubric_version,omitempty"` // required when Kind == judgment
	Model         string             `json:"model,omitempty"`          // "unknown" is a valid value
}

// EvidenceProvenance records what produced a piece of evidence and when.
type EvidenceProvenance struct {
	Tool      string `json:"tool"`
	Version   string `json:"version"`
	Revision  string `json:"revision"`
	Timestamp string `json:"timestamp"`
}

// SupportsPass is true only for observed evidence with a qualifying locator.
func (e Evidence) SupportsPass() bool { return e.Origin == OriginObserved && e.Locator.Qualifies() }
