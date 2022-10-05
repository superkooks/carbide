package common

import "github.com/vmihailenco/msgpack"

func WrapEvent(typ EventType, evt any) []byte {
	// msgpack should never return an error, as it is encoding to a bytes.Buffer
	b, _ := msgpack.Marshal(evt)
	out, _ := msgpack.Marshal(Event{
		Type: typ,
		Evt:  b,
	})
	return out
}

// RemoveFromSlice removes el from lst if it exists
func RemoveFromSlice[K comparable](lst []K, el K) []K {
	for k, v := range lst {
		if v == el {
			return append(lst[:k], lst[k+1:]...)
		}
	}
	return lst
}
