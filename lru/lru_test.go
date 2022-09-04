package lru

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	keys := make([]string, 0)
	// 回调函数
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("1"))
	lru.Add("key2", String("12"))
	//if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1" {
	//	t.Fatalf("cache hit key1=1234 failed")
	//}
	//if _, ok := lru.Get("key2"); !ok {
	//	t.Fatalf("cache miss key2 failed")
	//}

	expect := []string{"key1", "key2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("equal :%s,%s", expect, keys)
	}

}
