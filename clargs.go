package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tau/lib"
)

var commandMap = map[string]lib.CommandArg{
	"install":   lib.CommandInstall,
	"i":         lib.CommandInstall,
	"uninstall": lib.CommandUninstall,
	"pack":      lib.CommandPack,
	"unpack":    lib.CommandUnpack,
	"version":   lib.CommandVersion,
	"info":      lib.CommandInfo,
	"help":      lib.CommandHelp,
}

type ClArgError struct {
	arg     string
	message string
}

func (e *ClArgError) Error() string {
	return fmt.Sprintf("%s - %s", e.arg, e.message)
}

func packArgs(conf *lib.TauConfig, args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if err := conf.AddFlagString(arg); err != nil {
				log.Fatalf("Error parsing argument '%s': %s", arg, err)
			}
			continue
		}
		abs, err := filepath.Abs(arg)
		if err != nil {
			log.Fatalf("Error parsing argument '%s': %s", arg, err)
		}
		_, err = os.Stat(abs)
		if err != nil {
			return &ClArgError{abs, fmt.Sprintf("Unknown command line argument, %s", err)}
		}
		conf.Files = append(conf.Files, arg)
		conf.BaseDir = filepath.Dir(abs)
	}
	return nil
}

func unpackArgs(conf *lib.TauConfig, args []string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if err := conf.AddFlagString(arg); err != nil {
				log.Fatalf("Error parsing argument '%s': %s", arg, err)
			}
			continue
		}
		_, err := os.Stat(arg)
		if err == nil {
			conf.Files = append(conf.Files, arg)
		}
	}
	return nil
}

func ParseClArgs(conf *lib.TauConfig) {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tau <install|uninstall|pack|unpack|version> [options]")
		os.Exit(1)
	}
	conf.Cmd = commandMap[os.Args[1]]
	if conf.Cmd == lib.CommandPack {
		if err := packArgs(conf, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}
	if conf.Cmd == lib.CommandUnpack {
		if err := unpackArgs(conf, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") {
			if err := conf.AddFlagString(arg); err != nil {
				log.Fatalf("Error parsing argument '%s': %s", arg, err)
			}
		}
	}
}
