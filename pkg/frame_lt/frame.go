package framelt

import (
	"fmt"
	"hash/fnv"
	"slices"
	"sync"
	"unsafe"

	constants "github.com/wanderer69/frm/pkg/constants"
	"github.com/wanderer69/frm/pkg/value"
	valueType "github.com/wanderer69/frm/pkg/value_types"
)

// Entry represents a node in the linked list for collision resolution.
type Entry struct {
	Key string
	//	Values *list.List
	Next  *Entry
	Value string
}

func (e *Entry) Iterate() func() (*Entry, bool) {
	n := e
	return func() (*Entry, bool) {
		isFinished := false
		var p *Entry
		if n.Next != nil {
			isFinished = true
			p = n
			n = n.Next
		}
		return p, isFinished
	}
}

// HashTable is a custom hash table implementation with separate chaining.
type HashTable struct {
	buckets     []*Entry
	size        int     // Number of unique keys
	capacity    int     // Number of buckets
	loadFactor  float64 // Threshold for resizing based on load
	maxChainLen int     // Maximum allowed chain length before resizing
}

// NewHashTable creates a new hash table with initial capacity.
func NewHashTable() *HashTable {
	return &HashTable{
		buckets:     make([]*Entry, 16),
		size:        0,
		capacity:    4,
		loadFactor:  0.75,
		maxChainLen: 5, // Arbitrary threshold for max chain length
	}
}

// hash computes the hash value for a key.
func (ht *HashTable) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// resize doubles the capacity of the hash table and rehashes all entries.
func (ht *HashTable) resize() {
	newCapacity := ht.capacity * 2
	newBuckets := make([]*Entry, newCapacity)

	for _, entry := range ht.buckets {
		for e := entry; e != nil; e = e.Next {
			//fmt.Printf("resize %s\r\n", e.Key)
			index := ht.hash(e.Key) % uint32(newCapacity)
			ee := newBuckets[index]
			if ee == nil {
				newBuckets[index] = &Entry{Key: e.Key, Value: e.Value}
				continue
			}
			// New key
			newEntry := &Entry{Key: e.Key, Next: ee, Value: e.Value}
			newBuckets[index] = newEntry
		}
	}

	ht.buckets = newBuckets
	ht.capacity = newCapacity
}

func (ht *HashTable) Capacity() int {
	return ht.capacity
}

func (ht *HashTable) MaxChainLen() int {
	return ht.maxChainLen
}

// Put adds or appends a value to the list for the given key.
func (ht *HashTable) Put(key string, value string) {
	index := ht.hash(key) % uint32(ht.capacity)
	entry := ht.buckets[index]

	// Check for existing key
	for e := entry; e != nil; e = e.Next {
		//fmt.Printf("%s %s\r\n", key, e.Key)
		if e.Key == key {
			e.Value = value
			ht.checkResize()
			return
		}
	}
	//fmt.Printf("%s \r\n", key)

	// New key
	newEntry := &Entry{Key: key, Next: entry, Value: value}
	ht.buckets[index] = newEntry
	ht.size++

	// Check chain length
	chainLen := 0
	for e := ht.buckets[index]; e != nil; e = e.Next {
		chainLen++
	}
	if chainLen > ht.maxChainLen {
		ht.resize()
		return
	}

	ht.checkResize()
}

// checkResize checks if the table needs resizing based on load factor.
func (ht *HashTable) checkResize() {
	if float64(ht.size)/float64(ht.capacity) > ht.loadFactor {
		ht.resize()
	}
}

// Get retrieves the list of values for the given key.
func (ht *HashTable) Get(key string) (string, bool) {
	index := ht.hash(key) % uint32(ht.capacity)
	entry := ht.buckets[index]

	for e := entry; e != nil; e = e.Next {
		if e.Key == key {
			return e.Value, true
		}
	}
	return "", false
}

// MaxChainLength returns the length of the longest chain.
func (ht *HashTable) MaxChainLength() int {
	maxLen := 0
	for _, entry := range ht.buckets {
		len := 0
		for e := entry; e != nil; e = e.Next {
			len++
		}
		if len > maxLen {
			maxLen = len
		}
	}
	return maxLen
}

// ApproximateSize estimates the memory usage of the hash table.
func (ht *HashTable) ApproximateSize() uintptr {
	size := unsafe.Sizeof(*ht) + uintptr(ht.capacity)*unsafe.Sizeof((*Entry)(nil))
	for _, entry := range ht.buckets {
		for e := entry; e != nil; e = e.Next {
			size += unsafe.Sizeof(*e) + uintptr(len(e.Value)) + uintptr(len(e.Key))
		}
	}
	return size
}

// Frame represents the frame with a name and slots hash table.
type Frame struct {
	name  string
	slots *HashTable
}

// NewFrame creates a new frame.
func NewFrame(name string) *Frame {
	return &Frame{
		name:  name,
		slots: NewHashTable(),
	}
}

func (f *Frame) Name() string {
	return f.name
}

func (f *Frame) Rename(name string) {
	f.name = name
}

func (f *Frame) String() string {
	s := fmt.Sprintf("%s {", f.name)
	for _, entry := range f.slots.buckets {
		if entry == nil {
			continue
		}
		s = s + " ["
		for e := entry; e != nil; e = e.Next {
			s = s + e.Key + " " + e.Value //+ " "
		}
		s = s + "]"
	}
	s = s + "}"
	return s
}

