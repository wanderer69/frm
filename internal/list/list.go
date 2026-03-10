package list

import (
	"strings"
	"sync"
	"unsafe"

	"github.com/wanderer69/frm/internal/value"
)

type ListItem struct {
	Value value.Value
	Next  *ListItem
}

type List struct {
	mu   sync.Mutex
	Root *ListItem
	End  *ListItem
	size int
}

func (l *List) Size() int {
	return l.size
}

func (l *List) Add(value value.Value) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := &ListItem{
		Value: value,
	}
	l.size += 1
	if l.Root == nil {
		l.Root = n
		l.End = n
		return
	}
	l.End.Next = n
	l.End = n
}

func (l *List) Find(fn func(value value.Value) (bool, bool)) {
	var prev *ListItem
	for e := l.Root; e != nil; e = e.Next {
		if fn != nil {
			isDeleted, isFinished := fn(e.Value)
			if isDeleted {
				l.mu.Lock()
				if prev != nil {
					prev.Next = e.Next
				} else {
					p := l.Root
					l.Root = e.Next
					if p == l.End {
						l.End = e.Next
					}
				}
				l.size -= 1
				l.mu.Unlock()
			}
			if isFinished {
				break
			}
		}
		prev = e
	}
}

func (l *List) Capacity() uintptr {
	size := unsafe.Sizeof(*l)
	for e := l.Root; e != nil; e = e.Next {
		size += unsafe.Sizeof(*e)
		size += unsafe.Sizeof(e.Value)
	}
	return size
}

func (l *List) Get(n int) value.Value {
	i := 0
	for e := l.Root; e != nil; e = e.Next {
		if i == n {
			return e.Value
		}
		i++
	}
	return nil
}

func (l *List) String() (string, error) {
	ss := "["
	sl := []string{}
	for e := l.Root; e != nil; e = e.Next {
		vv, err := e.Value.To(value.ValueTypeString)
		if err != nil {
			return "", err
		}
		sl = append(sl, vv.Value().(string))
	}
	ss += strings.Join(sl, ", ") + "]"
	return ss, nil
}
