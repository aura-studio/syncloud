package pusher

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type TaskList struct {
	Tasks map[string][]Pair
}

type Pair struct {
	RemoteFilePath string
	LocalFilePath  string
}

func (f *TaskList) Add(remote string, remoteFilePath string, localFilePath string) {
	f.Tasks[remote] = append(f.Tasks[remote], Pair{
		RemoteFilePath: remoteFilePath,
		LocalFilePath:  localFilePath,
	})
}

func NewTaskList(c Config) *TaskList {
	fileList := &TaskList{
		Tasks: make(map[string][]Pair),
	}

	for _, remote := range c.Remotes {
		u, err := url.Parse(remote)
		if err != nil {
			log.Panicf("parsing remote url error: %v", err)
		}
		var remotePath string
		if strings.HasPrefix(u.Path, "/") {
			remotePath = u.Path[1:]
		} else {
			remotePath = u.Path
		}

		for _, local := range c.Locals {
			// check local is dir
			info, err := os.Stat(local)
			if err != nil {
				log.Panicf("get local file info error: %v", err)
			}
			if info.IsDir() {
				filepath.Walk(local, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						log.Panicf("find local files error: %v", err)
					}
					if !info.IsDir() {
						relFilePath, err := filepath.Rel(local, path)
						if err != nil {
							log.Panicf("get relative path error: %v", err)
						}
						remoteFilePath := filepath.ToSlash(filepath.Join(remotePath, relFilePath))
						localFilePath := path
						fileList.Add(remote, remoteFilePath, localFilePath)
					}
					return nil
				})
			} else {
				remoteFilePath := filepath.ToSlash(filepath.Join(remotePath, filepath.Base(local)))
				localFilePath := local
				fileList.Add(remote, remoteFilePath, localFilePath)
			}
		}
	}

	return fileList
}
