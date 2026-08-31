package store

// index is a plain in-memory hash map. No mutex — Store is the single lock authority (Plan §16).
type index struct {
	m map[string][]byte
}

func newIndex() *index {
	return &index{m: make(map[string][]byte)}
}

func (idx *index) set(key, val []byte) {
	// Copy to avoid aliasing caller buffers.
	k := string(append([]byte(nil), key...))
	v := append([]byte(nil), val...)
	idx.m[k] = v
}

func (idx *index) delete(key []byte) {
	delete(idx.m, string(key))
}

func (idx *index) get(key []byte) ([]byte, bool) {
	v, ok := idx.m[string(key)]
	if !ok {
		return nil, false
	}
	// Return copy so caller cannot mutate index.
	out := append([]byte(nil), v...)
	return out, true
}

func (idx *index) len() int { return len(idx.m) }
