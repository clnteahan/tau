package lib

import (
	"archive/tar"
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Pack struct {
	localPath   string
	installPath string
}

type PackList []Pack

func NewPackList() *PackList {
	return &PackList{}
}

// Add file to pack list
func (pl *PackList) Add(localPath string, installPath string) {
	*pl = append(*pl, Pack{localPath: localPath, installPath: installPath})
}

// Remove the first instance of localFile from package if it exists, otherwise do nothing
func (pl *PackList) Remove(localPath string) {
	for i, v := range *pl {
		if v.localPath == localPath {
			*pl = append((*pl)[:i], (*pl)[i+1:]...)
			return
		}
	}
}

func (pl *PackList) ContainsLocal(localPath string) bool {
	for _, v := range *pl {
		if v.localPath == localPath {
			return true
		}
	}
	return false
}

func (pl *PackList) ContainsInstall(installPath string) bool {
	for _, v := range *pl {
		if v.installPath == installPath {
			return true
		}
	}
	return false
}

func (pl *PackList) Pack(pd *PackageDetails) error {
	var buf bytes.Buffer
	zl := zlib.NewWriter(&buf)
	tw := tar.NewWriter(zl)
	defer func(tw *tar.Writer) {
		err := tw.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(tw)
	for i, pack := range *pl {

		body, err := os.ReadFile(pack.localPath)

		if err != nil {
			return err
		}

		pSplit := strings.Split(pack.installPath, string(os.PathSeparator))
		cName := pSplit[len(pSplit)-1:][0]
		localDir := strings.Join(pSplit[:len(pSplit)-1], string(os.PathSeparator))
		hash := sha256.Sum256([]byte(localDir))
		cPath := fmt.Sprintf("files/%s/%s", base64.URLEncoding.EncodeToString(hash[:]), cName)
		fmt.Printf("Packing %s to %s\n", pack.localPath, cPath)

		pd.Files[i][0] = cPath

		hdr := &tar.Header{
			Name: cPath,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			return err
		}

	}

	body, err := json.Marshal(pd)
	if err != nil {
		return err
	}

	hdr := &tar.Header{
		Name:    fmt.Sprintf("%s.json", pd.Name),
		Mode:    0600,
		ModTime: time.Now(),
		Size:    int64(len(body)),
	}
	if err = tw.WriteHeader(hdr); err != nil {
		return err
	}

	if _, err := tw.Write([]byte(body)); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	err = zl.Close()
	if err != nil {
		return err
	}

	return os.WriteFile(fmt.Sprintf("%s.tau", pd.Name), buf.Bytes(), 0644)
}

func Unpack(path string) error {
	destDir := strings.ReplaceAll(path, ".tau", "")

	cFile, err := os.Open(path)
	if err != nil {
		return err
	}

	zr, err := zlib.NewReader(cFile)
	if err != nil {
		return err
	}

	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, hdr.Name)

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			_ = f.Close()
			return err
		}
		if f.Close() != nil {
			return err
		}
	}
	return nil
}
