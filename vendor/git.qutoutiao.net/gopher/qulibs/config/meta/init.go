package meta

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"git.qutoutiao.net/pedestal/discovery/logger"
)

var (
	value     = atomic.Value{}
	valueOnce sync.Once
)

func Init() {
	valueOnce.Do(func() {
		filename, err := findMetaFilename()
		if err != nil {
			value.Store(&Metadata{
				data:     map[string]string{},
				filename: "",
				err:      err,
			})

			return
		}

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			value.Store(&Metadata{
				data:     map[string]string{},
				filename: filename,
				err:      err,
			})

			return
		}

		meta := &Metadata{
			data:     map[string]string{},
			filename: filename,
			err:      err,
		}

		for _, line := range strings.Split(string(data), "\n") {
			kv := strings.SplitN(line, "=", 2)
			if len(kv) != 2 {
				continue
			}

			meta.data[kv[0]] = kv[1]
		}

		value.Store(meta)
	})
}

func App() string {
	return Load().App()
}

func Git() string {
	return Load().Git()
}

func Version() string {
	return Load().Version()
}

func findMetaFilename() (filename string, err error) {
	defer func() {
		logger.Warnf("resolved %s from %s", GitMetaName, filename)
	}()

	root, err := os.Getwd()
	if err != nil {
		return
	}

	logger.Infof("resolve %s from %s ...", GitMetaName, root)

	for {
		if len(root) == 0 || root == "/" {
			return
		}

		filename = filepath.Join(root, GitMetaName)

		fstat, ferr := os.Stat(filename)
		if ferr == nil && !fstat.IsDir() {
			return
		}

		folders := strings.Split(root, string(filepath.Separator))
		if len(folders) <= 1 {
			return
		}

		root = string(filepath.Separator) + filepath.Join(folders[:len(folders)-1]...)
	}
}
