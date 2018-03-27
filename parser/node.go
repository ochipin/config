package parser

import (
	"fmt"
	"strings"
)

// Node インターフェースは、Value構造体を継承した構造体を取り扱う
type Node interface {
	Text() []byte                      // Value.text を返却する
	Stat() int                         // Value.stat を返却する
	Pos() int                          // Value.pos を返却する
	End() int                          // Value.end を返却する
	Cnt(int)                           // Value.cnt をセットする
	Getidx() int                       // Value.cnt を返却する
	Keyname() string                   // Value.keyname を返却する
	Analyze(byte) (interface{}, error) // 構文解析関数
}

// Value 構造体は、あるルールに基づく文字列から、適切な型と値
type Value struct {
	text []byte // 解析する文字列
	stat int    // 解析状態
	pos  int    // 解析開始位置
	end  int    // 解析終了位置
	cnt  int    // textの現在参照している位置
	key  string // キー名
}

// Text 関数は、処理する文字列を[]byteで返却する
func (v *Value) Text() []byte {
	return v.text
}

// Pos 関数は、解析する位置を返却する
func (v *Value) Pos() int {
	return v.pos
}

// End 関数は、解析終了した位置を返却する
func (v *Value) End() int {
	return v.end
}

// Cnt 関数は、現在参照している位置をセットする
func (v *Value) Cnt(i int) {
	v.cnt = i
}

// Getidx 関数は、現在参照している箇所を返却する
func (v *Value) Getidx() int {
	return v.cnt
}

// Stat 関数は、現在のステータスを返却する
func (v *Value) Stat() int {
	return v.stat
}

// Keyname 関数は、キー名を取得する
func (v *Value) Keyname() string {
	return v.key
}

// Param 関数は、値を返却する
func (v *Value) Param() string {
	return string(v.text[v.pos:v.end])
}

// Trim 関数は、指定された1文字のみ、先頭/最後尾から除去する
func (v *Value) Trim(param string, b byte) string {
	if len(param) > 0 {
		if param[0] == b {
			param = param[1:]
		}
	}
	if len(param) > 0 {
		if param[len(param)-1] == b {
			param = param[:len(param)-1]
		}
	}
	return param
}

// Prev は、1つ前のバイト値を取得する
func (v *Value) Prev(i int) byte {
	if v.cnt == 0 || i == 0 || v.cnt-i < 0 {
		return 0
	}
	return v.text[v.cnt-i]
}

// Set 関数は、取得したキーで、値をdataへ格納する
func Set(keynames string, value, i interface{}, mode string) error {
	// ex) app.key.name ---> app, key, name へ分割して処理
	var keys = strings.Split(mode+"."+keynames, ".")
	var last string
	if len(keys) > 1 {
		// app key name 等複数ある場合は、最後のキーのみ、別変数へ格納する
		last = keys[len(keys)-1]
		keys = keys[:len(keys)-1]
	}

	// 指定されたキー分ループし、情報を構築する
	data, _ := i.(map[string]interface{})
	for _, key := range keys {
		if v1, ok := data[key]; !ok {
			// data[key] がnilの場合、生成する
			data[key] = make(map[string]interface{})
			data = data[key].(map[string]interface{})
		} else {
			if v2, ok := v1.(map[string]interface{}); ok {
				// 存在した場合、次の要素を指定する
				data = v2
			} else {
				// 存在するが、map[string]interface{}型ではない場合、エラーとする
				return fmt.Errorf("\"%s\" already exists.", keynames)
			}
		}
	}

	// 既にデータがある場合は、エラーとする
	if _, ok := data[last]; ok {
		return fmt.Errorf("\"%s\" already exists.", keynames)
	}
	// 値をセット
	data[last] = value

	return nil
}
