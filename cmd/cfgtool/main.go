package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ochipin/config/parser"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("%s", help())
		return
	}

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s", help())
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "check":
		err = check(os.Args[2])
	case "json":
		err = tojson(os.Args[2])
	default:
		err = fmt.Errorf("error: %s sub command unknown", os.Args[1])
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func help() string {
	mes := []string{
		"",
		"Usage:",
		"    cfgtool check <filename>  Check configuration file.",
		"    cfgtool json <filename>   Configuration file to JSON.",
		"",
		"Example:",
		"    cfgtool check app.conf",
		"",
		"",
	}
	return strings.Join(mes, "\n")
}

func check(fname string) error {
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	if _, err := parser.Parse(buf); err != nil {
		return err
	}
	return nil
}

func tojson(fname string) error {
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	p, err := parser.Parse(buf)
	if err != nil {
		return err
	}

	buf, err = json.Marshal(p.Data())
	if err != nil {
		return err
	}

	fmt.Println(string(buf))
	return nil
}
