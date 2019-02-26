package storage

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type typeName int

func (types typeName) String() string {
	var result string
	switch types {
	case typeInt:
		result = "int"
	case typeInt64:
		result = "int64"
	case typeString:
		result = "string"
	case typeBool:
		result = "bool"
	case typeFloat:
		result = "float"
	case typeUint:
		result = "uint"
	case typeUint64:
		result = "uint64"
	case typeMap:
		result = "map"
	case typeTime:
		result = "datetime"
	case typeSlice:
		result = "slice"
	}
	return result
}

// Storage : mapデータ管理型
type Storage map[string]interface{}

const (
	typeInt    typeName = iota // int 型
	typeInt64                  // int64 型
	typeString                 // string 型
	typeBool                   // bool 型
	typeFloat                  // float
	typeUint                   // uint
	typeUint64                 // uint64
	typeMap                    // map[string]interface{} 型
	typeTime                   // time.Time 型
	typeSlice                  // []interface{} 型
)

// Get : 指定した keyname から該当するデータを取得する
func (storage Storage) Get(keyname string) interface{} {
	var copy interface{} = storage
	// 指定した keyname から、該当する値を取得する。
	// 'path.to.name' => [path, to, name] として解析する
	for _, key := range strings.Split(keyname, ".") {
		v, ok := copy.(Storage)
		// Storage 型ではない場合、map[string]interface{}型へキャストする
		if !ok {
			// map[string]interface{}型でもない場合、nilを返却する
			if v, ok = copy.(map[string]interface{}); !ok {
				return nil
			}
		}
		// 指定した keyname が存在しない場合、nil を返却する
		if _, ok = v[key]; !ok {
			return nil
		}
		copy = v[key]
	}
	// 指定した keyname に該当するデータを返却する
	return copy
}

// Set : keyname, value とセットでデータを登録する
func (storage Storage) Set(keyname string, value interface{}) {
	var mapcopy map[string]interface{} = storage
	var names = strings.Split(keyname, ".")
	var lastname = names[len(names)-1]

	// セットする value の位置を選定する
	// 'path.to.name' => [path, to] として選定する。最後の name は、ループを抜けた後にセットする。
	for _, key := range names[:len(names)-1] {
		// 指定したkeynameで値が見つかった
		if v, ok := mapcopy[key]; ok {
			// 見つかった値が map[string]interface{} 型か確認する
			if _, ok := v.(map[string]interface{}); ok {
				// map[string]interface{} 型の場合、次の要素へ
				mapcopy = v.(map[string]interface{})
			} else {
				// map[string]interface{} 型ではない場合、mapを生成する
				mapcopy[key] = make(map[string]interface{})
				mapcopy = mapcopy[key].(map[string]interface{})
			}
		} else {
			// 指定したkeynameで値が見つからなかった場合、mapを生成する
			mapcopy[key] = make(map[string]interface{})
			mapcopy = mapcopy[key].(map[string]interface{})
		}
	}

	// nil データの場合、指定されたキーを削除する
	if value == nil {
		if _, ok := mapcopy[lastname]; ok {
			delete(mapcopy, lastname)
		}
		return
	}

	// ポインタの場合、エレメントを取得する
	ref := reflect.ValueOf(value)
	for ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}

	// データ登録
	if v := toCastData(ref); v != nil {
		mapcopy[lastname] = v
	}
}

// Unmarshal : 指定したキー名のデータを、マップ、または構造体に格納する
func (storage Storage) Unmarshal(name string, i interface{}) error {
	// データを取得
	value := storage.Get(name)
	// データが、存在しない場合エラーとする
	if value == nil {
		return fmt.Errorf("'%s' is undefined", name)
	}

	var err error
	if v, ok := value.(map[string]interface{}); ok {
		// 取得データが、マップの場合
		err = unmarshalMap(v, i)
	} else if v, ok := value.([]interface{}); ok {
		// 取得データが、スライスの場合
		err = unmarshalSlice(v, i)
	} else {
		err = fmt.Errorf("'%s' type not slice or map", name)
	}

	return err
}

// 格納データがマップ、または構造体の場合コールされる
func unmarshalMap(v map[string]interface{}, i interface{}) error {
	ref := reflect.ValueOf(i)

	// 格納先がポインタではない場合、エラーとする
	if ref.Kind() != reflect.Ptr || ref.IsNil() || ref.IsValid() == false {
		return fmt.Errorf("unmarshal error. arguments no pointer")
	}
	ref = ref.Elem()

	if ref.Kind() == reflect.Map && ref.Type().Key().Kind() == reflect.String {
		// データ格納先がマップの場合、リフレクションでデータを格納する
		ref.Set(reflect.ValueOf(v))
	} else {
		// データ格納先がマップ以外の場合は、json.Unmarshalでデータを格納する
		buf, _ := json.Marshal(v)
		if err := json.Unmarshal(buf, i); err != nil {
			return err
		}
	}

	return nil
}

