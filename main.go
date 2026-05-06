package main

import (
	"C"
	"fmt"
	"os"
	"tau/lib"
	"unsafe"
)

func main() {
	pkgData := lib.NewPackageDetails()
	conf := lib.NewTauConfig()
	ParseClArgs(conf)
	err := pkgData.FromFile("example/tau.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if conf.Cmd == lib.CommandPack {
		err := pack(conf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if conf.Cmd == lib.CommandUnpack {
		err := unpack(conf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	fmt.Println(pkgData)
}

func pack(conf *lib.TauConfig) error {
	pkgData := lib.NewPackageDetails()
	if err := pkgData.FromFile(conf.Files[0]); err != nil {
		return err
	}
	packList := lib.NewPackList()
	for _, v := range pkgData.Files {
		packList.Add(v[0], v[1])
	}
	if err := packList.Pack(pkgData, conf); err != nil {
		return err
	}
	return nil
}

func unpack(conf *lib.TauConfig) error {
	return lib.Unpack(conf.Files[0])
}

//export lib.NewPackList

//export CNewPackList
func CNewPackList() unsafe.Pointer {
	return unsafe.Pointer(lib.NewPackList())
}

//export CUnpack
func CUnpack(path *C.char) {
	lib.Unpack(C.GoString(path))
}
