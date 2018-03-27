package config

import (
	"fmt"
	"testing"
)

type ConfigTest struct {
	App  App
	Http Http
}

type ConfigTest2 struct {
	App  App2
	Http Http
}

type App struct {
	Name string
	Flag bool
}

type App2 struct {
	Name int
	Flag string
}

type Http struct {
	Log Log
}

type Log struct {
	Name string
}

func TestConfig(t *testing.T) {
	var app ConfigTest
	// 存在しないファイルの読み込み
	if err := Parse("test/noconf", "development", &app); err == nil {
		t.Fatal(err)
	}
	if err := Parse("test/normal_test1.conf", "development", &app); err != nil {
		t.Fatal(err)
	}
	if err := Parse("test/normal_test2.conf", "config2", &app); err != nil {
		t.Fatal(err)
	}
	if err := Parse("test/normal_test3.conf", "development", &app); err != nil {
		t.Fatal(err)
	}
	if err := Parse("test/normal_test2.conf", "config4", &app); err == nil {
		t.Fatal(err)
	}
	if err := Parse("test/normal_test1.conf", "development", nil); err == nil {
		t.Fatal(err)
	}
	if err := Parse("test/normal_test4.conf", "development", nil); err == nil {
		t.Fatal(err)
	}
	if err := Parse("test/normal_test5.conf", "development", nil); err == nil {
		t.Fatal(err)
	}
	var app2 ConfigTest2
	if err := Parse("test/normal_test1.conf", "development", &app2); err == nil {
		t.Fatal(err)
	}
}

func TestConfigMode(t *testing.T) {
	if _, err := ParseMode("test/noconf"); err == nil {
		t.Fatal(err)
	}
	if _, err := ParseMode("test/normal_test1.conf"); err == nil {
		t.Fatal(err)
	}

	p, err := ParseMode("test/normal_test6.conf")
	if err != nil {
		t.Fatal(err)
	}
	// map[string]interface{}
	if fmt.Sprint(p.DataAll()) != "map[config:map[app:map[name:test]]]" {
		t.Fatal("p.DataAll() is error")
	}
	if fmt.Sprint(p.Data("config")) != "map[app:map[name:test]]" {
		t.Fatal("p.Data() is error")
	}
	if p.Data("nodata") != nil {
		t.Fatal("p.Data() is error")
	}

	p, err = ParseMode("test/normal_test2.conf")
	if err != nil {
		t.Fatal(err)
	}
	d1 := p.Data("config1")
	d2 := p.Data("config2")

	p.Merge(d1, d2)

	var merge = make(map[string]interface{})
	if err := p.Unmarshal(d1, &merge); err != nil {
		t.Fatal(err)
	}
}
