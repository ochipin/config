package parser

import (
	"fmt"
	"strings"
)

// Boolean 構造体は、key = value で渡された value 値から、真偽値を解析する
type Boolean struct {
	Value
	array bool
}

// NewBoolean 関数は、真偽値解析用ノードを生成する
func NewBoolean(p Node) Node {
	_, ok := p.(*Array)
	return &Boolean{
		Value: Value{
			text: p.Text(),
			cnt:  p.Getidx(),
			pos:  p.Pos(),
			end:  p.End(),
			key:  p.Keyname(),
		},
		array: ok,
	}
}

// Analyze 関数は、真偽値解析を実施する
func (boolean *Boolean) Analyze(b byte) (interface{}, error) {
	switch b {
	// 改行コードがあった時点で、終了とする
	case '\n':
		// 状態を元に戻す
		boolean.stat = ParserNone
		// 不正な指定方法であった場合エラーとする
		// if boolean.end < boolean.pos {
		//	return nil, fmt.Errorf("\"%s\" boolean invalid value", boolean.key)
		// }
		// 取得したパラメータが正しいか検証
		param := strings.Trim(boolean.Param(), " ")
		if param != "true" && param != "false" {
			return nil, fmt.Errorf("\"%s = %s\" boolean invalid value", boolean.key, param)
		}
		return param == "true", nil
	// コメント行
	case '#':
		boolean.end = boolean.cnt
		boolean.stat = ParserComment
	// 空白はスルーする
	case ' ':
	// 上記以外は、一時的に許可する
	default:
		// 配列内の整数の場合、, があった時点で終了とする
		if boolean.array {
			if b == ',' || b == ']' {
				boolean.end = boolean.cnt
				return boolean.Analyze('\n')
			}
		}
		// コメント行ではない場合、次要素を指す
		if boolean.stat != ParserComment {
			boolean.end = boolean.cnt + 1
		}
	}
	return nil, nil
}
