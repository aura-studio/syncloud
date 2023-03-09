package pusher

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
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

		for _, local := range c.Locals {
			// find all files
			filepath.Walk(local, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Panicf("find local files error: %v", err)
				}
				if !info.IsDir() {
					relFilePath, err := filepath.Rel(local, path)
					if err != nil {
						log.Panicf("get relative path error: %v", err)
					}
					remoteFilePath := filepath.Join(u.Path, relFilePath)
					localFilePath := path
					fileList.Add(remote, remoteFilePath, localFilePath)
				}
				return nil
			})
		}
	}

	return fileList
}
