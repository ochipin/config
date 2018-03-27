package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 日付チェック用正規表現
var regexpDate = regexp.MustCompile(`^\d{4}\-\d{2}\-\d{2}\s\d{2}:\d{2}:\d{2}$`)

// Number 構造体は、数字の並びを解析し、整数、小数点、日付等を返却する
type Number struct {
	Value       // Value構造体をミックスイン
	sign   bool // +/-の符号付きの場合 true
	signok bool // +/-の符号の次に付与される数字を判定するために使用する
	keep   int  // 解析状態の維持
	array  bool // 配列内の値として使用される場合 true
}

// NewNumber 関数は、数字解析ノードを生成する
func NewNumber(p Node) Node {
	var text = p.Text()
	var cnt = p.Getidx()
	// +/- 符号が付いている場合 true とする
	var sign = text[cnt] == '+' || text[cnt] == '-'
	// 先頭が 0 から始まらない場合は、整数か、小数点、もしくは日付として判定処理を開始する
	var stat = ParserNumber
	// もし先頭が 0 から始まっていた場合は、8/16進数の可能性を考慮するステータスに変更する
	if text[cnt] == '0' {
		stat = ParserNumberAny
	}
	_, ok := p.(*Array)
	return &Number{
		Value: Value{
			text: text,
			stat: stat,
			cnt:  cnt,
			pos:  p.Pos(),
			end:  p.End(),
			key:  p.Keyname(),
		},
		array:  ok,
		sign:   sign,
		signok: sign,
	}
}

// Analyze 関数は、真偽値解析を実施する
func (number *Number) Analyze(b byte) (i interface{}, err error) {
	// コメント行の場合
	if number.stat == ParserComment {
		if b != '\n' {
			return
		}
		number.stat = number.keep
	}
	switch number.stat {
	// 小数点、8進数、16進数のいずれかの場合
	case ParserNumberAny:
		i, err = number.parseNumberAny(b)
	// 小数点、整数、符号付整数、日付のいずれかの場合
	case ParserNumber:
		i, err = number.parseNumber(b)
	// 整数の場合
	case ParserNumberInt:
		i, err = number.parseNumberInt(b)
	// 小数点の場合
	case ParserNumberFloat:
		i, err = number.parseNumberFloat(b)
	// 8進数の場合
	case ParserNumberOct:
		i, err = number.parseNumberOct(b)
	// 16進数の場合
	case ParserNumberHex:
		i, err = number.parseNumberHex(b)
	// 日付の場合
	case ParserNumberDate:
		i, err = number.parseNumberDate(b)
	// 時間指定の場合
	case ParserNumberTime:
		i, err = number.parseNumberTime(b)
	// サイズ指定の場合
	case ParserNumberSize:
		i, err = number.parseNumberSize(b)
	}
	return i, err
}

// 小数点、8進数、16進数、いずれか不明な場合にコールされる
func (number *Number) parseNumberAny(b byte) (interface{}, error) {
	switch b {
	// 8進数
	case '0', '1', '2', '3', '4', '5', '6', '7':
		number.stat = ParserNumberOct
		number.end = number.cnt + 1
	// 16進数
	case 'x':
		number.stat = ParserNumberHex
		number.end = number.cnt + 1
	// 小数点
	case '.':
		number.stat = ParserNumberFloat
		number.end = number.cnt + 1
	// 時間指定の場合
	case 'm', 's', 'h', 'd':
		number.stat = ParserNumberTime
		number.end = number.cnt + 1
	// サイズ指定の場合
	case 'B', 'K', 'M', 'G', 'T':
		number.stat = ParserNumberSize
		number.end = number.cnt + 1
	// 改行コードがあった場合は、 "0" 以外ありえないので0をセット
	case '\n':
		number.stat = ParserNone
		return 0, nil
	// 空白はスルーする
	case ' ':
	// コメント行
	case '#':
		number.keep = ParserNumberAny
		number.stat = ParserComment
	// 8, 9 が指定された場合は、8進数エラーとして扱う
	case '8', '9':
		return nil, fmt.Errorf("\"%s\" oct invalid value", number.key)
	// 上記以外
	default:
		// 配列の場合、, または]があった時点で終了とする
		if number.array {
			if b == ',' || b == ']' {
				number.stat = ParserNone
				return 0, nil
			}
		}
		// 不正とみなし、エラーを返却する
		return nil, fmt.Errorf("\"%s\" invalid value", number.key)
	}
	return nil, nil
}

