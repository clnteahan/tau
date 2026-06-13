package main

import (
	"C"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"tau/lib"
	"unsafe"
)

func main() {
	conf := lib.NewTauConfig()
	ParseClArgs(conf)
}

type ClArgError struct {
	arg     string
	message string
}

func (e *ClArgError) Error() string {
	return fmt.Sprintf("%s - %s", e.arg, e.message)
}

func packArgs(conf *lib.TauConfig, args []string) error {
	if len(args) < 1 {
		log.Fatalf("Usage: %s pack <manifest>", os.Args[0])
	}
	conf.Cmd = lib.CommandPack
	fset := flag.NewFlagSet("pack", flag.ContinueOnError)
	manifest := fset.String("manifest", "", "Manifest file to pack")
	err := fset.Parse(args)
	if err != nil {
		return err
	}
	if *manifest == "" {
		manifest = &args[0]
	}
	conf.BaseDir = path.Dir(*manifest)
	conf.Files = make([]string, 1)
	conf.Files[0] = *manifest

	err = pack(conf)

	return err
}

func unpackArgs(conf *lib.TauConfig, args []string) error {
	if len(args) < 1 {
		log.Fatalf("Usage: %s unpack <file>", os.Args[0])
	}
	conf.Cmd = lib.CommandUnpack

	_, err := os.Stat(args[0])
	if err != nil {
		return err
	}
	conf.Files = make([]string, 1)
	conf.Files[0] = args[0]

	err = unpack(conf)
	return err
}

func addArgs(conf *lib.TauConfig, args []string) error {
	var files []string
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") {
			if err := conf.AddFlagString(arg); err != nil {
				log.Fatalf("Error parsing argument '%s': %s", arg, err)
			}
			continue
		}
		_, err := os.Stat(arg)
		if err != nil {
			return &ClArgError{arg, fmt.Sprintf("Unable to open file: %s", err)}
		}
		files = append(files, arg)
	}
	return lib.AddFiles(args[0], files)
}

func ParseClArgs(conf *lib.TauConfig) {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <install|uninstall|pack|unpack|version>", os.Args[0])
	}
	var err error = nil
	switch os.Args[1] {
	case "pack":
		err = packArgs(conf, os.Args[2:])
	case "unpack":
		err = unpackArgs(conf, os.Args[2:])
	default:
		err = &ClArgError{os.Args[1], fmt.Sprintf("Unknown argument '%s'", os.Args[1])}
	}

	if err != nil {
		log.Fatalf("Error parsing argument '%s': %s", os.Args[1], err)
	}

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
