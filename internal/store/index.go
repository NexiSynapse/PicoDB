package store

// index maintains an in-memory key-to-value directory (Bitcask-style).
// Notice that index has no internal mutex because Store is the single lock authority.
type index struct {
	m map[string][]byte
}

func newIndex() *index {
	return &index{
		m: make(map[string][]byte),
	}
}

func (i *index) set(key, val []byte) {
	vCopy := make([]byte, len(val))
	copy(vCopy, val)
	i.m[string(key)] = vCopy
}

func (i *index) delete(key []byte) {
	delete(i.m, string(key))
}

func (i *index) get(key []byte) ([]byte, bool) {
	val, ok := i.m[string(key)]
	if !ok {
		return nil, false
	}
	vCopy := make([]byte, len(val))
	copy(vCopy, val)
	return vCopy, true
}

func (i *index) len() int {
	return len(i.m)
}