// 小数点、整数、いずれか不明な場合にコールされる
func (number *Number) parseNumber(b byte) (interface{}, error) {
	switch b {
	// 小数点の場合
	case '.':
		number.stat = ParserNumberFloat
		number.end = number.cnt + 1
	// 整数の場合
	case ',':
		// 配列内の整数の場合、, があった時点で終了とする
		if number.array {
			number.end = number.cnt
			return number.parseNumber('\n')
		}
		number.stat = ParserNumberInt
		number.end = number.cnt + 1
	// 日付指定の場合
	case '-':
		number.end = number.cnt + 1
		param := number.Param()
		// 符号付きの場合、日付指定はできない
		if number.sign {
			return nil, fmt.Errorf("\"%s = %s\" datetime invalid value", number.key, param)
		}
		number.stat = ParserNumberDate
	// '0' を指定されている場合 ex) +0
	case '0':
		number.end = number.cnt + 1
		// 符号付き 0 の場合、不正とみなしエラーとする
		if number.signok {
			param := number.Param()
			return nil, fmt.Errorf("\"%s = %s\" oct invalid value", number.key, param)
		}
	// 0-9 の場合
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		number.end = number.cnt + 1
	// 時間指定の場合
	case 'm', 's', 'h', 'd':
		number.stat = ParserNumberTime
		number.end = number.cnt + 1
	// サイズ指定の場合
	case 'B', 'K', 'M', 'G', 'T':
		number.stat = ParserNumberSize
		number.end = number.cnt + 1
	// 改行コードがあった場合終了とする
	case '\n':
		// 状態を元に戻す
		number.stat = ParserNone
		// 終了ポイントが未設定の場合は設定する
		// if number.end < number.pos {
		//	number.end = number.cnt - 1
		// }
		// 値を取得
		param := strings.Trim(number.Param(), " ")
		// 不正な文字列がないかチェック
		if strings.Index(param, " ") != -1 {
			return nil, fmt.Errorf("\"%s = %s\" integer invalid value", number.key, param)
		}
		// 取得した値を整数へ変換する
		result, err := strconv.ParseInt(param, 10, 32)
		if err != nil {
			return nil, err
		}
		return int(result), nil
	// 空白はスルーする
	case ' ':
	// コメント行
	case '#':
		number.end = number.cnt
		number.keep = ParserNumber
		number.stat = ParserComment
	// 上記以外
	default:
		// 配列内の整数の場合、, があった時点で終了とする
		if number.array {
			if b == ']' {
				number.end = number.cnt
				return number.parseNumber('\n')
			}
		}
		// エラーを返却する
		number.end = number.cnt + 1
		return nil, fmt.Errorf("\"%s = %s\" integer invalid value", number.key, number.Param())
	}
	number.signok = false
	return nil, nil
}

// 整数(10進数)を処理する場合にコールされる
func (number *Number) parseNumberInt(b byte) (interface{}, error) {
	switch b {
	// 整数は許可する
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		number.end = number.cnt + 1
	// 整数は、カンマ区切りでの指定を認める
	case ',':
		number.end = number.cnt + 1
	// 時間指定の場合
	case 'm', 's', 'h', 'd':
		number.stat = ParserNumberTime
		number.end = number.cnt + 1
	// サイズ指定の場合
	case 'B', 'K', 'M', 'G', 'T':
		number.stat = ParserNumberSize
		number.end = number.cnt + 1
	// 改行コードがあった時点で、終了とする
	case '\n':
		// 状態を元に戻す
		number.stat = ParserNone
		// データの取得
		param := strings.Trim(number.Param(), " ")
		// データ内に空白が紛れ込んでいた場合は、エラーとする
		if strings.Index(param, " ") != -1 {
			return nil, fmt.Errorf("\"%s = %s\" integer invalid value", number.key, param)
		}
		// 先頭、最後尾に,が発見された場合は、エラーとする
		if param != "" && (param[0] == ',' || param[len(param)-1] == ',') {
			return nil, fmt.Errorf("\"%s = %s\" integer invalid value", number.key, param)
		}
		// 数字文字列内に存在する,を除去する
		param = strings.Replace(param, ",", "", -1)
		// 整数へ変換する
		result, err := strconv.ParseInt(param, 10, 32)
		if err != nil {
			return nil, err
		}
		return int(result), nil
	// 空白はスルーする
	case ' ':
	// コメント行
	case '#':
		number.keep = ParserNumberInt
		number.stat = ParserComment
	// 上記以外はエラーとする
	default:
		// エラーとする
		number.end = number.cnt + 1
		return nil, fmt.Errorf("\"%s = %s\" integer invalid value", number.key, number.Param())
	}
	return nil, nil
}

