package conf

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/fileutil"
	"gopkg.in/yaml.v2"
)

// GetConfDir return the config dir
func GetConfDir() string {
	return fileutil.GetConfDir()
}

// Get is for to get the value of configuration key
func Get(key string) interface{} {
	return qconf.Get(key)
}

// Exist check the configuration key existence
func Exist(key string) bool {
	if value := qconf.Get(key); value != nil {
		return true
	}
	return false
}

// Unmarshal unmarshals the config into a Struct. Make sure that the tags
// on the fields of the structure are properly set.
func Unmarshal(obj interface{}) error {
	content, err := ioutil.ReadFile(fileutil.AppConfigPath())
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	return decoder.Decode(obj)
}

// UnmarshalFile 指定文件名来unmarshal数据
func UnmarshalFile(fname string, obj interface{}) error {
	content, err := ioutil.ReadFile(filepath.Join(GetConfDir(), fname))
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	return decoder.Decode(obj)
}

// GetBool is gives the key value in the form of bool
func GetBool(key string, defaultValue bool) bool {
	return qconf.GetBool(key, defaultValue)
}

// GetFloat64 gives the key value in the form of float64
func GetFloat64(key string, defaultValue float64) float64 {
	return qconf.GetFloat64(key, defaultValue)
}

// GetInt gives the key value in the form of GetInt
func GetInt(key string, defaultValue int) int {
	return qconf.GetInt(key, defaultValue)
}

// GetString gives the key value in the form of GetString
func GetString(key string, defaultValue string) string {
	return qconf.GetString(key, defaultValue)
}

// GetConfigs gives the information about all configurations
func GetConfigs() map[string]interface{} {
	return qconf.AllSettings()
}
