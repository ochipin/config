package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ochipin/config/parser"
)

// mapにデータを追加/上書きする
func setdata(all map[string]interface{}, data interface{}, keys []string) {
	// app.key.name ---> [app key], [name] の2つへ分離
	last := keys[len(keys)-1]
	keys = keys[:len(keys)-1]
	// [app key] キーの値のみ検証
	for _, key := range keys {
		if v, ok := all[key].(map[string]interface{}); ok {
			// data[key] が map の場合、次の要素へ
			all = v
		} else {
			// data[key] が存在しない場合、mapを生成して次へ
			all[key] = make(map[string]interface{})
			all = all[key].(map[string]interface{})
		}
	}
	// 最後に、data[app][key][name] = data とする
	all[last] = data
}

// map1にmap2をマージする。既に存在する要素がある場合、上書きを実施する
func mergedata(map1, map2 map[string]interface{}, keys ...string) {
	// マージしたいデータをループで全データを処理
	// map[app][key][name] = "merge"
	for key, value := range map2 {
		keys = append(keys, key)
		if v, ok := value.(map[string]interface{}); ok {
			// map[app] も map の場合、再帰する
			mergedata(map1, v, keys...)
		} else {
			// 終端にたどり着いた時点で、データをマージする
			setdata(map1, value, keys)
		}
		if len(keys) > 0 {
			keys = keys[:len(keys)-1]
		}
	}
}

// パースしたデータを構造体に格納する
func unmarshal(data map[string]interface{}, i interface{}) error {
	buf, _ := json.Marshal(data)

	if err := json.Unmarshal(buf, i); err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			if err.Struct != "" || err.Field != "" {
				return fmt.Errorf("marshal error: " + err.Value + " into struct field " + err.Struct + "." + err.Field + " of type " + err.Type.String())
			}
		}
		return fmt.Errorf("marshal error: %s", err)
	}

	return nil
}

// 指定した設定ファイルの内容をパースし、構造体、またはマップに格納する
func Parse(path, mode string, i interface{}) error {
	// 指定されたパスから、設定ファイルを読み込む
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// 読み込んだ設定ファイル内容を map[string]interface{} へパースする
	p, err := parser.Parse(buf)
	if err != nil {
		return err
	}

	// パース内容を変数へ格納
	data := p.Data().(map[string]interface{})

	// 全体設定情報を変数へ格納
	_, ok1 := data["_all_"]
	_, ok2 := data[mode]

	if ok1 && ok2 {
		// 全体設定領域、マージデータ双方存在する場合は、データをインターフェースへ格納する
		all := data["_all_"].(map[string]interface{})
		mrg := data[mode].(map[string]interface{})
		mergedata(all, mrg)
		return unmarshal(all, i)
	} else if ok1 {
		// 全体設定領域しか存在しない場合、全体設定領域のみをインターフェースへ格納する
		return unmarshal(data["_all_"].(map[string]interface{}), i)
	} else if ok2 {
		// マージデータしか存在しない場合、マージデータのみをインターフェースへ格納する
		return unmarshal(data[mode].(map[string]interface{}), i)
	}

	// データが存在しない場合、エラーを返却する
	return fmt.Errorf("no configuration")
}

// 設定ファイル操作構造体
type Config struct {
	data map[string]interface{}
}

// 登録されている全データを取得する
func (c *Config) DataAll() map[string]interface{} {
	return c.data
}

// 一部のモード名のみを抜き出す
func (c *Config) Data(mode string) map[string]interface{} {
	if m, ok := c.data[mode]; ok {
		return m.(map[string]interface{})
	}
	return nil
}

// 構造体、またはマップにデータを格納する
func (c *Config) Unmarshal(data map[string]interface{}, i interface{}) error {
	return unmarshal(data, i)
}

// データ1にデータ2をマージする
func (c *Config) Merge(data1, data2 map[string]interface{}) {
	mergedata(data1, data2)
}

// ParseMode 関数は、設定ファイル内容を解析、パースする。冒頭にモード指定がされていないと設定ファイルを解析しない
func ParseMode(path string) (*Config, error) {
	// 指定されたパスから、設定ファイルを読み込む
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 読み込んだ設定ファイル内容を map[string]interface{} へパースする
	p, err := parser.ParseModeAll(buf)
	if err != nil {
		return nil, err
	}

	// 設定ファイルパース内容を操作する構造体を返却する
	return &Config{p.Data().(map[string]interface{})}, nil
}
