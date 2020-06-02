package steamcmd

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type WorkshopDownloader struct {
	steamCmdPath, ArchivePath string
	cmd                       *SteamCmd
	Delete                    bool //delete when done
}

func NewWorkshopDownloader(scPath string, aPath string, d bool) (*WorkshopDownloader, error) {
	c, err := NewSteamCmd(scPath + "/steamcmd.sh")
	if err != nil {
		return nil, err
	}
	if err = c.Run(); err != nil {
		return nil, err
	}
	return &WorkshopDownloader{cmd: c, steamCmdPath: scPath, ArchivePath: aPath, Delete: d}, nil
}

func (w *WorkshopDownloader) Cache(appId, publishedFileId string) (string, error) {
	script := fmt.Sprintf("workshop_download_item %s %s", appId, publishedFileId)
	result, err := w.RunScript(script)
	if err != nil {
		return "", err
	}
	if strings.Contains(result, "ERROR!") || !strings.Contains(result, "Success") {
		//下载失败
		return "", errors.New(result)
	}
	//find path
	pos1 := strings.IndexByte(result, '"')
	pos2 := strings.LastIndexByte(result, '"')
	itemPath := result[pos1+1 : pos2]
	appPath := w.steamCmdPath + "/steamapps/workshop/content/" + appId + "/"
	filename := fmt.Sprintf("%s/%s.zip", w.ArchivePath, publishedFileId)
	file, err := os.Create(filename)
	//create zip file
	if err != nil {
		return "", err
	}
	defer file.Close()

	//filing
	archive := zip.NewWriter(file)
	defer archive.Close()
	archive.SetComment("archive create by 993651481@qq.com ")

	err = filepath.Walk(itemPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		fn := strings.Replace(path, appPath, "", -1)
		w, err := archive.Create(fn)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})

	if w.Delete {
		err = os.RemoveAll(itemPath)
	}

	return filename, err
}

func (w *WorkshopDownloader) RunScript(script string) (string, error) {
	result, err := w.cmd.RunScript(script)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
