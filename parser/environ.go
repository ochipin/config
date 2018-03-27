package parser

import (
	"fmt"
	"os"
	"strings"
)

// Environ 構造体は、環境変数を解析する
type Environ struct {
	Value
	array bool
}

// NewEnviron 関数は、環境変数解析用ノードを生成する
func NewEnviron(p Node) Node {
	_, ok := p.(*Array)
	return &Environ{
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

// Analyze 関数は、環境変数を解析する
func (environ *Environ) Analyze(b byte) (interface{}, error) {
	switch b {
	// 改行コードの時点で終了とする
	case '\n':
		// 状態を元に戻す
		environ.stat = ParserNone
		// 不正な指定方法であった場合エラーとする
		// if environ.end < environ.pos {
		//	return nil, fmt.Errorf("\"%s\" environ invalid value", environ.key)
		// }
		// 取得したパラメータが正しいか検証
		param := strings.Trim(environ.Param(), " ")
		if param == "" || param[1:] == "" {
			return nil, fmt.Errorf("\"%s\" environ invalid value", environ.key)
		}
		if strings.Index(param, " ") != -1 {
			return nil, fmt.Errorf("\"%s = %s\" environ invalid value", environ.key, param)
		}
		return os.Getenv(param[1:]), nil
	// 空白はスルーする
	case ' ':
	// コメント行
	case '#':
		environ.end = environ.cnt
		environ.stat = ParserComment
	// 上記以外は、認める
	default:
		// 配列内の値の場合、, or ] があった時点で終了とする
		if environ.array {
			if b == ',' || b == ']' {
				environ.end = environ.cnt
				return environ.Analyze('\n')
			}
		}
		// コメント行ではない場合、次要素を指す
		if environ.stat != ParserComment {
			environ.end = environ.cnt + 1
		}
	}
	return nil, nil
}
