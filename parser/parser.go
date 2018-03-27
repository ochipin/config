package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser.stat, Parser.keep に設定する解析状態
const (
	ParserNone             = iota + 1 // 未解析状態
	ParserComment                     // コメント行を処理中
	ParserKeyname                     // key = value のkey解析時の状態
	ParserValue                       // key = value のvalue解析時の状態
	ParserModename                    // モード名を解析
	ParserNumberAny                   // 小数点、または8/16進数のいずれか
	ParserNumber                      // 小数点、または整数
	ParserNumberInt                   // 整数
	ParserNumberFloat                 // 小数点
	ParserNumberOct                   // 8進数
	ParserNumberHex                   // 16進数
	ParserNumberDate                  // 日付
	ParserNumberTime                  // 1ms/1s/1m/1h/1d の時間表記解析
	ParserNumberSize                  // KB/MB/GB/TB のサイズ表記解析
	ParserBeginString                 // 文字列開始
	ParserEndString                   // 文字列終了
	ParserMultiBeginString            // 複数行文字列の解析開始
	ParserMultiEndString              // 複数行文字列の解析終了
	ParserBeginArray                  // 配列解析
)

// Parser 構造体は、設定ファイルを解析する
type Parser struct {
	Value             // Value 構造体をミックスイン
	data  interface{} // 保持するデータ
	line  int         // 行番号
	mode  string      // モード名
	node  Node        // 値解析用ノード
}

// Analyze 関数は、ダミー。解析時に使用する関数の引数に渡すためだけに実装している。
func (p *Parser) Analyze(b byte) (interface{}, error) {
	return nil, nil
}

// ポジションをクリアする
func (p *Parser) clear() {
	p.pos = 0
	p.end = 0
	p.node = nil
}

// 設定ファイルパース関数
func (p *Parser) parse(b byte) (err error) {
	// 値を解析する場合は、値解析用ノードを使用する
	if p.node != nil {
		// 現在の参照ポイントを設定
		p.node.Cnt(p.cnt)
		// 解析関数をコール
		data, err := p.node.Analyze(b)
		if err != nil {
			return err
		}
		// 解析終了の場合、値をセットする
		if p.node.Stat() == ParserNone {
			p.stat = ParserNone
			if err = Set(p.key, data, p.data, p.mode); err != nil {
				return err
			}
			p.clear()
		}
		// 改行コードの場合、行番号をカウント
		if b == '\n' {
			p.line++
		}
		return nil
	}

	switch p.stat {
	// 未解析状態
	case ParserNone:
		err = p.none(b)
	// コメント行を処理している状態
	case ParserComment:
		if b == '\n' {
			p.line++
			p.stat = ParserNone
		}
	// キー名を解析
	case ParserKeyname:
		err = p.keyname(b)
	// 右辺値を解析
	case ParserValue:
		err = p.value(b)
	case ParserModename:
		err = p.modename(b)
	}

	return err
}

// 設定ファイル未解析状態の場合にコールされる
func (p *Parser) none(b byte) (err error) {
	switch b {
	// コメント行の場合
	case '#':
		p.stat = ParserComment
	// 改行コードの場合
	case '\n':
		p.line++
	// [modename]の始まりを解析する状態
	case '[':
		p.stat = ParserModename
		p.pos = p.cnt
		p.end = 0
	// 未解析状態時では、使用できない特殊文字
	case '?', '!', '@', '$', '%', '^', '&', '*', '(', ')', '+', '|', '\\', ']':
		return fmt.Errorf("key name specified is not special character")
	// 未解析状態時では、使用できない特殊文字
	case '`', '"', '-', '{', '}', ':', ';', '<', '>', '/', ',', '~', 39, '=':
		return fmt.Errorf("key name specified is not special character")
	case ' ':
	// key = value の key を解析する場合
	default:
		if p.mode == "" {
			return fmt.Errorf("mode name is empty")
		}
		p.stat = ParserKeyname
		p.pos = p.cnt
	}
	return err
}

// モード名を取得する
func (p *Parser) modename(b byte) (err error) {
	// 値がセットされていないが、改行コードがあった場合、エラーとする
	if b == '\n' && (p.end <= p.pos) {
		return fmt.Errorf("no set modename")
	}
	// 終了タグではない場合、何もせず復帰する
	if b != ']' {
		return nil
	}

	// モード名を取得
	p.end = p.cnt
	p.mode = p.Param()
	p.clear()

	// 先頭、最後尾の [] を外す
	if p.mode != "" && p.mode[0] == '[' {
		p.mode = strings.Trim(p.mode[1:], " ")
	}
	// モード名が空文字列の場合、エラーとする
	if p.mode == "" {
		return fmt.Errorf("mode name is empty")
	}
	// モード名の値を検証する
	if strings.Index(p.mode, " ") != -1 {
		return fmt.Errorf("\"%s\" mode name is invalid", p.mode)
	}
	if p.mode[0] == '_' {
		return fmt.Errorf("\"%s\" can not specify '_' first character", p.mode)
	}
	// 既に使用済みのモード名の場合、エラーとする
	if p.data != nil {
		if v, ok := p.data.(map[string]interface{}); ok {
			if _, ok := v[p.mode]; ok {
				return fmt.Errorf("\"%s\" mode is already exists", p.mode)
			}
		}
	}
	p.stat = ParserNone
	return nil
}

