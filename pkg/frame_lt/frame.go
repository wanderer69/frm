package framelt

import (
	"fmt"
	"hash/fnv"
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
	Name  string
	Slots *HashTable
}

// NewFrame creates a new frame.
func NewFrame(name string) *Frame {
	return &Frame{
		Name:  name,
		Slots: NewHashTable(),
	}
}

func (f *Frame) String() string {
	s := fmt.Sprintf("%s {", f.Name)
	for _, entry := range f.Slots.buckets {
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
	if n < 0 || n > len(f.Slots.buckets) {
		return nil
	}
	return f.Slots.buckets[n]
}

func (f *Frame) Len() int {
	return len(f.Slots.buckets)
}

func (f *Frame) Set(key string, value value.Value) {
	f.Slots.Put(key, value.Type().ToString())
}

func (f *Frame) Get(key string, args ...int) value.Value {
	/*
		pos := 0
		if len(args) != 0 {
			pos = args[0]
		}
	*/
	l, ok := f.Slots.Get(key)
	if !ok {
		return constants.ValueNil
	}
	return &valueType.ValueString{String: l}
}