// 小数点を処理する場合にコールされる
func (number *Number) parseNumberFloat(b byte) (interface{}, error) {
	switch b {
	// 0-9 の数字は許可する
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		number.end = number.cnt + 1
	// 改行コードがあった時点で、終了とする
	case '\n':
		// 状態を元に戻す
		number.stat = ParserNone
		// 値の取得
		param := strings.Trim(number.Param(), " ")
		// 値のチェック
		if strings.Index(param, " ") != -1 || param[len(param)-1] == '.' {
			return nil, fmt.Errorf("\"%s = %s\" float invalid value", number.key, param)
		}
		result, err := strconv.ParseFloat(param, 32)
		if err != nil {
			return nil, err
		}
		return float32(result), nil
	// 空行はスルーする
	case ' ':
	// コメント行は無視するよう、状態を変更する
	case '#':
		number.keep = ParserNumberFloat
		number.stat = ParserComment
	// 上記以外は、小数点には使用できない文字列とする
	default:
		// 配列内の整数の場合、, があった時点で終了とする
		if number.array {
			if b == ',' || b == ']' {
				number.end = number.cnt
				return number.parseNumberFloat('\n')
			}
		}
		number.end = number.cnt + 1
		return nil, fmt.Errorf("\"%s = %s\" float invalid value", number.key, number.Param())
	}
	return nil, nil
}

// 8進数を処理する場合にコールされる
func (number *Number) parseNumberOct(b byte) (interface{}, error) {
	switch b {
	// 0-7までの数字の場合は許可する
	case '0', '1', '2', '3', '4', '5', '6', '7':
		number.end = number.cnt + 1
	// 改行コードがあった場合は終了とする
	case '\n':
		// 状態を元に戻す
		number.stat = ParserNone
		// 値を取得
		param := strings.Trim(number.Param(), " ")
		// 値を検証
		if strings.Index(param, " ") != -1 {
			return nil, fmt.Errorf("\"%s = %s\" oct invalid value", number.key, param)
		}
		result, err := strconv.ParseInt(param, 8, 33)
		if err != nil {
			return nil, err
		}
		return int(result), nil
	// 空行はスルーする
	case ' ':
	// コメント行
	case '#':
		number.keep = ParserNumberOct
		number.stat = ParserComment
	// 上記以外はエラーとする
	default:
		// 配列内の整数の場合、, があった時点で終了とする
		if number.array {
			if b == ',' || b == ']' {
				number.end = number.cnt
				return number.parseNumberOct('\n')
			}
		}
		// エラーとする
		number.end = number.cnt + 1
		return nil, fmt.Errorf("\"%s = %s...\" oct invalid value", number.key, number.Param())
	}

	return nil, nil
}

// 16進数を処理する場合にコールされる
func (number *Number) parseNumberHex(b byte) (interface{}, error) {
	switch b {
	// 0-9a-fA-F 文字のみ許可する
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		number.end = number.cnt + 1
	case 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F':
		number.end = number.cnt + 1
	// 改行コードがあった場合終了とする
	case '\n':
		// 状態を元に戻す
		number.stat = ParserNone
		// 値を取得
		param := strings.Trim(number.Param(), " ")
		// 不正な文字列でないかチェック
		if strings.Index(param, " ") != -1 || len(param) == 2 {
			return nil, fmt.Errorf("\"%s = %s\" hex invalid value", number.key, param)
		}
		result, err := strconv.ParseInt(param[2:], 16, 33)
		if err != nil {
			return nil, err
		}
		return int(result), nil
	// 空行はスルーする
	case ' ':
	// コメント行
	case '#':
		number.keep = ParserNumberHex
		number.stat = ParserComment
	// 上記以外はエラーとする
	default:
		// 配列内の整数の場合、, があった時点で終了とする
		if number.array {
			if b == ',' || b == ']' {
				number.end = number.cnt
				return number.parseNumberHex('\n')
			}
		}
		number.end = number.cnt + 1
		return nil, fmt.Errorf("\"%s = %s...\" hex invalid value", number.key, number.Param())
	}
	return nil, nil
}

// 日付を解析する
func (number *Number) parseNumberDate(b byte) (interface{}, error) {
	switch b {
	case '\n':
		// 状態を元に戻す
		number.stat = ParserNone
		// 取得したパラメータが正しいか検証
		param := strings.Trim(number.Param(), " ")
		if !regexpDate.MatchString(param) {
			return nil, fmt.Errorf("\"%s = %s\" datetime invalid value", number.key, param)
		}
		// 日付型へ変換する
		result, err := time.Parse("2006-01-02 15:04:05", param)
		// 変換失敗の場合はエラーを返却
		if err != nil {
			return nil, fmt.Errorf("\"%s = %s\" datetime invalid value", number.key, param)
		}
		return result, nil
	// 0-9, -, : 文字は日付文字として許可
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', ':':
		number.end = number.cnt + 1
	// 空白はスルーする
	case ' ':
	// コメント行とする
	case '#':
		number.end = number.cnt
		number.stat = ParserComment
		number.keep = ParserNumberDate
	// 上記以外
	default:
		// 配列内の値の場合、, or ] があった時点で終了とする
		if number.array {
			if b == ',' || b == ']' {
				number.end = number.cnt
				return number.parseNumberDate('\n')
			}
		}
		// 不正とみなしエラーとする
		number.end = number.cnt + 1
		return nil, fmt.Errorf("\"%s = %s...\" datetime invalid error", number.key, number.Param())
	}
	return nil, nil
}

