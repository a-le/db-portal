package jsminifier

import (
	"os"
	"path/filepath"
	"time"

	"github.com/dchest/jsmin"
)

func getFolderModTime(folder string) (latestModTime time.Time, err error) {
	fileList, err := filepath.Glob(filepath.Join(folder, "*.js"))
	if err != nil {
		return time.Time{}, err
	}

	for _, filePath := range fileList {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return time.Time{}, err
		}

		if fileInfo.ModTime().After(latestModTime) {
			latestModTime = fileInfo.ModTime()
		}
	}

	return
}

func getFileModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func minifyJSContent(jsContent []byte) ([]byte, error) {
	minifiedContent, err := jsmin.Minify(jsContent)
	if err != nil {
		return nil, err
	}
	return minifiedContent, nil
}

type infos struct {
	folderModTime time.Time
	fileModTime   time.Time
	Expired       bool
}

func (infos infos) ModTime() time.Time {
	return infos.folderModTime
}

func GetInfos(folder string, filePath string) (infos infos, err error) {
	infos.folderModTime, err = getFolderModTime(folder)
	if err != nil {
		return
	}
	infos.fileModTime, err = getFileModTime(filePath)
	if err != nil {
		return
	}
	infos.Expired = infos.folderModTime.After(infos.fileModTime)
	return
}

func Combinify(folder string, minFilePath string) error {

	var combined []byte
	var minified []byte

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".js" {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if len(combined) > 0 {
				combined = append(combined, "\n"...)
			}
			combined = append(combined, b...)
		}
		return nil
	})

	if err != nil {
		return err
	}

	minified, err = minifyJSContent(combined)
	if err != nil {
		return err
	}

	return os.WriteFile(minFilePath, minified, 0644)
}
