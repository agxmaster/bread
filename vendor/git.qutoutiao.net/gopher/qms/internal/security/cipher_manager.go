package security

import (
	"errors"
	"fmt"
	"os"

	"git.qutoutiao.net/gopher/qms/internal/pkg/goplugin"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

const pluginSuffix = ".so"

//CipherPlugins is a map
var cipherPlugins map[string]func() Cipher

//InstallCipherPlugin is a function
func InstallCipherPlugin(name string, f func() Cipher) {
	cipherPlugins[name] = f
}

//GetCipherNewFunc is a function
func GetCipherNewFunc(name string) (func() Cipher, error) {
	if f, ok := cipherPlugins[name]; ok {
		return f, nil
	}
	qlog.Tracef("try to load cipher [%s] from go plugin", name)
	f, err := loadCipherFromPlugin(name)
	if err == nil {
		cipherPlugins[name] = f
		return f, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	return nil, fmt.Errorf("unknown cipher plugin [%s]", name)
}

func loadCipherFromPlugin(name string) (func() Cipher, error) {
	c, err := goplugin.LookUpSymbolFromPlugin(name+pluginSuffix, "Cipher")
	if err != nil {
		return nil, err
	}
	customCipher, ok := c.(Cipher)
	if !ok {
		return nil, errors.New("symbol from plugin is not type Cipher")
	}
	f := func() Cipher {
		return customCipher
	}
	return f, nil
}

func init() {
	cipherPlugins = make(map[string]func() Cipher)
}
