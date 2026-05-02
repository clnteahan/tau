package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type UnknownFlagError error

type CommandArg int

const (
	CommandUnknown CommandArg = iota
	CommandInstall
	CommandUninstall
	CommandPack
	CommandUnpack
	CommandVersion
	CommandInfo
	CommandHelp
)

var commandMap = map[string]CommandArg{
	"install":   CommandInstall,
	"i":         CommandInstall,
	"uninstall": CommandUninstall,
	"pack":      CommandPack,
	"unpack":    CommandUnpack,
	"version":   CommandVersion,
	"info":      CommandInfo,
	"help":      CommandHelp,
}

type Flag int

const (
	FlagUnknown Flag = iota
	FlagAll
	FlagVerbose
	FlagHelp
)

var flagMap = map[string]Flag{
	"-a":        FlagAll,
	"--all":     FlagAll,
	"-h":        FlagHelp,
	"--help":    FlagHelp,
	"-v":        FlagVerbose,
	"--verbose": FlagVerbose,
}

type TauConfig struct {
	InstallDir string
	targetAll  bool
	help       bool
	Verbose    bool
}

func (conf *TauConfig) AddFlag(flag Flag) error {
	if flag == FlagUnknown {
		return UnknownFlagError(nil)
	}
	if flag == FlagAll {
		conf.targetAll = true
	}
	if flag == FlagHelp {
		conf.help = true
	}
	if flag == FlagVerbose {
		conf.Verbose = true
	}

	return nil
}

func (conf *TauConfig) AddFlagString(flags string) error {
	if !strings.HasPrefix(flags, "-") {
		return UnknownFlagError(nil)
	}
	if strings.HasPrefix(flags, "--") {
		return conf.AddFlag(flagMap[flags])
	}
	if len(flags) == 2 {
		return conf.AddFlagString(flags)
	}
	for i := 1; i < len(flags); i++ {
		if err := conf.AddFlag(flagMap[fmt.Sprintf("-%s", flags[i])]); err != nil {
			return err
		}
	}

	return UnknownFlagError(nil)
}

func NewTauConfig() *TauConfig {
	return &TauConfig{
		InstallDir: "/usr/share/tau",
	}
}

func ParseClArgs(conf *TauConfig) CommandArg {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tau <install|uninstall|pack|unpack|version> [options]")
		os.Exit(1)
	}
	cmd := commandMap[os.Args[1]]
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") {
			if err := conf.AddFlagString(arg); err != nil {
				log.Fatalf("Error parsing argument '%s': %s", arg, err)
			}
		}
	}
	return cmd
}
