package proof

// Package frameproof provides a simple knowledge base system using frames
// and relations (slots) to prove theorems via unification and backtracking for queries consisting of a chain of triples.
// Each triple is of the form [subject, relation, object], where subjects/objects can be frame names or variables starting with '?'.
// Relations are assumed to be constant strings.

import (
	"errors"
	"strings"

	"github.com/wanderer69/frm/internal/frame"
	"github.com/wanderer69/frm/internal/list"
	"github.com/wanderer69/frm/internal/value"
	valueType "github.com/wanderer69/frm/internal/value_types"
)

// KnowledgeBase is a collection of frames.
type KnowledgeBase struct {
	list    *list.List
	nameKey string
}

// findFrame finds a frame by its name in the knowledge base.
// Returns the frame and true if found, nil and false otherwise.
func (kb KnowledgeBase) findFrame(name string) (*frame.Frame, bool) {
	var f *frame.Frame
	kb.list.Find(func(valueItem value.Value) (bool, bool) {
		if valueItem.Type() != value.ValueTypeFrame {
			return false, false
		}
		vv, ok := valueItem.Value().(*frame.Frame)
		if !ok {
			return false, false
		}
		if len(vv.Name) > 0 {
			if vv.Name == name {
				f = vv
				return false, true
			}
		} else {
			ll := vv.Slots.Get(kb.nameKey)
			nn := ll.Get(0).Value().(string)
			if nn == name {
				f = vv
				return false, true
			}
		}
		return false, false
	})
	if f != nil {
		return f, true
	}
	return nil, false
}

// isVariable checks if a string represents a variable (starts with '?').
func isVariable(s value.Value) (string, bool) {
	if s.Type() != value.ValueTypeString {
		return "", false
	}
	v, ok := s.Value().(string)
	if !ok {
		return "", false
	}
	return v, strings.HasPrefix(v, "?")
}

func getName(s value.Value) (string, bool) {
	if s.Type() != value.ValueTypeString {
		return "", false
	}
	v, ok := s.Value().(string)
	if !ok {
		return "", false
	}
	return v, true
}

func copyBindings(b map[string]value.Value) map[string]value.Value {
	newB := make(map[string]value.Value)
	for k, v := range b {
		newB[k] = v
	}
	return newB
}

func resolve(term value.Value, b map[string]value.Value) (value.Value, bool) {
	name, isVariable := isVariable(term)
	if !isVariable {
		return term, true
	}
	val, ok := b[name]
	return val, ok
}

// proveFrom is a recursive helper for Prove, handling backtracking over the chain of triples and collecting all solutions.
func (kb KnowledgeBase) proveFrom(pos int, queryTemplates []*valueType.ValueFrame, bindings map[string]value.Value) []map[string]value.Value {
	var results []map[string]value.Value
	if pos >= len(queryTemplates) {
		return []map[string]value.Value{copyBindings(bindings)}
	}

	subjTermLst := queryTemplates[pos].Frame.Slots.Get("subj")
	subjTerm := subjTermLst.Get(0)
	relTermLst := queryTemplates[pos].Frame.Slots.Get("rel")
	relTerm := relTermLst.Get(0)
	objTermLst := queryTemplates[pos].Frame.Slots.Get("obj")
	objTerm := objTermLst.Get(0)

	// Relation must be a constant
	_, isVar := isVariable(relTerm)
	if isVar {
		return results
	}

	// Determine candidates for subject
	var subjCandidates []value.Value
	subjResolved, subjBound := resolve(subjTerm, bindings)
	if subjBound {
		subjCandidates = []value.Value{subjResolved}
	} else {
		// Free variable: try all frame names
		kb.list.Find(func(valueItem value.Value) (bool, bool) {
			subjCandidates = append(subjCandidates, valueItem)
			return false, false
		})

	}

	for _, subjName := range subjCandidates {
		newBindings := copyBindings(bindings)
		name, isVar := isVariable(subjTerm)
		if isVar && !subjBound {
			newBindings[name] = subjName
		}

		name, ok := subjName.Value().(string)
		if !ok {
			continue
		}
		// Find the frame
		f, found := kb.findFrame(name)
		if !found {
			continue
		}

		name, ok = getName(relTerm)
		if ok {
			// Get the slot values (multiple possible)
			values := f.Slots.Get(name)
			if values.Size() == 0 {
				continue
			}

			values.Find(func(slotValue value.Value) (bool, bool) {
				var subResults []map[string]value.Value
				tempBindings := copyBindings(newBindings)
				objResolved, objBound := resolve(objTerm, tempBindings)
				var val string
				var ok bool
				valStr, err := slotValue.To(value.ValueTypeString)
				if err != nil {
					goto loop1
				}
				if objBound {
					_, isEq := objResolved.Op("cmp", valStr)
					if !isEq {
						//continue
						goto loop1
					}
				} else {
					val, ok = objTerm.Value().(string)
					if !ok {
						goto loop1
					}
					tempBindings[val] = valStr
				}

				// Recurse to the next triple
				subResults = kb.proveFrom(pos+1, queryTemplates, tempBindings)
				results = append(results, subResults...)
			loop1:
				return false, false
			})
		}
	}

	return results
}

// Prove checks if the given query (chain of triples) can be satisfied via unification and backtracking in the knowledge base.
// Each triple is [subject, relation, object].
// Returns all possible bindings that make the entire chain true (empty slice if none).
// query []Triple,
func (kb KnowledgeBase) Prove(queryTemplates []*valueType.ValueFrame) []map[string]value.Value {
	if len(queryTemplates) == 0 {
		return []map[string]value.Value{} // Invalid triple
	}
	return kb.proveFrom(0, queryTemplates, make(map[string]value.Value))
}

// AddFrame adds a new frame to the knowledge base.
// Returns an error if a frame with the same name already exists.
func (kb *KnowledgeBase) AddFrame(f *frame.Frame) error {
	_, found := kb.findFrame(f.Name)
	if found {
		return errors.New("frame with this name already exists")
	}
	kb.list.Add(&valueType.ValueFrame{Frame: f})
	return nil
}

func (kb *KnowledgeBase) AddFrames(fs []*frame.Frame) error {
	for i := range fs {
		f := fs[i]
		_, found := kb.findFrame(f.Name)
		if found {
			return errors.New("frame with this name already exists")
		}
		kb.list.Add(&valueType.ValueFrame{Frame: f})
	}
	return nil
}

func NewKnowledgeBase(fs []*frame.Frame) *KnowledgeBase {
	kb := &KnowledgeBase{
		list:    &list.List{},
		nameKey: "name",
	}
	err := kb.AddFrames(fs)
	if err != nil {
		panic(err)
	}
	return kb
}

func NewKnowledgeBaseByList(fs *list.List) *KnowledgeBase {
	kb := &KnowledgeBase{
		list:    fs,
		nameKey: "наименование",
	}
	return kb
}
