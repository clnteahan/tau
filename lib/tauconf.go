package lib

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
	CommandAdd
)

type Flag int

const (
	FlagUnknown Flag = iota
	FlagAll
	FlagVerbose
	FlagHelp
)

type TauConfig struct {
	InstallDir   string
	ManifestPath string
	OutPath      string
	targetAll    bool
	Verbose      bool
	Files        []string
	BaseDir      string
}

func (conf *TauConfig) AddFlag(flag Flag) error {
	if flag == FlagUnknown {
		return UnknownFlagError(nil)
	}
	if flag == FlagAll {
		conf.targetAll = true
	}
	if flag == FlagVerbose {
		conf.Verbose = true
	}

	return nil
}

func NewTauConfig() *TauConfig {
	return &TauConfig{
		InstallDir: "/usr/share/tau",
		Verbose:    false,
	}
}

func AddFiles(pdDir string, files ...[]string) error {
	pd := NewPackageDetails()
	if err := pd.FromFile(pdDir); err != nil {
		return err
	}
	pd.Files = append(pd.Files, files...)
	return pd.ToFile(pdDir)
}
