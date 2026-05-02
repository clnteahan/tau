package main

import (
	"fmt"
	"os"
	"tau/lib"
)

func main() {
	pkgData := lib.NewPackageDetails()
	conf := NewTauConfig()
	command := ParseClArgs(conf)
	err := pkgData.FromFile("example/tau.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if command == CommandPack {
		err := pack(conf)
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	fmt.Println(pkgData)
}

func pack(conf *TauConfig) error {
	return nil
}