func (f *Frame) Item(n int) *Entry {
	if n < 0 || n > len(f.slots.buckets) {
		return nil
	}
	return f.slots.buckets[n]
}

func (f *Frame) Len() int {
	return len(f.slots.buckets)
}

func (f *Frame) Set(key string, value value.Value) {
	f.slots.Put(key, value.Type().ToString())
}

func (f *Frame) Get(key string) value.Value {
	l, ok := f.slots.Get(key)
	if !ok {
		return constants.ValueNil
	}
	return &valueType.ValueString{String: l}
}

func (f *Frame) Iter(fn func(key string, value string)) {
	for _, entry := range f.slots.buckets {
		if entry == nil {
			continue
		}
		for e := entry; e != nil; e = e.Next {
			if fn != nil {
				fn(e.Key, e.Value)
			}
		}
	}
}

func (f *Frame) Iterator() func() (string, string, bool) {
	i := 0
	state := 0
	var entry, e *Entry
	return func() (string, string, bool) {
		var isFinished bool
		for {
			switch state {
			case 0:
				if isFinished {
					return "", "", isFinished
				}
				if len(f.slots.buckets) <= i+1 {
					isFinished = true
				}
				entry = f.slots.buckets[i]
				e = entry
				i = i + 1
				//				if !isFinished {
				if f.slots.buckets[i] == nil {
					for j := i; j < len(f.slots.buckets); j++ {
						if len(f.slots.buckets) <= j+1 {
							isFinished = true
						}
						i = j
						entry = f.slots.buckets[j]
						if entry != nil {
							break
						}
					}
				}
				//				}
				/*
					if entry == nil {
						continue
					}
				*/
				state = 1
			case 1:
				if e == nil {
					if isFinished {
						return "", "", isFinished
					}
					state = 0
					continue
				}
				key := e.Key
				value := e.Value
				e = e.Next
				return key, value, isFinished
			}
		}
	}
}

type FrameDB struct {
	frames         []*Frame
	mu             sync.Mutex
	frameByName    map[string]*Frame
	framesBySlot   map[string][]*Frame
	framePosByName map[string]int
}

func NewFrameDB() *FrameDB {
	return &FrameDB{
		frameByName:    make(map[string]*Frame),
		framesBySlot:   make(map[string][]*Frame),
		framePosByName: make(map[string]int),
		mu:             sync.Mutex{},
	}
}

func (fdb *FrameDB) NewFrame(name string) (*Frame, error) {
	fdb.mu.Lock()
	defer fdb.mu.Unlock()
	_, ok := fdb.frameByName[name]
	if ok {
		return nil, fmt.Errorf("duplicate frame name %s", name)
	}
	f := NewFrame(name)
	fdb.frameByName[name] = f
	fdb.frames = append(fdb.frames, f)
	fdb.framePosByName[name] = len(fdb.frames)

	return f, nil
}

func (fdb *FrameDB) AddSlotByName(name string, slotName string, value string) (*Frame, error) {
	fdb.mu.Lock()
	defer fdb.mu.Unlock()
	f, ok := fdb.frameByName[name]
	if !ok {
		return nil, fmt.Errorf("not found frame name %s", name)
	}
	f.slots.Put(slotName, value)

	fl, ok := fdb.framesBySlot[slotName]
	if !ok {
		fl = []*Frame{}
	}
	fl = append(fl, f)
	fdb.framesBySlot[slotName] = fl

	return f, nil
}

func (fdb *FrameDB) AddSlot(f *Frame, slotName string, value string) (*Frame, error) {
	fdb.mu.Lock()
	defer fdb.mu.Unlock()
	_, ok := fdb.frameByName[f.name]
	if !ok {
		return nil, fmt.Errorf("not found frame name %s", f.name)
	}
	f.slots.Put(slotName, value)

	fl, ok := fdb.framesBySlot[slotName]
	if !ok {
		fl = []*Frame{}
	}
	fl = append(fl, f)
	fdb.framesBySlot[slotName] = fl

	return f, nil
}

func (fdb *FrameDB) GetFrameBySlotByName(slotName string) []*Frame {
	fdb.mu.Lock()
	defer fdb.mu.Unlock()
	return fdb.framesBySlot[slotName]
}

func (fdb *FrameDB) GetFrameByName(name string) *Frame {
	fdb.mu.Lock()
	defer fdb.mu.Unlock()
	return fdb.frameByName[name]
}

func (fdb *FrameDB) DeleteFrameByName(name string) (*Frame, error) {
	fdb.mu.Lock()
	defer fdb.mu.Unlock()
	f, ok := fdb.frameByName[name]
	if ok {
		return nil, fmt.Errorf("not found frame name %s", name)
	}
	delete(fdb.frameByName, name)

	f.Iter(func(key, value string) {
		delete(fdb.framesBySlot, key)
	})

	pos, ok := fdb.framePosByName[name]
	if ok {
		return nil, fmt.Errorf("not found frame name %s", name)
	}
	fdb.frames = slices.Delete(fdb.frames, pos, pos)

	return f, nil
}
