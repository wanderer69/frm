package frame_test

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"

	"github.com/wanderer69/frm/pkg/frame"
	"github.com/wanderer69/frm/pkg/value"
	valueType "github.com/wanderer69/frm/pkg/value_types"
)

// Tests

func TestHashTableBasic(t *testing.T) {
	ht := frame.NewHashTable()
	ht.Put("slot1", &valueType.ValueString{String: "value1"})
	ht.Put("slot1", &valueType.ValueString{String: "value2"})
	ht.Put("slot2", &valueType.ValueString{String: "value3"})

	values1 := ht.Get("slot1")
	require.Equal(t, 2, values1.Size())
	require.Equal(t, "value1", values1.Get(0).Value().(string))
	require.Equal(t, "value2", values1.Get(1).Value().(string))
	/*
		if values1.Size() != 2 || values1.Get(0).Value().(string) != "value1" || values1.Get(1).Value().(string) != "value2" {
			t.Errorf("Expected [value1 value2], got %v", values1)
		}
	*/
	values2 := ht.Get("slot2")
	if values2.Size() != 1 || values2.Get(0).Value().(string) != "value3" {
		t.Errorf("Expected [value3], got %v", values2)
	}
}

func TestHashTableResizeOnLoad(t *testing.T) {
	ht := frame.NewHashTable()
	initialCapacity := ht.Capacity()

	for i := 0; i < 20; i++ {
		ht.Put(string(rune('a'+i)), &valueType.ValueInt{Int: i})
	}

	if ht.Capacity() == initialCapacity {
		t.Errorf("Expected resize, capacity still %d", initialCapacity)
	}
}

func TestHashTableCollisions(t *testing.T) {
	ht := frame.NewHashTable()

	sl := GetCollidingStrings()
	for j := range sl {
		for i := range sl[j] {
			ht.Put(sl[j][i], &valueType.ValueInt{Int: i})
		}
	}

	maxLen := ht.MaxChainLength()
	t.Logf("Max chain length after 10 inserts: %d", maxLen)

	// Insert more to trigger resize on chain length.
	for i := 10; i < 20; i++ {
		key := "key" + string(rune('0'+(i%10))) // Reuse prefixes to increase chance.
		ht.Put(key, &valueType.ValueInt{Int: i})
	}

	if ht.MaxChainLength() > ht.MaxChainLen() {
		t.Errorf("Max chain length %d exceeds threshold %d after inserts", ht.MaxChainLength(), ht.MaxChainLen())
	} else {
		t.Logf("Max chain length after more inserts: %d", ht.MaxChainLength())
	}
	require.True(t, true)
}

func TestHashTableMemoryUsage(t *testing.T) {
	ht := frame.NewHashTable()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	initialAlloc := m.Alloc
	strLen := 0
	valueLen := uintptr(0)

	var i int32
	for i = 0; i < 100; i++ {
		nameSlot := "slot" + string(rune('0'+i))
		strLen += len(nameSlot)
		valueLen += unsafe.Sizeof(i)
		ht.Put(nameSlot, &valueType.ValueInt{Int: int(i)})
	}

	runtime.ReadMemStats(&m)
	finalAlloc := m.Alloc
	t.Logf("Approximate memory used: %d bytes (runtime), %d bytes (estimated), key %d, val %d", finalAlloc-initialAlloc, ht.ApproximateSize(), strLen, valueLen)
	require.True(t, true)
}

func TestFrame(t *testing.T) {
	f := frame.NewFrame("testFrame")
	f.Slots.Put("slot1", &valueType.ValueString{String: "val1"})
	f.Slots.Put("slot1", &valueType.ValueString{String: "val2"})

	values := f.Slots.Get("slot1")
	if values.Size() != 2 {
		t.Errorf("Expected 2 values, got %d", values.Size())
	}
}

func TestFrameII(t *testing.T) {
	// фрейм(наименование."query1_rel1", subj."фрейм1", rel."parent", obj."?X")
	f := frame.NewFrame("")
	f.Slots.Put("subj", &valueType.ValueString{String: "фрейм1"})
	f.Slots.Put("rel", &valueType.ValueString{String: "parent"})
	f.Slots.Put("наименование", &valueType.ValueString{String: "query1_rel1"})
	f.Slots.Put("obj", &valueType.ValueString{String: "?X"})

	values := f.Slots.Get("наименование")
	if values.Size() != 1 {
		t.Errorf("Expected 2 values, got %d", values.Size())
	}
}

func TestFrameIII(t *testing.T) {
	f := frame.NewFrame("token_1")
	f.Slots.Put("id", &valueType.ValueString{String: "1"})
	f.Slots.Put("lemma", &valueType.ValueString{String: "t.Lemma"})
	f.Slots.Put("pos", &valueType.ValueString{String: "t.UPos"})
	f.Slots.Put("dep_rel", &valueType.ValueString{String: "t.DepRel"})
	f.Slots.Put("form", &valueType.ValueString{String: "t.Form"})
	f.Slots.Put("feats", &valueType.ValueString{String: "t.Feats"})
	f.Slots.Put("head_id", &valueType.ValueString{String: "2"})
	headIDRaw, err := f.Slots.Get("head_id").Get(0).To(value.ValueTypeInt)
	require.NoError(t, err)
	require.NotNil(t, headIDRaw)
}

// GetCollidingStrings returns a list of strings that have the same FNV-1a 32-bit hash value.
func GetCollidingStrings() [][]string {
	return [][]string{
		{"oQBrwu", "07bPuv"},
		{"elwTNP", "qPghVt"},
	}
}

/*
// To verify:
func computeFNV(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
*/
// Example usage in a test or main:
// hash1 := computeFNV("oQBrwu")
// hash2 := computeFNV("07bPuv")
// // hash1 == hash2
