package lib

import (
	"fmt"
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
	Cmd        CommandArg
	InstallDir string
	targetAll  bool
	help       bool
	Verbose    bool
	Files      []string
	BaseDir    string
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
