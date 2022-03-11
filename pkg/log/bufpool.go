package log

import "sync"

const size = 256

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, size)
	},
}

// get return an empty []byte
func get() []byte {
	return pool.Get().([]byte)[:0]
}

func put(bytes []byte) {
	if cap(bytes) < size*4 {
		pool.Put(bytes)
	}
}
