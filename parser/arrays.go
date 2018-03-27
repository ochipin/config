package parser

import (
	"fmt"
	"reflect"
	"time"
)

// Array 構造体は、配列を解析する
type Array struct {
	Value             // Value 構造体をミックスイン
	keep  int         // コメント時のキープ処理
	node  Node        // インナー配列を処理
	data  interface{} // 格納するデータ型
	next  bool        // カンマの位置や、連続したカンマの制御に使用
	kind  string      // 配列内で、型が違うデータがあった場合にエラーにするために使用する
	comp  bool        // 配列内の値解析完了フラグ
}

// NewArray 関数は、配列解析用ノードを生成する
func NewArray(p Node) Node {
	return &Array{
		Value: Value{
			text: p.Text(),
			stat: ParserBeginArray,
			cnt:  p.Getidx(),
			pos:  p.Pos(),
			end:  p.End(),
			key:  p.Keyname(),
		},
		next: true,
		data: nil,
	}
}

// データ追加関数
func (array *Array) adddata(data interface{}, inner bool) error {
	// 型情報と値情報を取得する
	valueof := reflect.ValueOf(data)
	kind := valueof.Type().String()
	if array.kind == "" {
		array.kind = kind
	}

	// 型が違う者同士の配列の場合、エラーとする
	if array.kind != kind {
		return fmt.Errorf("\"%s\" array of different types are confused", array.key)
	}

	// int, float32, string, bool, time.Time の方の場合
	switch data.(type) {
	case int:
		if array.data == nil {
			array.data = []int{}
		}
		array.data = append(array.data.([]int), data.(int))
		return nil
	case float32:
		if array.data == nil {
			array.data = []float32{}
		}
		array.data = append(array.data.([]float32), data.(float32))
		return nil
	case string:
		if array.data == nil {
			array.data = []string{}
		}
		array.data = append(array.data.([]string), data.(string))
		return nil
	case bool:
		if array.data == nil {
			array.data = []bool{}
		}
		array.data = append(array.data.([]bool), data.(bool))
		return nil
	case time.Time:
		if array.data == nil {
			array.data = []time.Time{}
		}
		array.data = append(array.data.([]time.Time), data.(time.Time))
		return nil
	}

	// array.data がまだ未登録の場合
	if array.data == nil {
		// ex) []int をベースに、[][]int 配列を作成する
		v := reflect.MakeSlice(reflect.SliceOf(valueof.Type()), 0, 0)
		// ex) [][]int に、 []int{0...} を格納する []int[0] = []int を格納するイメージ
		valueof = reflect.Append(v, valueof)
		array.data = valueof.Interface()
		return nil
	}
	// array.data が既にある場合、動的配列を生成する
	v := reflect.MakeSlice(reflect.ValueOf(array.data).Type(), 0, 0)
	// 配列の結合を行う。 []int[0] = array.data[0], []int[1] = array.data[1] を格納するイメージ
	v = reflect.AppendSlice(v, reflect.ValueOf(array.data))
	// 配列の追加を行う。 []int[2] = valueof を格納するイメージ
	valueof = reflect.Append(v, valueof)

	array.data = valueof.Interface()
	return nil
}

// Analyze 関数は、配列を解析する
func (array *Array) Analyze(b byte) (i interface{}, err error) {
	// コメント処理の場合
	if array.stat == ParserComment {
		if b != '\n' {
			return
		}
		array.stat = array.keep
	}

	// 配列内を解析する node を array が所持していた場合
	if array.node != nil {
		// 現在の参照ポイントを設定
		array.node.Cnt(array.cnt)
		// 解析関数をコール
		data, err := array.node.Analyze(b)
		if err != nil {
			return nil, err
		}
		// インナー配列か否かを判定するフラグを取得する
		_, inner := array.node.(*Array)
		// 値の取得が完了した場合、配列に要素を追加する
		if array.node.Stat() == ParserNone {
			err := array.adddata(data, inner)
			if err != nil {
				return nil, err
			}
			data = array.data
			array.node = nil
			array.stat = ParserBeginArray
			array.next = false // 次の要素を許可
			array.comp = true  // 値解析完了
		} else {
			return nil, nil
		}
		// インナー配列の場合、関数を抜ける
		if inner {
			return data, err
		}
	}

	// 値解析完了済みだが、カンマがなく、次の要素を指していた場合、エラーとする
	if array.comp {
		switch b {
		case '+', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 't', 'f', '$', '"', 39, '#':
			return nil, fmt.Errorf("\"%s\" separator is invalid", array.key)
		}
	}

	// 配列の中身を解析
	switch array.stat {
	case ParserBeginArray:
		i, err = array.parseBeginArray(b)
	}
	return
}

// 配列開始時に、コールされる
func (array *Array) parseBeginArray(b byte) (interface{}, error) {
	switch b {
	// インナー配列
	case '[':
		array.node = NewArray(array)
	// 配列終了
	case ']':
		array.stat = ParserNone
		return array.data, nil
	// 整数を解析する
	case '+', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		array.pos = array.cnt
		array.end = array.cnt + 1
		array.node = NewNumber(array)
	// 真偽値を解析
	case 't', 'f':
		array.pos = array.cnt
		array.end = array.cnt + 1
		array.node = NewBoolean(array)
	// 文字列を解析する
	case '"', 39:
		array.pos = array.cnt
		array.end = array.cnt + 1
		array.node = NewString(array)
	// 環境変数を解析する
	case '$':
		array.pos = array.cnt
		array.end = array.cnt + 1
		array.node = NewEnviron(array)
	// 空白はスルーする
	case ' ', '\n', '\t':
	// コメント行としてみなす
	case '#':
		array.keep = array.stat
		array.stat = ParserComment
	// , の場合は、セパレータとしてみなす
	case ',':
		// 連続したカンマは不正記述とする
		if array.next {
			return nil, fmt.Errorf("\"%s\" separator is invalid", array.key)
		}
		// カンマがあった場合は、次の要素とする
		array.next = true
		array.comp = false
	// 上記以外
	default:
		// 認めない文字列があった場合は、エラーとする
		return nil, fmt.Errorf("\"%s\" array value is invalid", array.key)
	}
	return nil, nil
}
