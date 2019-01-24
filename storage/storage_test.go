package storage

import (
	"testing"
	"time"
)

type MyType int

/*
func Test__SET_STORAGE_STRUCT(t *testing.T) {
	var storage = make(Storage)
	now := time.Now()
	storage.Set("value", struct {
		Int     MyType
		Int64   int64
		String  string
		Bool    bool
		Ints    []MyType
		Ints64  []int64
		Strings []string
		Bools   []bool
		Time    time.Time
		Func    func()
	}{
		Int:     100,
		Int64:   200,
		String:  "string",
		Bool:    true,
		Ints:    []MyType{100, 101, 102},
		Ints64:  []int64{200, 201, 202},
		Strings: []string{"str1", "str2", "str3"},
		Bools:   []bool{true, false, true},
		Time:    now,
		Func:    func() {},
	})

	if storage.Int("value.Int") != 100 {
		t.Fatal("storage.Int", storage.Int("value.Int"), storage.IsInt("value.Int"))
	}

	i := storage.Get("value.Ints")
	for _, v := range i.([]interface{}) {
		fmt.Println(1, v.(int))
	}
}
*/

func Test__SET_STORAGE_TYPE_ERROR(t *testing.T) {
	now := time.Now()
	storage := NewStorage(now)

	if err := storage.IsInt("value.uint"); err == nil {
		t.Fatal("storage.Int")
	}
	if err := storage.IsInt64("value.uint64"); err == nil {
		t.Fatal("storage.Int64")
	}
	if err := storage.IsFloat("value.bool"); err == nil {
		t.Fatal("storage.Float32")
	}
	if err := storage.IsUint("value.string"); err == nil {
		t.Fatal("storage.uint")
	}
	if err := storage.IsUint64("value.string"); err == nil {
		t.Fatal("storage.uint64")
	}
	if err := storage.IsStr("value.int"); err == nil {
		t.Fatal("storage.string")
	}
	if storage.IsTime("value.int") == nil {
		parse, _ := time.Parse(time.RFC3339, now.Format(time.RFC3339))
		if storage.Time("value.time").Unix() != parse.Unix() {
			t.Fatal("storage.Time")
		}
	}
	if err := storage.IsBool("value.string"); err == nil {
		t.Fatal("storage.string")
	}
	if err := storage.IsBool("value.noname"); err == nil {
		t.Fatal("storage.string")
	}
	if err := storage.IsSlice("value.string"); err == nil {
		t.Fatal("storage.string")
	}
	if err := storage.IsMap("value.string"); err == nil {
		t.Fatal("storage.string")
	}
	if storage.Get("value.func") != nil {
		t.Fatal("storage.Get")
	}

	storage.Set("path.to.dir", "dirname")
	storage.Set("path.to.dir", "change_dirname")
	storage.Set("path.to.dir.name", "ok")

	if storage.Str("path.to.dir.name") != "ok" {
		t.Fatal("storage.Str")
	}

	var value = new(int)
	*value = 999
	storage.Set("path.to.int", value)
	if storage.Int("path.to.int") != 999 {
		t.Fatal("storage.Int")
	}
	if storage.Str("path.to.dir.name") != "ok" {
		t.Fatal("storage.Str")
	}
	storage.Set("path.to.int", nil)
	storage["ptr"] = map[string]interface{}{
		"path": map[int]interface{}{
			1: struct {
				Int int
			}{999},
		},
	}
	if storage.Get("ptr.path.name") != nil {
		t.Fatal("storage.Get")
	}
}

func Test__SET_STORAGE_TYPES(t *testing.T) {
	now := time.Now()
	storage := NewStorage(now)

	if err := storage.IsInt("value.int"); err != nil {
		t.Fatal("storage.Int")
	}
	if err := storage.IsInt64("value.int64"); err != nil {
		t.Fatal("storage.Int64")
	}
	if err := storage.IsFloat("value.float32"); err != nil {
		t.Fatal("storage.Float32")
	}
	if err := storage.IsUint("value.uint"); err != nil {
		t.Fatal("storage.uint")
	}
	if err := storage.IsUint64("value.uint64"); err != nil {
		t.Fatal("storage.uint64")
	}
	if err := storage.IsStr("value.string"); err != nil {
		t.Fatal("storage.string")
	}
	if storage.IsTime("value.time") == nil {
		parse, _ := time.Parse(time.RFC3339, now.Format(time.RFC3339))
		if storage.Time("value.time").Unix() != parse.Unix() {
			t.Fatal("storage.Time")
		}
	}
	if err := storage.IsBool("value.bool"); err != nil {
		t.Fatal("storage.string")
	}
	if err := storage.IsSlice("value.slice"); err != nil {
		t.Fatal("storage.string")
	}
	if err := storage.IsMap("value.map"); err != nil {
		t.Fatal("storage.string")
	}
	if storage.Get("value.func") != nil {
		t.Fatal("storage.Get")
	}
}