// 指定されたキー名を解析する
func (p *Parser) keyname(b byte) (err error) {
	p.end = p.cnt
	// 改行コードの場合、エラーとする
	if b == '\n' {
		return fmt.Errorf("invalid configuration")
	}
	// = が出現するまで左辺値として扱う
	if b != '=' {
		return
	}
	// = が出現した時点で、次回から右辺値を解析できるように準備する
	p.stat = ParserValue
	// 指定されたキーが、正しいかチェックする
	p.key = strings.ToLower(strings.Trim(p.Param(), " "))
	for _, v := range p.key {
		if (v >= 'A' && v <= 'Z') || (v >= 'a' && v <= 'z') || (v >= '0' && v <= '9') || v == '.' || v == '_' {
		} else {
			return fmt.Errorf("\"%s\" key name is invalid", p.key)
		}
	}
	// 先頭、最後尾に "." があった場合は、不正なキー名として扱う
	if p.key != "" && p.key[0] == '.' || p.key[len(p.key)-1] == '.' || p.key[0] == '_' || p.key[len(p.key)-1] == '_' {
		return fmt.Errorf("\"%s\" key name is invalid", p.key)
	}
	// "."区切りのキー名の先頭が、数字の場合は不正なキー名として扱う
	for _, keyname := range strings.Split(p.key, ".") {
		if keyname[0] >= '0' && keyname[0] <= '9' || keyname[0] == '_' || keyname[len(keyname)-1] == '_' {
			return fmt.Errorf("\"%s\" key name is invalid", p.key)
		}
	}
	// 開始、終了の範囲をクリアする
	p.clear()
	return
}

// 右辺値(value)を解析する
func (p *Parser) value(b byte) error {
	switch b {
	// 小数点、日付、10/8/16進数のいずれかの場合
	case '0', '+', '-', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		p.pos = p.cnt
		p.end = p.cnt + 1
		p.node = NewNumber(p)
	// 真偽値
	case 't', 'f':
		p.pos = p.cnt
		p.end = p.cnt + 1
		p.node = NewBoolean(p)
	// 文字列
	case '"', 39:
		p.pos = p.cnt
		p.end = p.cnt + 1
		p.node = NewString(p)
	// 環境変数
	case '$':
		p.pos = p.cnt
		p.end = p.cnt + 1
		p.node = NewEnviron(p)
	// 配列
	case '[':
		// p.stat = ParserBeginArray
		p.pos = p.cnt + 1
		p.node = NewArray(p)
	// 空白は無視
	case ' ':
	default:
		return fmt.Errorf("\"%s\" invalid value", p.key)
	}
	return nil
}

// 整形した map[string]interface{} 型を返却する
func (p *Parser) Data() interface{} {
	return p.data
}

// Parse は、設定ファイル情報から map[string]interface{} 情報を構築する
func parse(buf []byte, mode string) (*Parser, error) {
	// CR+LF, CR 対策
	s := strings.Replace(string(buf), "\r\n", "\n", -1)
	s = strings.Replace(s, "\r", "\n", -1) + "\n"
	// パース構造体を生成
	var parser = &Parser{
		Value: Value{
			text: []byte(s),
			stat: ParserNone,
		},
		data: make(map[string]interface{}),
		mode: mode,
	}

	// パース処理開始
	for i := 0; i < len(parser.text); i++ {
		c := parser.text[i]
		// パースする
		parser.cnt = i
		if err := parser.parse(c); err != nil {
			if e, ok := err.(*strconv.NumError); ok {
				return nil, fmt.Errorf("parsing error:%d: \"%s\" setting value is \"%s\" %s", parser.line+1, parser.key, e.Num, e.Err)
			}
			return nil, fmt.Errorf("syntax error:%d: %s", parser.line+1, err)
		}
	}

	// 正しく解析終了したかチェックする
	if parser.stat != ParserNone {
		return nil, fmt.Errorf("syntax error: invalid configuration. probably cause \"%s\" parameters", parser.key)
	}
	parser.mode = ""

	return parser, nil
}

// 冒頭にモード指定がされていなくとも設定ファイルを解析する
func Parse(buf []byte) (*Parser, error) {
	return parse(buf, "_all_")
}

// 冒頭にモード指定がされていないと設定ファイルを解析しない
func ParseModeAll(buf []byte) (*Parser, error) {
	return parse(buf, "")
}
