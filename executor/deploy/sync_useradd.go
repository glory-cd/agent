package deploy

import "sync"

// strIntMap is a struct which contains a sync map
// strIntMap will guarantee that key is string and values is int
type strIntMap struct {
	m sync.Map
}

func (iMap *strIntMap) Delete(key string) {
	iMap.m.Delete(key)
}

func (iMap *strIntMap) Load(key string) (value int, ok bool) {
	v, ok := iMap.m.Load(key)
	if v != nil {
		value = v.(int)
	}
	return
}

func (iMap *strIntMap) LoadOrStore(key string, value int) (actual int, loaded bool) {
	a, loaded := iMap.m.LoadOrStore(key, value)
	actual = a.(int)
	return
}

func (iMap *strIntMap) Range(f func(key string, value int) bool) {
	f1 := func(key, value interface{}) bool {
		return f(key.(string), value.(int))
	}
	iMap.m.Range(f1)
}

func (iMap *strIntMap) Store(key string, value int) {
	iMap.m.Store(key, value)
}