// 格納データがスライスの場合コールされる
func unmarshalSlice(v []interface{}, i interface{}) error {
	ref := reflect.ValueOf(i)

	// 格納先がポインタではない場合、エラーとする
	if ref.Kind() != reflect.Ptr || ref.IsNil() || ref.IsValid() == false {
		return fmt.Errorf("unmarshal error. arguments no pointer")
	}
	ref = ref.Elem()

	// スライスデータを格納
	buf, _ := json.Marshal(v)
	if err := json.Unmarshal(buf, i); err != nil {
		return err
	}

	return nil
}

// Int : int型として値を取得する
func (storage Storage) Int(name string) (res int) {
	var v = storage.Int64(name)
	res = int(v)
	return
}

// IsInt : int型かチェックする
func (storage Storage) IsInt(name string) error {
	return storage.hasItem(name, typeInt)
}

// Int64 : int64型として値を取得する
func (storage Storage) Int64(name string) (res int64) {
	var i = storage.Get(name)
	if i == nil {
		res = 0
		return
	}
	var ref = reflect.ValueOf(i)
	if ref.IsValid() {
		switch ref.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			res = ref.Convert(reflect.TypeOf(int64(0))).Interface().(int64)
		default:
			res = 0
		}
	}
	return
}

// IsInt64 : int64型かチェックする
func (storage Storage) IsInt64(name string) error {
	return storage.hasItem(name, typeInt64)
}

// Str : 文字列型として取得する
func (storage Storage) Str(name string) (res string) {
	if storage.IsStr(name) == nil {
		res = storage.Get(name).(string)
	}
	return
}

// IsStr : 文字列型かチェックする
func (storage Storage) IsStr(name string) error {
	return storage.hasItem(name, typeString)
}

// Bool : bool型として値を取得する
func (storage Storage) Bool(name string) (res bool) {
	if storage.IsBool(name) == nil {
		res = storage.Get(name).(bool)
	}
	return
}

// IsBool : bool型かチェックする
func (storage Storage) IsBool(name string) error {
	return storage.hasItem(name, typeBool)
}

// Uint : uint 型でデータを取得する
func (storage Storage) Uint(name string) (res uint) {
	var v = storage.Uint64(name)
	res = uint(v)
	return
}

// IsUint : uint型かチェックする
func (storage Storage) IsUint(name string) error {
	return storage.hasItem(name, typeUint)
}

// Uint64 : uint64 型でデータを取得する
func (storage Storage) Uint64(name string) (res uint64) {
	var i = storage.Get(name)
	if i == nil {
		res = 0
		return
	}
	var ref = reflect.ValueOf(i)
	if ref.IsValid() {
		switch ref.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			res = ref.Convert(reflect.TypeOf(uint64(0))).Interface().(uint64)
		default:
			res = 0
		}
	}
	return
}

// IsUint64 : uint64型かチェックする
func (storage Storage) IsUint64(name string) error {
	return storage.hasItem(name, typeUint64)
}

// Float32 : float32 型でデータを取得する
func (storage Storage) Float32(name string) (res float32) {
	var v = storage.Float64(name)
	res = float32(v)
	return
}

// Float64 : float64 型でデータを取得する
func (storage Storage) Float64(name string) (res float64) {
	var i = storage.Get(name)
	if i == nil {
		res = 0
		return
	}
	var ref = reflect.ValueOf(i)
	if ref.IsValid() {
		switch ref.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			res = ref.Convert(reflect.TypeOf(float64(0))).Interface().(float64)
		default:
			res = 0
		}
	}
	return
}

// IsFloat : Float型かチェックする
func (storage Storage) IsFloat(name string) error {
	return storage.hasItem(name, typeFloat)
}

// Map : Map型としてデータを取得する
func (storage Storage) Map(name string) (res map[string]interface{}) {
	if storage.IsMap(name) == nil {
		res = storage.Get(name).(map[string]interface{})
	}
	return
}

// IsMap : Map型かチェックする
func (storage Storage) IsMap(name string) error {
	return storage.hasItem(name, typeMap)
}

// Slice : []interface{} 型でデータを取得する
func (storage Storage) Slice(name string) (res []interface{}) {
	if storage.IsSlice(name) == nil {
		v := reflect.ValueOf(storage.Get(name))
		for i := 0; i < v.Len(); i++ {
			res = append(res, v.Index(i).Interface())
		}
	}
	return
}

// IsSlice : []interface{}型かチェックする
func (storage Storage) IsSlice(name string) error {
	if err := storage.hasItem(name, typeSlice); err != nil {
		v := reflect.ValueOf(storage.Get(name))
		if v.IsValid() == false || v.Kind() != reflect.Slice {
			return err
		}
	}
	return nil
}

