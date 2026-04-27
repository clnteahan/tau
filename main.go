package main

import (
	"fmt"
	"os"
	"tau/lib"
)

func main() {
	pkgData := lib.NewPackageDetails()
	err := pkgData.FromFile("example/tau.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(pkgData)
}