// 1ms, 1s, 1m, 1h, 1d などの時間指定の場合
func (number *Number) parseNumberTime(b byte) (interface{}, error) {
	switch b {
	case 's':
		if number.Prev(1) != 'm' {
			return nil, fmt.Errorf("\"%s = %s...\" time invalid error", number.key, number.Param())
		}
		number.end = number.cnt + 1
	case '\n':
		// 状態を元に戻す
		number.stat = ParserNone
		// 取得したパラメータが正しいか検証
		param := strings.Trim(number.Param(), " ")
		// 不正な文字列でないかチェック
		if strings.Index(param, " ") != -1 || len(param) < 2 {
			return nil, fmt.Errorf("\"%s = %s\" time invalid value", number.key, param)
		}
		// 単位を求める
		var unit int64
		var numstr = param[:len(param)-1]
		switch param[len(param)-1] {
		case 's':
			unit = 1000
			if param[len(param)-2] == 'm' {
				unit = 1
				numstr = numstr[:len(numstr)-1]
			}
		case 'm':
			unit = 1000 * 60
		case 'h':
			unit = 1000 * 60 * 60
		case 'd':
			unit = 86400 * 1000
		}
		// 先頭、最後尾に,が発見された場合は、エラーとする
		if numstr != "" && (numstr[0] == ',' || numstr[len(numstr)-1] == ',') {
			return nil, fmt.Errorf("\"%s = %s\" size invalid value", number.key, numstr)
		}
		// 数字文字列内に存在する,を除去する
		numstr = strings.Replace(numstr, ",", "", -1)
		// 整数変換開始
		num, err := strconv.ParseInt(numstr, 10, 64)
		if err != nil {
			return nil, err
		}
		return num * unit, nil
	case ' ':
	case '#':
		number.end = number.cnt
		number.keep = number.stat
		number.stat = ParserComment
	default:
		// 配列内の値の場合、, or ] があった時点で終了とする
		if number.array {
			if b == ',' || b == ']' {
				number.end = number.cnt
				return number.parseNumberTime('\n')
			}
		}
		// 不正とみなしエラーとする
		number.end = number.cnt + 1
		return nil, fmt.Errorf("\"%s = %s...\" time invalid error", number.key, number.Param())
	}
	return nil, nil
}

// 1KB, 1MB, 1GB, 1TB などのサイズ指定の場合
func (number *Number) parseNumberSize(b byte) (interface{}, error) {
	switch b {
	case 'B':
		number.end = number.cnt + 1
	case '\n':
		// 状態を元に戻す
		number.stat = ParserNone
		// 取得したパラメータが正しいか検証
		param := strings.Trim(number.Param(), " ")
		// 不正な文字列でないかチェック
		if strings.Index(param, " ") != -1 || len(param) < 2 {
			return nil, fmt.Errorf("\"%s = %s\" size invalid value", number.key, param)
		}
		// 単位を求める
		var unit int64 = 1
		var idx int = 2
		switch param[len(param)-idx:] {
		case "KB":
			unit = 1024
		case "MB":
			unit = 1024 * 1024
		case "GB":
			unit = 1024 * 1024 * 1024
		case "TB":
			unit = 1024 * 1024 * 1024 * 1024
		default:
			// Bのみの場合、Byte として扱う
			if param[len(param)-1] == 'B' {
				idx = 1
			}
		}
		// ユニットを削除
		param = param[:len(param)-idx]
		// 先頭、最後尾に,が発見された場合は、エラーとする
		if param != "" && (param[0] == ',' || param[len(param)-1] == ',') {
			return nil, fmt.Errorf("\"%s = %s\" size invalid value", number.key, param)
		}
		// 数字文字列内に存在する,を除去する
		param = strings.Replace(param, ",", "", -1)
		// 整数変換開始
		num, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return nil, err
		}
		return num * unit, nil
	case ' ':
	case '#':
		number.end = number.cnt
		number.keep = number.stat
		number.stat = ParserComment
	default:
		// 配列内の値の場合、, or ] があった時点で終了とする
		if number.array {
			if b == ',' || b == ']' {
				number.end = number.cnt
				return number.parseNumberSize('\n')
			}
		}
		// 不正とみなしエラーとする
		number.end = number.cnt + 1
		return nil, fmt.Errorf("\"%s = %s...\" size invalid error", number.key, number.Param())
	}
	return nil, nil
}
