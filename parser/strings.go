package parser

import (
	"fmt"
	"strings"
)

// String 構造体は、環境変数を解析する
type String struct {
	Value
	keep  int
	quote byte
	prev  [3]byte
	array bool
}

// NewString 関数は、文字列解析用ノードを生成する
func NewString(p Node) Node {
	text := p.Text()
	cnt := p.Getidx()
	_, ok := p.(*Array)
	return &String{
		Value: Value{
			text: text,
			stat: ParserBeginString,
			cnt:  cnt,
			pos:  p.Pos(),
			end:  p.End(),
			key:  p.Keyname(),
		},
		array: ok,
		quote: text[cnt],
	}
}

// Analyze 関数は、文字列解析を実施する
func (str *String) Analyze(b byte) (i interface{}, err error) {
	// コメント行があった場合は、スルーする
	if str.stat == ParserComment {
		if b != '\n' {
			return
		}
		str.stat = str.keep
	}
	switch str.stat {
	// 文字列を解析
	case ParserBeginString:
		i, err = str.parseBeginString(b)
	// 文字列解析終了
	case ParserEndString:
		i, err = str.parseEndString(b)
	// 複数行文字列を解析
	case ParserMultiBeginString:
		i, err = str.parseMultiBeginString(b)
	// 複数行文字列の解析終了
	case ParserMultiEndString:
		i, err = str.parseMultiEndString(b)
	}
	return
}

// Param 関数は、先頭、最後尾の改行を1つだけ除去した結果を返却する
func (str *String) Param() string {
	// 値を取得
	param := strings.Trim(str.Value.Param(), " ")

	if str.quote == '"' {
		// " の場合、\\n, \\t, \\" を置き換える
		param = strings.Replace(param, "\\n", "\n", -1)
		param = strings.Replace(param, "\\t", "\t", -1)
		param = strings.Replace(param, `\"`, "\"", -1)
		param = strings.Replace(param, "\\\\", "\\", -1)
	} else if str.quote == 39 {
		// ' の場合、\' を置き換える
		param = strings.Replace(param, `\'`, "'", -1)
	}
	// 先頭、最後尾の" or 'を除去する
	param = str.Trim(param, str.quote)
	// 複数行文字列の場合
	if str.stat == ParserMultiEndString {
		param = str.Trim(param, str.quote)
		param = str.Trim(param, str.quote)
		param = str.Trim(param, '\n')
	}

	return param
}

// 文字列解析開始処理
func (str *String) parseBeginString(b byte) (interface{}, error) {
	switch b {
	// 文字列の閉じ"がない場合エラーとする
	case '\n':
		return nil, fmt.Errorf("\"%s\" string invalid value", str.key)
	// 閉じ"があった場合、終了処理へ移行する
	case str.quote:
		// \" ではない場合、終了処理となる
		if str.Prev(1) != '\\' {
			str.stat = ParserEndString
			str.end = str.cnt + 1
		}
	// 上記以外の場合、許可する
	default:
		str.end = str.cnt + 1
	}
	return nil, nil
}

// 文字列解析終了処理
func (str *String) parseEndString(b byte) (interface{}, error) {
	switch b {
	// 改行があった場合、終了
	case '\n':
		// 状態を元に戻す
		str.stat = ParserNone
		// 不正な指定方法であった場合エラーとする
		// if str.end < str.pos {
		//	return nil, fmt.Errorf("\"%s\" string invalid value", str.key)
		// }
		// パラメータを取得
		return str.Param(), nil
	// 閉じ"の後に、再度"があった場合
	case str.quote:
		if str.Prev(2) == str.quote && str.Prev(1) == str.quote && b == str.quote {
			// 3つ連続で"があった場合は、複数行文字列としてみなす
			str.stat = ParserMultiBeginString
			str.end = str.cnt + 1
		} else {
			// 上記以外ではない場合、エラーとする
			return nil, fmt.Errorf("\"%s\" string invalid value", str.key)
		}
	// コメント行
	case '#':
		str.stat = ParserComment
		str.keep = ParserEndString
	// 空白はスルーする
	case ' ':
	// 上記以外
	default:
		// 配列の場合、, または]があった時点で終了とする
		if str.array {
			if b == ',' || b == ']' {
				str.end = str.cnt
				return str.Analyze('\n')
			}
		}
		// エラーとする
		return nil, fmt.Errorf("\"%s\" string invalid value", str.key)
	}
	return nil, nil
}

// 複数行文字列解析開始
func (str *String) parseMultiBeginString(b byte) (interface{}, error) {
	str.end = str.cnt + 1

	if str.prev[1] == str.quote && str.prev[0] == str.quote && b == str.quote {
		// \""" となっていた場合は、まだ終了ではない
		if str.prev[2] != '\\' {
			str.stat = ParserMultiEndString
		}
	}

	str.prev[2] = str.prev[1]
	str.prev[1] = str.prev[0]
	str.prev[0] = b
	return nil, nil
}

// 複数行文字列解析終了
func (str *String) parseMultiEndString(b byte) (interface{}, error) {
	switch b {
	// 改行があった時点で、終了とする
	case '\n':
		// 不正な指定方法であった場合エラーとする
		// if str.end < str.pos {
		//	return nil, fmt.Errorf("\"%s\" string invalid value", str.key)
		// }
		param := str.Param()
		// 状態を元に戻す
		str.stat = ParserNone
		return param, nil
	// 空白はスルー
	case ' ':
	// コメント行
	case '#':
		str.end = str.cnt
		str.stat = ParserComment
		str.keep = ParserMultiEndString
	// 上記以外
	default:
		// 配列の場合、, または]があった時点で終了とする
		if str.array {
			if b == ',' || b == ']' {
				str.end = str.cnt
				return str.Analyze('\n')
			}
		}
		// エラーとする
		return nil, fmt.Errorf("\"%s\" string invalid value", str.key)
	}
	return nil, nil
}
