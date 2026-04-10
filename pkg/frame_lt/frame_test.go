package framelt

import (
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	//frame "github.com/wanderer69/frm/pkg/frame_lt"
)

func TestHashTableBasic(t *testing.T) {
	ht := NewHashTable()
	ht.Put("slot1", "value1")
	ht.Put("slot1", "value2")
	ht.Put("slot2", "value3")

	values1, ok := ht.Get("slot1")
	require.True(t, ok)
	require.Equal(t, 6, len(values1))
	require.Equal(t, "value2", values1)
	//require.Equal(t, "value2", values1)
	/*
		if values1.Size() != 2 || values1.Get(0).Value().(string) != "value1" || values1.Get(1).Value().(string) != "value2" {
			t.Errorf("Expected [value1 value2], got %v", values1)
		}
	*/
	values2, ok := ht.Get("slot2")
	require.True(t, ok)
	if len(values2) != 6 || values2 != "value3" {
		t.Errorf("Expected [value3], got %v", values2)
	}
}

func TestHashTableResizeOnLoad(t *testing.T) {
	ht := NewHashTable()
	initialCapacity := ht.Capacity()

	for i := 0; i < 20; i++ {
		ht.Put(string(rune('a'+i)), fmt.Sprintf("%d", i))
	}

	if ht.Capacity() == initialCapacity {
		t.Errorf("Expected resize, capacity still %d", initialCapacity)
	}
}

func TestHashTableCollisions(t *testing.T) {
	ht := NewHashTable()

	sl := GetCollidingStrings()
	for j := range sl {
		for i := range sl[j] {
			ht.Put(sl[j][i], fmt.Sprintf("%d", i))
		}
	}

	maxLen := ht.MaxChainLength()
	t.Logf("Max chain length after 10 inserts: %d", maxLen)

	// Insert more to trigger resize on chain length.
	for i := 10; i < 20; i++ {
		key := "key" + string(rune('0'+(i%10))) // Reuse prefixes to increase chance.
		ht.Put(key, fmt.Sprintf("%d", i))
	}

	maxChainLength := ht.MaxChainLength()
	maxChainLen := ht.MaxChainLen()
	fmt.Printf("%d %d\r\n", maxChainLength, maxChainLen)
	if maxChainLength > maxChainLen {
		t.Errorf("Max chain length %d exceeds threshold %d after inserts", ht.MaxChainLength(), ht.MaxChainLen())
	} else {
		t.Logf("Max chain length after more inserts: %d", ht.MaxChainLength())
	}
	require.True(t, true)
}

func TestHashTableMemoryUsage(t *testing.T) {
	ht := NewHashTable()
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
		ht.Put(nameSlot, fmt.Sprintf("%d", i))
	}

	runtime.ReadMemStats(&m)
	finalAlloc := m.Alloc
	t.Logf("Approximate memory used: %d bytes (runtime), %d bytes (estimated), key %d, val %d", finalAlloc-initialAlloc, ht.ApproximateSize(), strLen, valueLen)
	require.True(t, true)
}

func TestFrame(t *testing.T) {
	f := NewFrame("testFrame")
	f.slots.Put("slot1", "val1")
	f.slots.Put("slot1", "val2")

	values, ok := f.slots.Get("slot1")
	require.True(t, ok)
	if len(values) != 4 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
}

func TestFrameII(t *testing.T) {
	// фрейм(наименование."query1_rel1", subj."фрейм1", rel."parent", obj."?X")
	f := NewFrame("")
	f.slots.Put("subj", "фрейм1")
	f.slots.Put("rel", "parent")
	f.slots.Put("наименование", "query1_rel1")
	f.slots.Put("obj", "?X")

	values, ok := f.slots.Get("наименование")
	require.True(t, ok)
	if len(values) != 11 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
}

func TestFrameIII(t *testing.T) {
	f := NewFrame("token_1")
	f.slots.Put("id", "1")
	f.slots.Put("lemma", "t.Lemma")
	f.slots.Put("pos", "t.UPos")
	f.slots.Put("dep_rel", "t.DepRel")
	f.slots.Put("form", "t.Form")
	f.slots.Put("feats", "t.Feats")
	f.slots.Put("head_id", "2")
	fmt.Printf("%v\r\n", f.String())
	headIDRaw, ok := f.slots.Get("head_id") //.Get(0).To(value.ValueTypeInt)
	require.True(t, ok)
	//	require.NoError(t, err)
	require.NotNil(t, headIDRaw)
}

// GetCollidingStrings returns a list of strings that have the same FNV-1a 32-bit hash value.
func GetCollidingStrings() [][]string {
	return [][]string{
		{"oQBrwu", "07bPuv"},
		{"elwTNP", "qPghVt"},
	}
}

func TestFrameIV(t *testing.T) {
	fdb := NewFrameDB()

	for j := 0; j < 100; j++ {
		f, err := fdb.NewFrame("frame" + fmt.Sprintf("%d", j))
		require.NoError(t, err)
		for i := 0; i < 100; i++ {
			nameSlot := "slot" + string(rune('0'+i))
			ef, err := fdb.AddSlot(f, nameSlot, fmt.Sprintf("%d_%d", j, i))
			require.NoError(t, err)
			require.Equal(t, f, ef)
		}
	}
	for j := 0; j < 100; j++ {
		f, err := fdb.NewFrame("frame_" + fmt.Sprintf("%d", j))
		require.NoError(t, err)
		for i := 0; i < 100; i++ {
			nameSlot := "slot_" + string(rune('0'+i))
			ef, err := fdb.AddSlotByName(f.name, nameSlot, fmt.Sprintf("%d_%d", j, i))
			require.NoError(t, err)
			require.Equal(t, f, ef)
		}
	}
	for j := 0; j < 100; j++ {
		f := fdb.GetFrameByName("frame_" + fmt.Sprintf("%d", j))
		require.NotNil(t, f)
		f = fdb.GetFrameByName("frame" + fmt.Sprintf("%d", j))
		require.NotNil(t, f)
		for i := 0; i < 100; i++ {
			nameSlot := "slot_" + string(rune('0'+i))
			efs := fdb.GetFrameBySlotByName(nameSlot)
			require.NotNil(t, efs)
			nameSlot = "slot_" + string(rune('0'+i))
			efs = fdb.GetFrameBySlotByName(nameSlot)
			require.NotNil(t, efs)
		}
	}
}
