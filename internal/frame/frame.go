package frame

import (
	"fmt"
	"hash/fnv"
	"unsafe"

	"github.com/wanderer69/frm/internal/list"
	"github.com/wanderer69/frm/internal/value"
)

// Entry represents a node in the linked list for collision resolution.
type Entry struct {
	Key    string
	Values *list.List
	Next   *Entry
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
			index := ht.hash(e.Key) % uint32(newCapacity)
			newBuckets[index] = e
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
func (ht *HashTable) Put(key string, value value.Value) {
	index := ht.hash(key) % uint32(ht.capacity)
	entry := ht.buckets[index]

	// Check for existing key
	for e := entry; e != nil; e = e.Next {
		if e.Key == key {
			e.Values.Add(value)
			ht.checkResize()
			return
		}
	}

	// New key
	newEntry := &Entry{Key: key, Next: entry, Values: &list.List{}}
	newEntry.Values.Add(value)
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
func (ht *HashTable) Get(key string) *list.List {
	index := ht.hash(key) % uint32(ht.capacity)
	entry := ht.buckets[index]

	for e := entry; e != nil; e = e.Next {
		if e.Key == key {
			return e.Values
		}
	}
	return nil
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
			size += unsafe.Sizeof(*e) + uintptr(e.Values.Capacity())*unsafe.Sizeof(interface{}(nil)) + uintptr(len(e.Key))
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
			s = s + e.Key
			e.Values.Find(func(value value.Value) (bool, bool) {
				s = s + fmt.Sprintf(" %s ", value)
				return false, false
			})
			s = s + " "
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
