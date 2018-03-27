設定ファイルパーサ
===========================================================================
ini形式の設定ファイルをパースするライブラリです。  
次のような設定ファイルを、 `map[string]interface{}`, または `struct` へ変換します。

```conf
appname = "SampleTest"

timeout.min = 10ms
timeout.max = 10s

detail  = """
Sample
Configuration
"""

app.enable  = true
app.release = 2018-03-10 14:32:11
app.version = '1.0'

app.servers = [
    "example.com",
    "example.ne.jp",
    "example.net"
]
app.ports = [
    [8080, 8081, 8082],
    [9001, 9002, 9003]
]

# executing mode
[development]
appname = "SampleTest-Dev"
```

```go
package main

import (
    "time"

    "github.com/ochipin/config"
)

type Config struct {
    Appname string   // SampleTest-Dev
    Timeout struct {
        Min int64    // 10
        Max int64    // 10000
    }
    Detail string    // Sample\nConfiguration
    App    App
}

type App struct {
    Enable  bool      // true
    Release time.Time // 2018-03-10 14:32:11 +0000 UTC
    Version string    // 1.0
    Servers []string  // [example.com example.ne.jp example.net]
    Ports   [][]int   // [[8080 8081 8082] [9001 9002 9003]]
}

func main() {
    var conf Config
    if err := config.Parse("path/to/config.conf", "development", &conf); err != nil {
        panic(err)
    }
    fmt.Println(conf)
}
```

設定ファイルに記載する各種パラメータは、 `"パラメータ名 = 値"` 形式で記載します。

```
# パラメータ名  =  値
# --------------------
  paramname    =  200
```

## インストール方法
```
$ go get github.com/ochipin/config # ライブラリのインストール
$ go get github.com/ochipin/config/cmd/cfgtool # cfgtool のインストール (後述)
```

## 設定ファイル内のコメント
設定ファイル内では、「コメント」記述を許可しております。
コメントは、`#`から始まる文字列がコメントになります。

```conf
# 設定ファイルのコメントです。
# この # から始まる文字列は、すべて無視されます。
# ---------------------------------------------------------
app.name = "Application"
# app.debug = true # app.debugはコメントなので、設定上無効です
```

## パラメータ名
パラメータ名に指定できる文字は、半角英数字、アンダーバー、ピリオドです。
パラメータ名には以下の制約がある点に注意してください。

* パラメータ名の先頭、最後尾にピリオド、アンダーバーの文字は使用できない
* パラメータ名の先頭に、数字は使用できない
* 連続したピリオドは使用できない
* ピリオドの直後に、数字、アンダーバー等の文字は使用できない
* 大文字、小文字の区別はしない
* 重複したパラメータ名は使用できない

```conf
app.name = "Name"  # OK
app.0value = 100   # NG: ピリオドの直後に数字はNG
app._value = 200   # NG: ピリオドの直後に_はNG
app..value = 300   # NG: 連続したピリオドはNG
app.name = "Name2" # NG: 重複した名前はNG
APP.NAME = "Name3" # NG: 大文字、小文字は区別しないため、app.nameと同じ
0app.value = 500   # NG: 先頭に数字はNG
_app.value = 200   # NG: 先頭にアンダーバー、もしくはピリオドはNG
```

パラメータ名は、ピリオド区切りで指定することにより、Golang上では以下のように展開されます。

```conf
app.info.name = "sample"
app.info.date = 2018-01-01 00:00:00
# map[string]interface{}{
#     "app": map[string]interface{}{
#         "info": map[string]interface{}{
#             "name": "sample",                      // string
#             "date": 2018-01-01 00:00:00 +0000 UTC, // time.Time
#         },
#     },
# }
```

## パラメータの値

パラメータに指定できる値は、下記の通りです。

| 指定できる値  | 指定方法例 | 展開される型 |
|:--           |:-- |:--|
| 小数点       | `3.141592`|  float32 |
| 10進数       | `1,000,000` | int |
| 8進数        | `0644` | int |
| 16進数       | `0xFF` | int |
| 日付         | `2018-03-20 00:00:00`| time.Time |
| 時間         | `100s` | int64 |
| サイズ単位   | `100MB` | int64 |
| 真偽値       | `true` | bool |
| 文字列       | `"Hello World"` | string |
| 複数行文字列  | `"""Hello World"""` | string |
| 環境変数      | `$DEBUG` | string |
| 配列         | `[1, 2, 3]` | 各型の配列型。ex) []int{...}|

## 設定ファイルの「モード名」

「モード名」を使用することで、必要な設定のみを反映することができます。 

```conf
app.name  = "Development"
app.value = 100

# 本番環境用設定
[production]
app.name = "Production"
```

```go
    // app.name --->  Production
    // app.value ---> 100
    err := config.Parse("path/to/config.conf", "production", &conf);
    //                                         [production] を指定
```
本来パラメータ名の重複は不可能ですが、「モード名」が違う場合は同名のパラメータ名を使用することが可能です。

### 「モード名」を必須にする
設定ファイル内に、必ず「モード名」を付与したい場合もあります。その際には、 `ParseMode`関数を使用します。

```conf
[config1]
  http.port   = 8080
  http.server = "example.com"
  http.log    = "log/access1.log"

[config2]
  http.port   = 8081
  http.server = "example.ne.jp"
  http.log    = "log/access2.log"

[config3]
  http.port   = 8082
  http.server = "example.net"
  http.log    = "log/access3.log"
```

```go
    // 設定ファイルをパース。モード名が設定されていない、パラメータが存在する場合エラーとなる
    p, err := config.ParseMode("path/to/config.conf");
    if err != nil {
        panic(err)
    }

    var conf Config

    // http.port ---> 8080
    // http.server ---> "example.com"
    // http.log ---> "log/access1.log"
    p.Unmarshal(p.Data("config1"), &conf)
    fmt.Println(conf)

    // http.port ---> 8081
    // http.server ---> "example.ne.jp"
    // http.log ---> "log/access2.log"
    p.Unmarshal(p.Data("config2"), &conf)
    fmt.Println(conf)
```

## 付属ツール - cfgtool
`cfgtool`コマンドを使用することで、設定ファイルの記述内容のチェックや、設定ファイル内容をJSONに変換できます。

```
Usage:
    cfgtool check <filename>  Check configuration file.
    cfgtool json <filename>   Configuration file to JSON.

Example:
    cfgtool check app.conf
```

設定ファイルの内容をチェックする場合は、サブコマンドに`check`を渡します。
記述内容にエラーがあった場合は、エラー内容と、エラーのあった行番号を返却します。
```
[user@localhost ~]$ cfgtool check app.conf
syntax error:16: "app.flag" boolean invalid value
```

JSONに変換する場合は、サブコマンドに`json`を渡します。

```
[user@localhost ~]$ cfgtool json app.conf
{"app":{"name":"sample"}...}
```