// Time : 日付型としてデータを取得する
func (storage Storage) Time(name string) (res time.Time) {
	if storage.IsTime(name) == nil {
		res, _ = time.Parse(time.RFC3339, storage.Get(name).(string))
	}
	return
}

// IsTime : 日付型かチェックする
func (storage Storage) IsTime(name string) error {
	return storage.hasItem(name, typeTime)
}

// 指定したデータが正しく登録されているか確認する
func (storage Storage) hasItem(keyname string, types typeName) error {
	v := storage.Get(keyname)
	// keyname で登録されているデータを所持していない場合、エラーを返却する
	if v == nil {
		return fmt.Errorf("'%s' is undefined", keyname)
	}
	var ok bool
	switch types {
	case typeInt:
		_, ok = v.(int)
	case typeInt64:
		_, ok = v.(int64)
	case typeString:
		_, ok = v.(string)
	case typeBool:
		_, ok = v.(bool)
	case typeFloat:
		_, ok = v.(float64)
	case typeSlice:
		_, ok = v.([]interface{})
	case typeUint:
		_, ok = v.(uint)
	case typeUint64:
		_, ok = v.(uint64)
	case typeMap:
		_, ok = v.(map[string]interface{})
	case typeTime:
		if _, err := time.Parse(time.RFC3339, fmt.Sprint(v)); err == nil {
			ok = true
		}
	}
	// キャスト失敗した場合は、エラーを返却する
	if !ok {
		return fmt.Errorf("'%s' is not '%s' type", keyname, types)
	}
	// 正しくキャスト可能であれば、nilを返却する
	return nil
}

// 構造体を登録された場合コールされる
func toStruct(value map[string]interface{}, ref reflect.Value) {
	for i := 0; i < ref.NumField(); i++ {
		ftyp := ref.Type().Field(i)
		// エクスポート不可能なメンバは除外する
		if ftyp.Name[0] >= 97 && ftyp.Name[0] <= 122 {
			continue
		}
		// 関数は除外する
		if ftyp.Type.Kind() == reflect.Func {
			continue
		}
		fval := ref.Field(i)
		// ポインタの場合、エレメントを指すようにする
		for fval.Kind() == reflect.Ptr {
			fval = fval.Elem()
		}
		if v := toCastData(fval); v != nil {
			value[ftyp.Name] = v
		}
	}
}

// マップを登録された場合コールされる
func toMap(value map[string]interface{}, ref reflect.Value) {
	for _, key := range ref.MapKeys() {
		fval := ref.MapIndex(key)
		fval = reflect.ValueOf(fval.Interface())
		// ポインタの場合、エレメントを指すようにする
		for fval.Kind() == reflect.Ptr {
			fval = fval.Elem()
		}
		// 関数は除外する
		if fval.Type().Kind() == reflect.Func {
			continue
		}
		if v := toCastData(fval); v != nil {
			value[key.String()] = v
		}
	}
}

// 登録された各データをキャストする
func toCastData(fval reflect.Value) interface{} {
	var result interface{}

	switch fval.Kind() {
	// struct
	case reflect.Struct:
		// time.Time の場合、RFC3339で文字列へ変換する
		if fval.Type().String() == "time.Time" {
			t := fval.Interface().(time.Time)
			result = t.Format(time.RFC3339)
		} else {
			result = make(map[string]interface{})
			toStruct(result.(map[string]interface{}), fval)
		}
	// slice
	case reflect.Slice:
		var array []interface{}
		if v, ok := fval.Interface().([]byte); ok {
			// []byte の場合は、文字列として扱う
			result = string(v)
		} else {
			// 配列の場合は、[]interface{} へまとめる
			for i := 0; i < fval.Len(); i++ {
				array = append(array, toCastData(fval.Index(i)))
			}
		}
		result = array
	// map
	case reflect.Map:
		result = make(map[string]interface{})
		toMap(result.(map[string]interface{}), fval)
	// int
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		result = fval.Convert(reflect.TypeOf(int(0))).Interface().(int)
	// int64
	case reflect.Int64:
		result = fval.Convert(reflect.TypeOf(int64(0))).Interface().(int64)
	// float
	case reflect.Float32, reflect.Float64:
		result = fval.Convert(reflect.TypeOf(float64(0))).Interface().(float64)
	// string
	case reflect.String:
		result = fval.Convert(reflect.TypeOf(string(""))).Interface().(string)
	// uint
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		result = fval.Convert(reflect.TypeOf(uint(0))).Interface().(uint)
	// uint64
	case reflect.Uint64:
		result = fval.Convert(reflect.TypeOf(uint64(0))).Interface().(uint64)
	// bool
	case reflect.Bool:
		result = fval.Convert(reflect.TypeOf(bool(false))).Interface().(bool)
	// func: 無視する
	case reflect.Func:
		result = nil
	// 上記以外の場合
	default:
		result = fval.Interface()
	}
	return result
}