func Test__SET_STORAGE_VALUES(t *testing.T) {
	now := time.Now()
	storage := NewStorage(now)

	if storage.Int("value.int") != 100 {
		t.Fatal("value.int")
	}
	if storage.Int64("noname") != 0 || storage.Int64("value.string") != 0 {
		t.Fatal("int64")
	}
	if storage.Str("value.string") != "string" {
		t.Fatal("string")
	}
	if storage.Bool("value.bool") != true {
		t.Fatal("string")
	}
	if storage.Uint("value.uint") != 1000 {
		t.Fatal("value.uint")
	}
	if storage.Uint64("noname") != 0 || storage.Uint64("value.string") != 0 {
		t.Fatal("uint64")
	}
	if storage.Float32("value.float32") <= 0 {
		t.Fatal("value.float32")
	}
	if storage.Float64("noname") != 0 || storage.Float64("value.string") != 0 {
		t.Fatal("Float64")
	}
	if storage.Map("value") == nil {
		t.Fatal("map")
	}
	if storage.Slice("value.slice") == nil {
		t.Fatal("slice")
	}
}

type Info struct {
	App    string
	Port   int
	Detail string
	Flags  []bool
}

func Test__SET_STORAGE_UNMARSHAL(t *testing.T) {
	var storage = make(Storage)
	storage.Set("config.app", "app")
	storage.Set("config.port", 8080)
	storage.Set("config.detail", "sample app")
	storage.Set("config.flags", []bool{true, true, false})
	storage.Set("datetime", time.Now())

	var data map[string]interface{}
	if err := storage.Unmarshal("config", &data); err != nil {
		t.Fatal(err)
	}
	var info Info
	if err := storage.Unmarshal("config", &info); err != nil {
		t.Fatal(err)
	}
	var slice []bool
	if err := storage.Unmarshal("config.flags", &slice); err != nil {
		t.Fatal(err)
	}

	// errors
	if err := storage.Unmarshal("noname", &data); err == nil {
		t.Fatal("unmarshal")
	}
	if err := storage.Unmarshal("config", info); err == nil {
		t.Fatal("unmarshal")
	}
	if err := storage.Unmarshal("config.flags", info); err == nil {
		t.Fatal("unmarshal")
	}
	if err := storage.Unmarshal("config", &slice); err == nil {
		t.Fatal("unmarshal")
	}
	if err := storage.Unmarshal("config.flags", &info); err == nil {
		t.Fatal("unmarshal")
	}
	if err := storage.Unmarshal("datetime", &slice); err == nil {
		t.Fatal("unmarshal")
	}
}

func NewStorage(now time.Time) Storage {
	var storage = make(Storage)

	// データをセット
	storage.Set("value", map[string]interface{}{
		"int":     100,
		"int64":   int64(200),
		"float32": 3.141592,
		"float64": 6.283184,
		"uint":    uint(1000),
		"uint64":  uint64(2000),
		"string":  "string",
		"bool":    true,
		"time":    now,
		"func":    func() {},
		"slice": [][]int{
			[]int{1, 2, 3},
			[]int{4, 5, 6},
			[]int{7, 8, 9},
		},
		"pointer": new(uint),
		"byte":    []byte("[]byte"),
		"struct": struct {
			Int      MyType
			Int64    int64
			String   string
			Bool     bool
			Ints     []MyType
			Ints64   []int64
			Strings  []string
			Bools    []bool
			Time     time.Time
			Func     func()
			noexport bool
			Pointer  *int
		}{
			Int:      100,
			Int64:    200,
			String:   "string",
			Bool:     true,
			Ints:     []MyType{100, 101, 102},
			Ints64:   []int64{200, 201, 202},
			Strings:  []string{"str1", "str2", "str3"},
			Bools:    []bool{true, false, true},
			Time:     now,
			Func:     func() {},
			noexport: false,
			Pointer:  new(int),
		},
		"map": map[string]interface{}{
			"int":     100,
			"int64":   int64(200),
			"float32": 3.141592,
			"float64": 6.283184,
			"uint":    uint(1000),
			"uint64":  uint64(2000),
			"string":  "string",
			"bool":    true,
			"time":    now,
			"func":    func() {},
			"slice": [][]int{
				[]int{1, 2, 3},
				[]int{4, 5, 6},
				[]int{7, 8, 9},
			},
		},
	})
	storage.Set("func", func() {})
	storage.Set("default", complex128(3.14))
	storage.Set("a.b.c", 1000)

	return storage
}
