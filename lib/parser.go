package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"strings"
)

type PackageInstall struct {
	X64 string `json:"x64,omitempty"`
	ARM string `json:"arm,omitempty"`
}

type PackageDetails struct {
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Install PackageInstall `json:"install"`
	Deps    []string       `json:"deps"`
	Tags    []string       `json:"tags"`
	Files   [][]string     `json:"files,omitempty"`
}

// Gets the installation setting of the current architecture
func (pd *PackageDetails) getInstall() (string, error) {
	if runtime.GOARCH == "amd64" {
		return pd.Install.X64, nil
	}
	if runtime.GOARCH == "arm" {
		return pd.Install.ARM, nil
	}
	return "", errors.New("unsupported architecture")
}

func (pd *PackageDetails) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Name:\t")
	buffer.WriteString(pd.Name)
	buffer.WriteString("\nVersion:\t")
	buffer.WriteString(pd.Version)
	buffer.WriteString("\nInstall:\t")
	install, err := pd.getInstall()
	if err != nil {
		panic(err)
	}
	buffer.WriteString(install)
	if len(pd.Deps) > 0 {
		buffer.WriteString("\nDeps:\t")
		for _, dep := range pd.Deps {
			buffer.WriteString(dep)
		}
	}
	if len(pd.Tags) > 0 {
		buffer.WriteString("\nTags:\t")
		buffer.WriteString(strings.Join(pd.Tags, ","))
	}

	return buffer.String()
}

func (pd *PackageDetails) FromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, pd); err != nil {
		return err
	}
	return nil
}

func NewPackageDetails() *PackageDetails {
	return &PackageDetails{}
}

func (pd *PackageDetails) LocalFiles() []string {
	newList := make([]string, len(pd.Files))
	for i, file := range pd.Files {
		newList[i] = file[0]
	}
	return newList
}

func (pd *PackageDetails) InstallFiles() []string {
	newList := make([]string, len(pd.Files))
	for i, file := range pd.Files {
		newList[i] = file[1]
	}
	return newList
}
