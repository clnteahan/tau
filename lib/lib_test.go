package lib

import (
	"fmt"
	"os"
	"testing"
)

func absWd(relPath string) string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd + relPath
}

func parseTau() (*PackageDetails, error) {
	pkgData := NewPackageDetails()
	if err := pkgData.FromFile(absWd("/../example/tau.json")); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return pkgData, nil
}

func TestParse(t *testing.T) {
	// Test to be added
}

func TestPack(t *testing.T) {
	pkgData, err := parseTau()
	if err != nil {
		t.Fatal(err)
	}
	packList := NewPackList()
	packList.Add(absWd("/packing.go"), "/usr/share/tau/packing.go")
	packList.Add(absWd("/parser.go"), "/usr/share/tau/parser.go")
	packList.Add(absWd("/../main.go"), "/usr/share/tau/main.go")
	if err := packList.Pack(pkgData); err != nil {
		t.Error(err)
	}

	if err := Unpack(absWd("/tau.tau")); err != nil {
		t.Error(err)
	}
}
