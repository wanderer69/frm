package proof

import (
	"reflect"
	"testing"

	"github.com/wanderer69/frm/pkg/frame"
	"github.com/wanderer69/frm/pkg/value"
	valueType "github.com/wanderer69/frm/pkg/value_types"
)

func equalBindingsSlices(a, b []map[string]value.Value) bool {
	if len(a) != len(b) {
		return false
	}
	used := make([]bool, len(b))
	for _, m1 := range a {
		found := false
		for j, m2 := range b {
			if !used[j] && reflect.DeepEqual(m1, m2) {
				used[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestProve(t *testing.T) {
	fA := frame.NewFrame("A")
	fB := frame.NewFrame("B")
	fC := frame.NewFrame("C")
	fD := frame.NewFrame("D")
	fA.Slots.Put("parent", &valueType.ValueString{String: "B"})
	fB.Slots.Put("child", &valueType.ValueString{String: "C"})

	fA.Slots.Put("rel1", &valueType.ValueString{String: "B"})
	fB.Slots.Put("rel2", &valueType.ValueString{String: "C"})
	fC.Slots.Put("rel3", &valueType.ValueString{String: "D"})

	fQ1R1 := frame.NewFrame("query1_rel1")
	fQ1R1.Slots.Put("subj", &valueType.ValueString{String: "A"})
	fQ1R1.Slots.Put("rel", &valueType.ValueString{String: "parent"})
	fQ1R1.Slots.Put("obj", &valueType.ValueString{String: "?X"})

	fQ1R2 := frame.NewFrame("query1_rel2")
	fQ1R2.Slots.Put("subj", &valueType.ValueString{String: "?X"})
	fQ1R2.Slots.Put("rel", &valueType.ValueString{String: "child"})
	fQ1R2.Slots.Put("obj", &valueType.ValueString{String: "C"})

	fQ2R1 := frame.NewFrame("query2_rel1")
	fQ2R1.Slots.Put("subj", &valueType.ValueString{String: "A"})
	fQ2R1.Slots.Put("rel", &valueType.ValueString{String: "rel1"})
	fQ2R1.Slots.Put("obj", &valueType.ValueString{String: "?X"})

	fQ2R2 := frame.NewFrame("query2_rel2")
	fQ2R2.Slots.Put("subj", &valueType.ValueString{String: "?X"})
	fQ2R2.Slots.Put("rel", &valueType.ValueString{String: "rel2"})
	fQ2R2.Slots.Put("obj", &valueType.ValueString{String: "?Y"})

	fQ2R3 := frame.NewFrame("query2_rel3")
	fQ2R3.Slots.Put("subj", &valueType.ValueString{String: "?Y"})
	fQ2R3.Slots.Put("rel", &valueType.ValueString{String: "rel3"})
	fQ2R3.Slots.Put("obj", &valueType.ValueString{String: "D"})

	tests := []struct {
		name string
		kb   *KnowledgeBase
		//		query            []Triple
		queryTemplate    []*valueType.ValueFrame
		expectedBindings []map[string]value.Value
	}{
		{
			name: "Simple chain exists (2 relations)",
			kb: NewKnowledgeBase([]*frame.Frame{
				fA,
				fB,
				fC,
			}),
			/*
				query: []Triple{
					{
						Subj: &valueType.ValueString{String: "A"},
						Rel:  &valueType.ValueString{String: "parent"},
						Obj:  &valueType.ValueString{String: "?X"},
					},
					{
						Subj: &valueType.ValueString{String: "?X"},
						Rel:  &valueType.ValueString{String: "child"},
						Obj:  &valueType.ValueString{String: "C"},
					},
				},
			*/
			queryTemplate: []*valueType.ValueFrame{
				{Frame: fQ1R1}, {Frame: fQ1R2},
			},
			expectedBindings: []map[string]value.Value{{"?X": &valueType.ValueString{String: "B"}}},
		},
		{
			name: "Longer chain exists (3 relations)",
			kb: NewKnowledgeBase([]*frame.Frame{
				fA,
				fB,
				fC,
				fD,
			}),
			/*
				query: []Triple{
					{
						Subj: &valueType.ValueString{String: "A"},
						Rel:  &valueType.ValueString{String: "rel1"},
						Obj:  &valueType.ValueString{String: "?X"},
					},
					{
						Subj: &valueType.ValueString{String: "?X"},
						Rel:  &valueType.ValueString{String: "rel2"},
						Obj:  &valueType.ValueString{String: "?Y"},
					},
					{
						Subj: &valueType.ValueString{String: "?Y"},
						Rel:  &valueType.ValueString{String: "rel3"},
						Obj:  &valueType.ValueString{String: "D"},
					},
				},
			*/
			queryTemplate: []*valueType.ValueFrame{
				{Frame: fQ2R1}, {Frame: fQ2R2}, {Frame: fQ2R3},
			},
			expectedBindings: []map[string]value.Value{{"?X": &valueType.ValueString{String: "B"}, "?Y": &valueType.ValueString{String: "C"}}},
		},
		/*
			{
				name: "Single relation chain",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"rel1": {"B"}}},
					{Name: "B", Slots: map[string][]string{}},
				},
				query:            [][]string{{"A", "rel1", "B"}},
				expectedBindings: []map[string]string{{}},
			},
			{
				name: "Chain with repeated variable (cycle check)",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"self": {"A"}}},
				},
				query:            [][]string{{"A", "self", "?X"}, {"?X", "self", "?X"}},
				expectedBindings: []map[string]string{{"?X": "A"}},
			},
			{
				name: "Chain with repeated variable mismatch",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"rel1": {"B"}}},
					{Name: "B", Slots: map[string][]string{"rel2": {"C"}}},
				},
				query:            [][]string{{"A", "rel1", "?X"}, {"?X", "rel2", "?X"}},
				expectedBindings: []map[string]string{},
			},
			{
				name: "Chain does not exist - missing relation",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"parent": {"B"}}},
					{Name: "B", Slots: map[string][]string{}},
					{Name: "C", Slots: map[string][]string{}},
				},
				query:            [][]string{{"A", "parent", "?X"}, {"?X", "child", "C"}},
				expectedBindings: []map[string]string{},
			},
			{
				name: "Chain does not exist - mismatch value",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"parent": {"B"}}},
					{Name: "B", Slots: map[string][]string{"child": {"D"}}},
					{Name: "C", Slots: map[string][]string{}},
				},
				query:            [][]string{{"A", "parent", "?X"}, {"?X", "child", "C"}},
				expectedBindings: []map[string]string{},
			},
			{
				name: "Missing starting frame",
				kb: KnowledgeBase{
					{Name: "B", Slots: map[string][]string{"child": {"C"}}},
					{Name: "C", Slots: map[string][]string{}},
				},
				query:            [][]string{{"A", "parent", "?X"}, {"?X", "child", "C"}},
				expectedBindings: []map[string]string{},
			},
			{
				name: "Missing intermediate frame",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"parent": {"B"}}},
					{Name: "C", Slots: map[string][]string{}},
				},
				query:            [][]string{{"A", "parent", "?X"}, {"?X", "child", "C"}},
				expectedBindings: []map[string]string{},
			},
			{
				name:             "Empty knowledge base",
				kb:               KnowledgeBase{},
				query:            [][]string{{"A", "parent", "?X"}, {"?X", "child", "C"}},
				expectedBindings: []map[string]string{},
			},
			{
				name:             "Invalid triple length",
				kb:               KnowledgeBase{},
				query:            [][]string{{"A", "parent"}},
				expectedBindings: []map[string]string{},
			},
			{
				name: "Starting with variable",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"parent": {"B"}}},
					{Name: "B", Slots: map[string][]string{"child": {"C"}}},
				},
				query:            [][]string{{"?X", "parent", "B"}, {"B", "child", "C"}},
				expectedBindings: []map[string]string{{"?X": "A"}},
			},
			{
				name: "Variable in object position",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"parent": {"B"}}},
					{Name: "B", Slots: map[string][]string{}},
				},
				query:            [][]string{{"A", "parent", "?Y"}},
				expectedBindings: []map[string]string{{"?Y": "B"}},
			},
			{
				name: "Starting variable no match",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"parent": {"B"}}},
					{Name: "B", Slots: map[string][]string{}},
				},
				query:            [][]string{{"?X", "child", "C"}},
				expectedBindings: []map[string]string{},
			},
			{
				name: "Repeated variable consistency",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"rel": {"B"}}},
					{Name: "B", Slots: map[string][]string{"rel": {"A"}}},
				},
				query: [][]string{{"?X", "rel", "?Y"}, {"?Y", "rel", "?X"}},
				expectedBindings: []map[string]string{
					{"?X": "A", "?Y": "B"},
					{"?X": "B", "?Y": "A"},
				},
			},
			{
				name: "Multiple values, one matches",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"parent": {"B", "D"}}},
					{Name: "B", Slots: map[string][]string{"child": {"C"}}},
					{Name: "D", Slots: map[string][]string{"child": {"E"}}},
					{Name: "C", Slots: map[string][]string{}},
					{Name: "E", Slots: map[string][]string{}},
				},
				query:            [][]string{{"A", "parent", "?X"}, {"?X", "child", "C"}},
				expectedBindings: []map[string]string{{"?X": "B"}},
			},
			{
				name: "Multiple values, none match",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"parent": {"B", "D"}}},
					{Name: "B", Slots: map[string][]string{"child": {"F"}}},
					{Name: "D", Slots: map[string][]string{"child": {"E"}}},
					{Name: "C", Slots: map[string][]string{}},
					{Name: "E", Slots: map[string][]string{}},
					{Name: "F", Slots: map[string][]string{}},
				},
				query:            [][]string{{"A", "parent", "?X"}, {"?X", "child", "C"}},
				expectedBindings: []map[string]string{},
			},
			{
				name: "Multiple paths through different branches",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"rel1": {"B", "D"}}},
					{Name: "B", Slots: map[string][]string{"rel2": {"C"}}},
					{Name: "D", Slots: map[string][]string{"rel2": {"C"}}},
					{Name: "C", Slots: map[string][]string{}},
				},
				query: [][]string{{"A", "rel1", "?X"}, {"?X", "rel2", "C"}},
				expectedBindings: []map[string]string{
					{"?X": "B"},
					{"?X": "D"},
				},
			},
			{
				name: "Cycle with three nodes",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"rel": {"B"}}},
					{Name: "B", Slots: map[string][]string{"rel": {"C"}}},
					{Name: "C", Slots: map[string][]string{"rel": {"A"}}},
				},
				query: [][]string{{"?X", "rel", "?Y"}, {"?Y", "rel", "?Z"}, {"?Z", "rel", "?X"}},
				expectedBindings: []map[string]string{
					{"?X": "A", "?Y": "B", "?Z": "C"},
					{"?X": "B", "?Y": "C", "?Z": "A"},
					{"?X": "C", "?Y": "A", "?Z": "B"},
				},
			},
			{
				name: "Cycle with self-loop and others",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"loop": {"A"}}},
					{Name: "B", Slots: map[string][]string{"loop": {"B"}}},
				},
				query: [][]string{{"?X", "loop", "?X"}},
				expectedBindings: []map[string]string{
					{"?X": "A"},
					{"?X": "B"},
				},
			},
			{
				name: "No cycle match",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"rel": {"B"}}},
					{Name: "B", Slots: map[string][]string{"rel": {"C"}}},
					{Name: "C", Slots: map[string][]string{"rel": {"D"}}},
				},
				query:            [][]string{{"?X", "rel", "?Y"}, {"?Y", "rel", "?Z"}, {"?Z", "rel", "?X"}},
				expectedBindings: []map[string]string{},
			},
			{
				name: "Cycle with three nodes II",
				kb: KnowledgeBase{
					{Name: "A", Slots: map[string][]string{"rel": {"B", "D"}}},
					{Name: "B", Slots: map[string][]string{"rel": {"C"}}},
					{Name: "D", Slots: map[string][]string{"rel": {"B"}}},
					{Name: "C", Slots: map[string][]string{"rel": {"A"}}},
					{Name: "E", Slots: map[string][]string{"rel": {"F"}}},
					{Name: "F", Slots: map[string][]string{"rel": {"A"}}},
				},
				query: [][]string{{"?X", "rel", "?Y"}, {"?Y", "rel", "?Z"}, {"?Z", "rel", "?X"}},
				expectedBindings: []map[string]string{
					{"?X": "A", "?Y": "B", "?Z": "C"},
					{"?X": "B", "?Y": "C", "?Z": "A"},
					{"?X": "C", "?Y": "A", "?Z": "B"},
				},
			},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.kb.Prove(tt.queryTemplate) // tt.query,
			if !equalBindingsSlices(got, tt.expectedBindings) {
				t.Errorf("Prove() = %v, want %v", got, tt.expectedBindings)
			}
		})
	}
}

/*
func TestAddFrame(t *testing.T) {
	kb := KnowledgeBase{}

	err := kb.AddFrame(Frame{Name: "A", Slots: map[string][]string{}})
	if err != nil {
		t.Errorf("AddFrame() error = %v, want nil", err)
	}

	err = kb.AddFrame(Frame{Name: "A", Slots: map[string][]string{}})
	if err == nil {
		t.Errorf("AddFrame() duplicate error = nil, want error")
	}

	if len(kb) != 1 {
		t.Errorf("KnowledgeBase length = %d, want 1", len(kb))
	}
}
*/
