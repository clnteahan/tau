package lib

import (
	"fmt"
	"os"
	"testing"
)

func parseTau() *PackageDetails {
	pkgData := NewPackageDetails()
	err := pkgData.FromFile("example/tau.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return pkgData
}

func TestParse(t *testing.T) {
	// Test to be added
}

func TestPack(t *testing.T) {
	pkgData := parseTau()
	packList := NewPackList()
	packList.Add("lib/packing.go", "/usr/share/tau/packing.go")
	packList.Add("lib/parser.go", "/usr/share/tau/parser.go")
	packList.Add("main.go", "/usr/share/tau/main.go")
	if err := packList.Pack(pkgData); err != nil {
		t.Error(err)
	}

	if err := Unpack("tau.tau"); err != nil {
		t.Error(err)
	}
}
