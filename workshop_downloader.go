package steamcmd

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/silenceper/pool"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type WorkshopDownloader struct {
	pool                      pool.Pool
	steamCmdPath, ArchivePath string
}

func NewWorkshopDownloader(pSize int, scPath string) (*WorkshopDownloader, error) {
	poolConfig := &pool.Config{
		InitialCap: pSize, //资源池初始连接数
		MaxIdle:    pSize, //最大空闲连接数
		MaxCap:     pSize, //最大并发连接数
		Factory: func() (interface{}, error) {
			cmd, err := NewSteamCmd(scPath + "/steamcmd.sh")
			if err != nil {
				return nil, err
			}
			if err = cmd.Run(); err != nil {
				return nil, err
			}
			return cmd, nil
		},
		Close: func(i interface{}) error {
			return i.(*SteamCmd).Close()
		},

		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		//IdleTimeout: 15 * time.Second,
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}

	return &WorkshopDownloader{pool: p, steamCmdPath: scPath}, nil
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
	pos1 := strings.IndexByte(result, '"')
	pos2 := strings.LastIndexByte(result, '"')
	itemPath := result[pos1+1 : pos2]
	appPath := w.steamCmdPath + "/steamapps/workshop/content/" + appId + "/"
	filename := fmt.Sprintf("%s/%s.zip", w.ArchivePath, publishedFileId)
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
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
		_, err = io.Copy(w, f)
		return err
	})
	return filename, err
}

func (w *WorkshopDownloader) RunScript(script string) (string, error) {
	p, err := w.pool.Get()
	if err != nil {
		return "", err
	}
	cmd := p.(*SteamCmd)
	defer w.pool.Put(p)
	result, err := cmd.RunScript(script)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
