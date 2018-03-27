package parser

import (
	"os"
	"strings"
	"testing"
)

// Parse 関数の正常系テスト
func TestNormalParseCase(t *testing.T) {
	// テストに使用する環境変数をセット
	os.Setenv("TESTDATA", "test")
	// 設定ファイルデータをテストとして用意する
	var strs = []string{
		"# Normal Test Data",
		"# For integer data testing",
		"integer.number1 = 0          # 0",
		"integer.number2 = 1          # 1",
		"integer.number3 = 1,000,000  # 1000",
		"integer.number4 = +1000      # 1000",
		"integer.number5 = -1000      # -1000",
		"",
		"# For float data testing",
		"float.number1 = 0.1 # 0.1",
		"float.number2 = 1.1 # 1.1",
		"",
		"# For size data testing",
		"byte.number1 = 0B     # 0",
		"byte.number2 = 1B     # 1",
		"byte.number3 = 10B    # 10",
		"byte.number4 = 1,000B",
		"",
		"kbyte.number1 = 0KB   # 0",
		"kbyte.number2 = 1KB   # 1024",
		"kbyte.number3 = 10KB  # 10240",
		"kbyte.number4 = 1,024KB",
		"",
		"mbyte.number1 = 0MB",
		"mbyte.number2 = 1MB",
		"mbyte.number3 = 10MB",
		"mbyte.number4 = 1,000MB",
		"",
		"gbyte.number1 = 0GB",
		"gbyte.number2 = 1GB",
		"gbyte.number3 = 10GB",
		"gbyte.number4 = 1,000GB",
		"",
		"tbyte.number1 = 0TB",
		"tbyte.number2 = 1TB",
		"tbyte.number3 = 10TB",
		"tbyte.number4 = 1,000TB",
		"",
		"# For time data testing",
		"ms.number1 = 0ms  # 0",
		"ms.number2 = 1ms  # 1",
		"ms.number3 = 10ms # 10",
		"ms.number4 = 1,000ms",
		"",
		"sec.number1 = 0s",
		"sec.number2 = 1s",
		"sec.number3 = 10s",
		"sec.number4 = 1,000s",
		"",
		"min.number1 = 0m",
		"min.number2 = 1m",
		"min.number3 = 10m",
		"min.number4 = 1,000m",
		"",
		"hour.number1 = 0h",
		"hour.number2 = 1h",
		"hour.number3 = 10h",
		"hour.number4 = 1,000h",
		"",
		"date.number1 = 0d",
		"date.number2 = 1d",
		"date.number3 = 10d",
		"date.number4 = 1,000d",
		"",
		"# For datetime data testing",
		"datetime = 2017-10-01 21:00:00 # 2017-10-01 21:00:00 +0000 UTC",
		"",
		"# For octal data testing",
		"oct.number1 = 0644 # 420",
		"oct.number2 = 0022",
		"",
		"# For hex data testing",
		"hex.number1 = 0xFF  # 255",
		"hex.number2 = 0x100 # 256",
		"hex.number3 = 0xff  # 255",
		"",
		"# For boolean data testing",
		"ok = true # true",
		"ng = false # false",
		"",
		"# For strings data testing",
		"strings.str = \"Hello World\" # str",
		"strings.char = 'Hello World'  # char",
		"strings.multistr = \"\"\"",
		"Hello World",
		"\"\"\" # multistr",
		"strings.multichar = '''",
		"Hello World",
		"''' # multichar",
		"",
		"# For environ data testing",
		"env = $TESTDATA # string(test)",
		"",
		"# For array data testing",
		"[arraymode]",
		"arrays.number1 = [0, 1, 0644, 0xFF, 200, 100] # array integer",
		"arrays.number2 = [100, 200] # array integer",
		"arrays.float1 = [3.1,4.2, 4.3]",
		"arrays.int64  = [100s, 1d, 100KB, 1TB]",
		"arrays.datetime = [",
		"  2017-01-01 00:00:00, 2017-02-01 01:00:00,",
		"  2017-03-01 03:00:00, 2017-04-01 04:00:00,",
		"]",
		"arrays.strings = [",
		"    \"Hello World\", # str",
		"    'Hello World',  # char",
		"    \"\"\"",
		"Hello World\"\"\", # Multistr",
		"    '''",
		"Hello World''', # Multichar",
		"]",
		"arrays.envname = [ $TESTDATA, $TESTDATA ] # test test",
		"arrays.boolean = [ true, false, true, false ]",
		"arrays.inner = [ [1,2,3], [4,5,6] ]",
	}
	// テストで生成した設定ファイル情報をパースする
	p, err := Parse([]byte(strings.Join(strs, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	if p.Data() == nil {
		t.Fatal("data is nil")
	}
	// ダミー関数のコール。意味はない。
	p.Analyze(0)
	if p.Prev(10000) != 0 {
		t.Fatal("Prev function error")
	}
}

// ParseModeAll 関数の正常系テスト
func TestNormalParseModeAllCase(t *testing.T) {
	var strs = []string{
		"[modename1]",
		"data = 100",
		"",
		"[modename2]",
		"data = 200",
	}
	// テストで生成した設定ファイル情報をパースする
	p, err := ParseModeAll([]byte(strings.Join(strs, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	if p.Data() == nil {
		t.Fatal("data is nil")
	}
}

// ParseModeAll 関数の異常系
func TestErrorParseModeAllCase(t *testing.T) {
	var strs = []string{
		"data = 100",
		"",
		"[modename2]",
		"data = 200",
	}
	// テストで生成した設定ファイル情報をパースする
	_, err := ParseModeAll([]byte(strings.Join(strs, "\n")))
	if err == nil {
		t.Fatal("modename is nil")
	}
}

// モード名の異常系テスト
func TestErrorModenameCase(t *testing.T) {
	if _, err := Parse([]byte("[]")); err == nil {
		t.Error("modename is nil")
	}
	if _, err := Parse([]byte("[")); err == nil {
		t.Error("modename is close ]")
	}
	if _, err := Parse([]byte("[_modename]")); err == nil {
		t.Error("modename is first char _ not use")
	}
	if _, err := Parse([]byte("[mode name]")); err == nil {
		t.Error("modename is first char _ not use")
	}
	if _, err := Parse([]byte("[modename]\na = 100\n[modename]")); err == nil {
		t.Error("duplicate modename error")
	}
}

// キー名の異常系テスト
func TestErrorKeynameCase(t *testing.T) {
	if _, err := Parse([]byte("%keyname")); err == nil {
		t.Error("keyname test failed")
	}
	if _, err := Parse([]byte(":keyname")); err == nil {
		t.Error("keyname test failed")
	}
	if _, err := Parse([]byte("keyname")); err == nil {
		t.Error("keyname test failed")
	}
	if _, err := Parse([]byte("keyname =")); err == nil {
		t.Error("keyname test failed")
	}
	if _, err := Parse([]byte(".keyname =")); err == nil {
		t.Error("keyname test failed")
	}
	if _, err := Parse([]byte("keyname. =")); err == nil {
		t.Error("keyname test failed")
	}
	if _, err := Parse([]byte("0keyname =")); err == nil {
		t.Error("keyname test failed")
	}
	if _, err := Parse([]byte("keyname.value& =")); err == nil {
		t.Error("keyname test failed")
	}
}

// 重複チェックの異常系テスト
func TestErrorDuplicateCase(t *testing.T) {
	var strs = []string{
		"value = 100",
		"value = 200", // ここでエラーになる
	}
	if _, err := Parse([]byte(strings.Join(strs, "\n"))); err == nil {
		t.Error("duplicate test failed")
	}
	strs = []string{
		"value    = false",
		"value.ok = true", // ここでエラーになる
	}
	if _, err := Parse([]byte(strings.Join(strs, "\n"))); err == nil {
		t.Error("duplicate test failed")
	}
}

// 数字系の異常系テスト
func TestErrorIntegerCase(t *testing.T) {
	// オーバーフロー
	if _, err := Parse([]byte("integer = 2147483648")); err == nil {
		t.Error("overflow test failed")
	}
	if _, err := Parse([]byte("integer = 2,147,483,648")); err == nil {
		t.Error("overflow test failed")
	}
	if _, err := Parse([]byte("integer = 0x100000000")); err == nil {
		t.Error("overflow test failed")
	}
	if _, err := Parse([]byte("integer = 070000000000000")); err == nil {
		t.Error("overflow test failed")
	}
	if _, err := Parse([]byte("integer = 1000000000000000000000000000000000000000d")); err == nil {
		t.Error("overflow test failed")
	}
	if _, err := Parse([]byte("integer = 1000000000000000000000000000000000000000TB")); err == nil {
		t.Error("overflow test failed")
	}
	if _, err := Parse([]byte("float = 1000000000000000000000000000000000000000.2987654321")); err == nil {
		t.Error("overflow test failed")
	}
	// 数字以外の入力はエラー
	if _, err := Parse([]byte("integer = 0a")); err == nil {
		t.Error("int number test failed")
	}
	if _, err := Parse([]byte("integer = 100a")); err == nil {
		t.Error("int number test failed")
	}
	if _, err := Parse([]byte("integer = 1,000a")); err == nil {
		t.Error("int number test failed")
	}
	if _, err := Parse([]byte("integer = 100ss")); err == nil {
		t.Error("int number test failed")
	}
	if _, err := Parse([]byte("float = 3.141592z")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 0xFFZ")); err == nil {
		t.Error("datetime test failed")
	}
	// 8進数での、 8,9 文字は使用不可能
	if _, err := Parse([]byte("integer = 0699")); err == nil {
		t.Error("oct number test failed")
	}
	// 8進数での、 8,9 文字は使用不可能
	if _, err := Parse([]byte("integer = 0999")); err == nil {
		t.Error("oct number test failed")
	}
	// 符号付、8進数表記は不可能
	if _, err := Parse([]byte("integer = +0")); err == nil {
		t.Error("oct number test failed")
	}
	// 符号付、日付の指定は不可能
	if _, err := Parse([]byte("datetime = +2018-12-01 00:01:02")); err == nil {
		t.Error("datetime test failed")
	}
	// 指定した日付以外の指定は不可能
	if _, err := Parse([]byte("datetime = 2018-12-01_00:01:02")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("datetime = 2018-12-01  00:01:02")); err == nil {
		t.Error("datetime test failed")
	}
	// 不正な時刻の指定は不可能
	if _, err := Parse([]byte("datetime = 2018-13-01 26:01:02")); err == nil {
		t.Error("datetime test failed")
	}
	// 空白ありの数字指定は不可能
	if _, err := Parse([]byte("integer = 1,00 100")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 100 100")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 100 s")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 100 KB")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 0x FF")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 06 44")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("float = 3. 141592")); err == nil {
		t.Error("datetime test failed")
	}
	// カンマが、先頭、最後尾に付与されていた場合はエラーとする
	if _, err := Parse([]byte("integer = 100KB,")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 100,KB")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 100,s")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 100s,")); err == nil {
		t.Error("datetime test failed")
	}
	if _, err := Parse([]byte("integer = 1,00,")); err == nil {
		t.Error("datetime test failed")
	}
}

// 真偽値の異常系テスト
func TestErrorBooleanCase(t *testing.T) {
	if _, err := Parse([]byte("boolean = tru")); err == nil {
		t.Error("boolean = tru")
	}
	if _, err := Parse([]byte("boolean = ok")); err == nil {
		t.Error("boolean = ok")
	}
}

// 文字列型の異常系テスト
func TestErrorStringsCase(t *testing.T) {
	if _, err := Parse([]byte("str = \"OK")); err == nil {
		t.Error("str = ng")
	}
	if _, err := Parse([]byte(`str = """Hello World""'`)); err == nil {
		t.Error("str = ng")
	}
	if _, err := Parse([]byte(`str = """Hello World\"""`)); err == nil {
		t.Error("str = ng")
	}
	if _, err := Parse([]byte(`str = "Hello World""`)); err == nil {
		t.Error("str = ng")
	}
	if _, err := Parse([]byte(`str = """Hello World"""NG`)); err == nil {
		t.Error("str = ng")
	}
	if _, err := Parse([]byte(`str = "Hello World"NG`)); err == nil {
		t.Error("str = ng")
	}
}

// 環境変数の異常系テスト
func TestErrorEnvironCase(t *testing.T) {
	if _, err := Parse([]byte("env = $")); err == nil {
		t.Error("env = ng")
	}
	if _, err := Parse([]byte("env = $TEST MODE")); err == nil {
		t.Error("env = ng")
	}
}

// 配列の異常系テスト
func TestErrorArrayCase(t *testing.T) {
	if _, err := Parse([]byte("array = [,1]")); err == nil {
		t.Error("array = ng")
	}
	if _, err := Parse([]byte(`array = ["Hello World", 200]`)); err == nil {
		t.Error("array = ng")
	}
	if _, err := Parse([]byte("array = [\n'Hello World'\n'Hello World' ]")); err == nil {
		t.Error("array = ng")
	}
	if _, err := Parse([]byte("array = [Hello World]")); err == nil {
		t.Error("array = ng")
	}
	if _, err := Parse([]byte("array = [2147483648]")); err == nil {
		t.Error("array = ng")
	}
}
