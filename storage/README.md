
データ管理用ライブラリ
===
マップデータや、構造体のデータを管理するライブラリです。

```go
package main

import (
    "fmt"

    "github.com/ochipin/config/storage"
)

type Info struct {
    App    string
    Port   int
    Detail string
}

func main() {
    var data = make(storage.Storage)
    /*
       map[string]interface{}{
           "config": map[string]interface{}{
               "app":    "appname",
               "port":   8080,
               "detail": "sample app",
           },
       }
    */
    data.Set("config.app", "appname")
    data.Set("config.port", 8080)
    data.Set("config.detail", "sample app")

    fmt.Println(data.Str("config.app"))    // appname
    fmt.Println(data.Int("config.port"))   // 8080
    fmt.Println(data.Str("config.detail")) // sample app

    for k, v := range data.Map("config") {
        // app : appname
        // port : 8080
        // detail : sample app
        fmt.Println(k, ":", v)
    }

    var info Info
    if err := data.Unmarshal("config", &info); err != nil {
        panic(err)
    }
    fmt.Println(info.App)    // app
    fmt.Println(info.Port)   // 8080
    fmt.Println(info.Detail) // sample app
}
```

## データの登録
データの登録は、すべて`Set`関数で行う。

```go
var data = make(storage.Storage)

data.Set("v01", 100)
data.Set("v02", int64(200))
data.Set("v03", uint(300))
data.Set("v04", uint64(400))
data.Set("v05", 3.14)
data.Set("v06", true)
data.Set("v07", "name")
data.Set("v08", []int{1, 2, 3})
// 多次元配列も可能
data.Set("v08", [][]uint{
    []uint{1, 2, 3},
    []uint{4, 5, 6},
    []uint{7, 8, 9},
})
// map登録時は、必ず map[string]interface{} であること
data.Set("v09", map[string]interface{}{...})
// 時刻登録
data.Set("v10", time.Now())
```

## 型チェック

登録されたデータの型のチェックを行うことができる。

```go
var data = make(storage.Storage)

// int: 100 をセット
data.Set("value", 100)

// "value" にセットした値が int 型であるかチェック
if err := data.IsInt("value"); err != nil {
    // int 型ではない場合、エラーとなる
    panic(err)
}
```

型チェックは、次の関数を用いて可能。

| 関数名     | 説明 |
|:--         |:-- |
| `IsInt`    | 登録されたデータが`int`型であるかチェックする |
| `IsInt64`  | 登録されたデータが`int64`型であるかチェックする |
| `IsUint`   | 登録されたデータが`uint`型であるかチェックする |
| `IsUint64` | 登録されたデータが`uint64`型であるかチェックする|
| `IsFloat`  | 登録されたデータが`float32`,`float64`型であるかチェックする |
| `IsBool`   | 登録されたデータが`bool`型であるかチェックする |
| `IsStr`    | 登録されたデータが`string`型であるかチェックする |
| `IsSlice`  | 登録されたデータが`[]interface{}`型であるかチェックする |
| `IsMap`    | 登録されたデータが`map[string]interface{}`型であるかチェックする |
| `IsTime`   | 登録されたデータが`time.Time`型であるかチェックする |

## 登録されたデータの取得
登録されたデータを取得する方法を説明する。

```go
var data = make(storage.Storage)

// ... データ処理 ...
data.Set("value", 100)
data.Set("map", map[string]interface{}{
    "name":  "Your name",
    "PI":    3.14,
    "flag":  true,
    "slice": []uint {
        1, 2, 3,
    }
})

// int や string, float, bool などのデータを取得する
var i = data.Int("value")
fmt.Println(i) // 100 を表示

// マップ処理
for k, v := range data.Map("map") {
    // name : Your name
    // PI : 3.14
    // flag : true
    // ...
    fmt.Println(k, ":", v)
}

// 配列処理
for i, v := range data.Slice("map.slice") {
    // 0 : 1
    // 1 : 2
    // 2 : 3
    fmt.Println(i, ":", v)
}

// int データを int64 として取得する
var i64 = data.Int64("value")
```
登録データは、次の関数で取り出す。

| 関数名     | 説明 |
|:--        |:-- |
| `Int`     | 登録されているデータを、`int`型で取得する |
| `Int64`   | 登録されているデータを、`int64`型で取得する |
| `Uint`    | 登録されているデータを、`uint`型で取得する |
| `Uint64`  | 登録されているデータを、`uint64`型で取得する |
| `Float32` | 登録されているデータを、`float32`型で取得する |
| `Float64` | 登録されているデータを、`float64`型で取得する |
| `Bool`    | 登録されているデータを、`bool`型で取得する |
| `Str`     | 登録されているデータを、`string`型で取得する |
| `Slice`   | 登録されているデータを、`[]interface{}`型で取得する |
| `Map`     | 登録されているデータを、`map[string]interface{}`型で取得する |
| `Time`    | 登録されているデータを、`time.Time`型で取得する |

## Unmarshal によるデータの移動
データが、`Storage`型にあると何かと不便であることが多い。  
`Unmarshal` を用いて、データを構造体、マップ、配列に移動することが可能。

### 構造体にデータを格納する方法
```go
type Info struct {
    App    string
    Port   int
    Detail string
}

func main() {
    var data = make(storage.Storage)
    data.Set("config.app", "appname")
    data.Set("config.port", 8080)
    data.Set("config.detail", "sample app")

    var info Info
    if err := data.Unmarshal("config", &info); err != nil {
        panic(err)
    }
    fmt.Println(info.App)    // app
    fmt.Println(info.Port)   // 8080
    fmt.Println(info.Detail) // sample app
}
```

### 配列データを格納する方法

```go
func main() {
    var data = make(storage.Storage)
    data.Set("config.app", "appname")
    data.Set("config.port", 8080)
    data.Set("config.detail", "sample app")
    data.Set("config.flags", []bool{true, true, false})

    var b []bool
    if err := data.Unmarshal("config", &b); err != nil {
        panic(err)
    }
    for i, v := range b {
        // 0 : true
        // 1 : true
        // 2 : false
        fmt.Println(i, ":", v)
    }
}
```

## 注意事項
登録データに、`func()`型が含まれていると、そのデータは無視される。

```go
var data = make(storage.Storage)

data.Set("name.func", func() {})
data.Get("name.func") // Nil
```

また、`Set`関数で指定するキー名の"."(ピリオド)は、データの区切りとして扱われる点に注意すること。

```go
var data = make(storage.Storage)

data.Set("config.app", "appname")
data.Set("config.port", 8080)
data.Set("config.detail", "sample app")
/*
   // データ構造は、次のようになる
   map[string]interface{}{
       "config": map[string]interface{}{
           "app":    "appname",
           "port":   8080,
           "detail": "sample app",
       },
   }
*/
```
