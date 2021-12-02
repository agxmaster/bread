// 参考viper的实现

// 优先级如下：
// overrides[SetXXX]
// flag[需开发]
// env[需开发]
// config
// default

// 和viper的区别
// 1. 支持多文件
// 2. 可以判断该Key是否存在
// 3. 不支持remote

// 支持多个文件类型的解析

package qconf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"git.qutoutiao.net/gopher/qms/pkg/json"
	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	jww "github.com/spf13/jwalterweatherman"
	"gopkg.in/yaml.v2"
)

// ConfigMarshalError happens when failing to marshal the configuration.
type ConfigMarshalError struct {
	err error
}

// Error returns the formatted configuration error.
func (e ConfigMarshalError) Error() string {
	return fmt.Sprintf("While marshaling config: %s", e.err.Error())
}

var q *Qconf

func init() {
	q = New()
}

// UnsupportedConfigError denotes encountering an unsupported
// configuration filetype.
type UnsupportedConfigError string

// Error returns the formatted configuration error.
func (str UnsupportedConfigError) Error() string {
	return fmt.Sprintf("Unsupported Config Type %q", string(str))
}

// UnsupportedRemoteProviderError denotes encountering an unsupported remote
// provider. Currently only etcd and Consul are supported.
type UnsupportedRemoteProviderError string

// Error returns the formatted remote provider error.
func (str UnsupportedRemoteProviderError) Error() string {
	return fmt.Sprintf("Unsupported Remote Provider Type %q", string(str))
}

// ConfigFileNotFoundError denotes failing to find configuration file.
type ConfigFileNotFoundError struct {
	name, locations string
}

// Error returns the formatted configuration error.
func (fnfe ConfigFileNotFoundError) Error() string {
	return fmt.Sprintf("Config File %q Not Found in %q", fnfe.name, fnfe.locations)
}

// ConfigFileAlreadyExistsError denotes failure to write new configuration file.
type ConfigFileAlreadyExistsError string

// Error returns the formatted error when configuration already exists.
func (faee ConfigFileAlreadyExistsError) Error() string {
	return fmt.Sprintf("Config File %q Already Exists", string(faee))
}

// A DecoderConfigOption can be passed to viper.Unmarshal to configure
// mapstructure.DecoderConfig options
type DecoderConfigOption func(*mapstructure.DecoderConfig)

// DecodeHook returns a DecoderConfigOption which overrides the default
// DecoderConfig.DecodeHook value, the default is:
//
//  mapstructure.ComposeDecodeHookFunc(
//		mapstructure.StringToTimeDurationHookFunc(),
//		mapstructure.StringToSliceHookFunc(","),
//	)
func DecodeHook(hook mapstructure.DecodeHookFunc) DecoderConfigOption {
	return func(config *mapstructure.DecoderConfig) {
		config.DecodeHook = hook
	}
}

func DecodeTagName(tagName string) DecoderConfigOption {
	return func(config *mapstructure.DecoderConfig) {
		config.TagName = tagName
	}
}

// Qconf is a prioritized configuration registry. It
// maintains a set of configuration sources, fetches
// values to populate those, and provides them according
// to the source's priority.
// The priority of the sources is the following:
// 1. overrides
// 2. flags
// 3. env. variables
// 4. config file
// 5. key/value store
// 6. defaults
//
// For example, if values from the following sources were loaded:
//
//  Defaults : {
//  	"secret": "",
//  	"user": "default",
//  	"endpoint": "https://localhost"
//  }
//  Config : {
//  	"user": "root"
//  	"secret": "defaultsecret"
//  }
//  Env : {
//  	"secret": "somesecretkey"
//  }
//
// The resulting config will have the following values:
//
//	{
//		"secret": "somesecretkey",
//		"user": "root",
//		"endpoint": "https://localhost"
//	}
type Qconf struct {
	// Delimiter that separates a list of keys
	// used to access a nested value in one go
	keyDelim string

	// A set of paths to look for the config file in
	configPaths []string

	// The filesystem to read config from.
	fs afero.Fs

	// Name of file to look for inside the path
	//configName        string
	//configFile        string
	//configType        string
	//envPrefix         string
	requiredFiles     []string
	optionalFiles     []string
	configPermissions os.FileMode

	//automaticEnvApplied bool
	//envKeyReplacer      StringReplacer
	//allowEmptyEnv       bool

	override map[string]interface{} // 1
	config   map[string]interface{} // 2
	defaults map[string]interface{} // 3
	aliases  map[string]string
	//pflags   		 map[string]FlagValue
	//env            map[string]string
	//kvstore        map[string]interface{}
	//typeByDefValue bool

	// Store read properties on the object so that we can write back in order with comments.
	// This will only be used if the configuration read is a properties file.
	//properties *properties.Properties

	//onConfigChange func(fsnotify.Event)

	logger
}

// New returns an initialized Qconf instance.
func New() *Qconf {
	qcfg := new(Qconf)
	qcfg.keyDelim = "."
	//qcfg.configName = "config"
	qcfg.configPermissions = os.FileMode(0644)
	qcfg.fs = afero.NewOsFs()
	qcfg.config = make(map[string]interface{})
	qcfg.override = make(map[string]interface{})
	qcfg.defaults = make(map[string]interface{})
	//qcfg.kvstore = make(map[string]interface{})
	//qcfg.pflags = make(map[string]FlagValue)
	//qcfg.env = make(map[string]string)
	qcfg.aliases = make(map[string]string)
	//qcfg.typeByDefValue = false
	//qcfg.logger = qlog.GetLogger()

	return qcfg
}

// Option configures Qconf using the functional options paradigm popularized by Rob Pike and Dave Cheney.
// If you're unfamiliar with this style,
// see https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
type Option interface {
	apply(q *Qconf)
}

type optionFunc func(q *Qconf)

func (fn optionFunc) apply(q *Qconf) {
	fn(q)
}

// KeyDelimiter sets the delimiter used for determining key parts.
// By default it's value is ".".
func KeyDelimiter(d string) Option {
	return optionFunc(func(q *Qconf) {
		q.keyDelim = d
	})
}

// NewWithOptions creates a new Qconf instance.
func NewWithOptions(opts ...Option) *Qconf {
	q := New()

	for _, opt := range opts {
		opt.apply(q)
	}

	return q
}

// Reset is intended for testing, will reset all to default settings.
// In the public interface for the viper package so applications
// can use it in their testing as well.
func Reset() {
	q = New()
	SupportedExts = []string{"json", "toml", "yaml", "yml"}
}

// SupportedExts are universally supported extensions.
var SupportedExts = []string{"json", "toml", "yaml", "yml"}

func AddRequiredFile(in ...string) { q.AddRequiredFile(in...) }
func (q *Qconf) AddRequiredFile(in ...string) {
	if len(in) > 0 {
		q.requiredFiles = append(q.requiredFiles, in...)
	}
}

func AddOptionalFile(in ...string) { q.AddOptionalFile(in...) }
func (q *Qconf) AddOptionalFile(in ...string) {
	if len(in) > 0 {
		q.optionalFiles = append(q.optionalFiles, in...)
	}
}

// searchMap recursively searches for a value for path in source map.
// Returns nil if not found.
// Note: This assumes that the path entries and map keys are lower cased.
func (q *Qconf) searchMap(source map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	next, ok := source[path[0]]
	if ok {
		// Fast path
		if len(path) == 1 {
			return next
		}

		// Nested case
		switch next.(type) {
		case map[interface{}]interface{}:
			return q.searchMap(cast.ToStringMap(next), path[1:])
		case map[string]interface{}:
			// Type assertion is safe here since it is only reached
			// if the type of `next` is the same as the type being asserted
			return q.searchMap(next.(map[string]interface{}), path[1:])
		default:
			// got a value but nested key expected, return "nil" for not found
			return nil
		}
	}
	return nil
}

// searchMapWithPathPrefixes recursively searches for a value for path in source map.
//
// While searchMap() considers each path element as a single map key, this
// function searches for, and prioritizes, merged path elements.
// e.g., if in the source, "foo" is defined with a sub-key "bar", and "foo.bar"
// is also defined, this latter value is returned for path ["foo", "bar"].
//
// This should be useful only at config level (other maps may not contain dots
// in their keys).
//
// Note: This assumes that the path entries and map keys are lower cased.
func (q *Qconf) searchMapWithPathPrefixes(source map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	// search for path prefixes, starting from the longest one
	for i := len(path); i > 0; i-- {
		prefixKey := strings.ToLower(strings.Join(path[0:i], q.keyDelim))

		next, ok := source[prefixKey]
		if ok {
			// Fast path
			if i == len(path) {
				return next
			}

			// Nested case
			var val interface{}
			switch next.(type) {
			case map[interface{}]interface{}:
				val = q.searchMapWithPathPrefixes(cast.ToStringMap(next), path[i:])
			case map[string]interface{}:
				// Type assertion is safe here since it is only reached
				// if the type of `next` is the same as the type being asserted
				val = q.searchMapWithPathPrefixes(next.(map[string]interface{}), path[i:])
			default:
				// got a value but nested key expected, do nothing and look for next prefix
			}
			if val != nil {
				return val
			}
		}
	}

	// not found
	return nil
}

// isPathShadowedInDeepMap makes sure the given path is not shadowed somewhere
// on its path in the map.
// e.g., if "foo.bar" has a value in the given map, it “shadows”
//       "foo.bar.baz" in a lower-priority map
func (q *Qconf) isPathShadowedInDeepMap(path []string, m map[string]interface{}) string {
	var parentVal interface{}
	for i := 1; i < len(path); i++ {
		parentVal = q.searchMap(m, path[0:i])
		if parentVal == nil {
			// not found, no need to add more path elements
			return ""
		}
		switch parentVal.(type) {
		case map[interface{}]interface{}:
			continue
		case map[string]interface{}:
			continue
		default:
			// parentVal is a regular value which shadows "path"
			return strings.Join(path[0:i], q.keyDelim)
		}
	}
	return ""
}

// isPathShadowedInFlatMap makes sure the given path is not shadowed somewhere
// in a sub-path of the map.
// e.g., if "foo.bar" has a value in the given map, it “shadows”
//       "foo.bar.baz" in a lower-priority map
func (q *Qconf) isPathShadowedInFlatMap(path []string, mi interface{}) string {
	// unify input map
	var m map[string]interface{}
	switch mi.(type) {
	case map[string]string, map[string]FlagValue:
		m = cast.ToStringMap(mi)
	default:
		return ""
	}

	// scan paths
	var parentKey string
	for i := 1; i < len(path); i++ {
		parentKey = strings.Join(path[0:i], q.keyDelim)
		if _, ok := m[parentKey]; ok {
			return parentKey
		}
	}
	return ""
}

// SetTypeByDefaultValue enables or disables the inference of a key value's
// type when the Get function is used based upon a key's default value as
// opposed to the value returned based on the normal fetch logic.
//
// For example, if a key has a default value of []string{} and the same key
// is set via an environment variable to "a b q", a call to the Get function
// would return a string slice for the key if the key's type is inferred by
// the default value and the Get function would return:
//
//   []string {"a", "b", "q"}
//
// Otherwise the Get function would return:
//
//   "a b q"
//func SetTypeByDefaultValue(enable bool) { q.SetTypeByDefaultValue(enable) }
//func (q *Qconf) SetTypeByDefaultValue(enable bool) {
//	q.typeByDefValue = enable
//}

// GetViper gets the global Qconf instance.
func GetQconf() *Qconf {
	return q
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Qconf will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) interface{} { return q.Get(key) }
func (q *Qconf) Get(key string) interface{} {
	lcaseKey := strings.ToLower(key)
	val := q.find(lcaseKey, true)
	if val == nil {
		return nil
	}

	//if q.typeByDefValue {
	//	// TODO(bep) this branch isn't covered by a single test.
	//	valType := val
	//	path := strings.Split(lcaseKey, q.keyDelim)
	//	defVal := q.searchMap(q.defaults, path)
	//	if defVal != nil {
	//		valType = defVal
	//	}
	//
	//	switch valType.(type) {
	//	case bool:
	//		return cast.ToBool(val)
	//	case string:
	//		return cast.ToString(val)
	//	case int32, int16, int8, int:
	//		return cast.ToInt(val)
	//	case uint:
	//		return cast.ToUint(val)
	//	case uint32:
	//		return cast.ToUint32(val)
	//	case uint64:
	//		return cast.ToUint64(val)
	//	case int64:
	//		return cast.ToInt64(val)
	//	case float64, float32:
	//		return cast.ToFloat64(val)
	//	case time.Time:
	//		return cast.ToTime(val)
	//	case time.Duration:
	//		return cast.ToDuration(val)
	//	case []string:
	//		return cast.ToStringSlice(val)
	//	case []int:
	//		return cast.ToIntSlice(val)
	//	}
	//}

	return val
}

// Sub returns new Qconf instance representing a sub tree of this instance.
// Sub is case-insensitive for a key.
func Sub(key string) *Qconf { return q.Sub(key) }
func (q *Qconf) Sub(key string) *Qconf {
	subv := New()
	data := q.Get(key)
	if data == nil {
		return nil
	}

	if reflect.TypeOf(data).Kind() == reflect.Map {
		subv.config = cast.ToStringMap(data)
		return subv
	}
	return nil
}

// GetString returns the value associated with the key as a string.
func GetString(key string, defaultValue ...string) string { return q.GetString(key, defaultValue...) }
func (q *Qconf) GetString(key string, defaultValue ...string) (result string) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToStringE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string, defaultValue ...bool) bool { return q.GetBool(key, defaultValue...) }
func (q *Qconf) GetBool(key string, defaultValue ...bool) (result bool) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToBoolE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string, defaultValue ...int) int { return q.GetInt(key, defaultValue...) }
func (q *Qconf) GetInt(key string, defaultValue ...int) (result int) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToIntE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetInt32 returns the value associated with the key as an integer.
func GetInt32(key string, defaultValue ...int32) int32 { return q.GetInt32(key, defaultValue...) }
func (q *Qconf) GetInt32(key string, defaultValue ...int32) (result int32) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToInt32E(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func GetInt64(key string, defaultValue ...int64) int64 { return q.GetInt64(key, defaultValue...) }
func (q *Qconf) GetInt64(key string, defaultValue ...int64) (result int64) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToInt64E(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetUint returns the value associated with the key as an unsigned integer.
func GetUint(key string, defaultValue ...uint) uint { return q.GetUint(key, defaultValue...) }
func (q *Qconf) GetUint(key string, defaultValue ...uint) (result uint) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToUintE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetUint32 returns the value associated with the key as an unsigned integer.
func GetUint32(key string, defaultValue ...uint32) uint32 { return q.GetUint32(key, defaultValue...) }
func (q *Qconf) GetUint32(key string, defaultValue ...uint32) (result uint32) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToUint32E(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64(key string, defaultValue ...uint64) uint64 { return q.GetUint64(key, defaultValue...) }
func (q *Qconf) GetUint64(key string, defaultValue ...uint64) (result uint64) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToUint64E(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string, defaultValue ...float64) float64 {
	return q.GetFloat64(key, defaultValue...)
}
func (q *Qconf) GetFloat64(key string, defaultValue ...float64) (result float64) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToFloat64E(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetTime returns the value associated with the key as time.
func GetTime(key string, defaultValue ...time.Time) time.Time { return q.GetTime(key, defaultValue...) }
func (q *Qconf) GetTime(key string, defaultValue ...time.Time) (result time.Time) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToTimeE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetDuration returns the value associated with the key as a duration.
func GetDuration(key string, defaultValue ...time.Duration) time.Duration {
	return q.GetDuration(key, defaultValue...)
}
func (q *Qconf) GetDuration(key string, defaultValue ...time.Duration) (result time.Duration) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToDurationE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func GetSlice(key string, defaultValue ...[]interface{}) []interface{} {
	return q.GetSlice(key, defaultValue...)
}
func (q *Qconf) GetSlice(key string, defaultValue ...[]interface{}) (result []interface{}) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToSliceE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func GetIntSlice(key string, defaultValue ...[]int) []int { return q.GetIntSlice(key, defaultValue...) }
func (q *Qconf) GetIntSlice(key string, defaultValue ...[]int) (result []int) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToIntSliceE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(key string, defaultValue ...[]string) []string {
	return q.GetStringSlice(key, defaultValue...)
}
func (q *Qconf) GetStringSlice(key string, defaultValue ...[]string) (result []string) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToStringSliceE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func GetStringMap(key string, defaultValue ...map[string]interface{}) map[string]interface{} {
	return q.GetStringMap(key, defaultValue...)
}
func (q *Qconf) GetStringMap(key string, defaultValue ...map[string]interface{}) (result map[string]interface{}) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToStringMapE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetStringMapString returns the value associated with the key as a map of strings.
func GetStringMapString(key string, defaultValue ...map[string]string) map[string]string {
	return q.GetStringMapString(key, defaultValue...)
}
func (q *Qconf) GetStringMapString(key string, defaultValue ...map[string]string) (result map[string]string) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToStringMapStringE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func GetStringMapStringSlice(key string, defaultValue ...map[string][]string) map[string][]string {
	return q.GetStringMapStringSlice(key, defaultValue...)
}
func (q *Qconf) GetStringMapStringSlice(key string, defaultValue ...map[string][]string) (result map[string][]string) {
	value := q.Get(key)
	if value != nil {
		valueE, err := cast.ToStringMapStringSliceE(value)
		if err == nil {
			return valueE
		}
		result = valueE
	}
	// 不存在或解析失败
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return
}

// UnmarshalKey takes a single key and unmarshals it into a Struct.
func UnmarshalKey(key string, rawVal interface{}, opts ...DecoderConfigOption) error {
	return q.UnmarshalKey(key, rawVal, opts...)
}
func (q *Qconf) UnmarshalKey(key string, rawVal interface{}, opts ...DecoderConfigOption) error {
	err := decode(q.Get(key), defaultDecoderConfig(rawVal, opts...))

	if err != nil {
		return err
	}

	return nil
}

// Unmarshal unmarshals the config into a Struct. Make sure that the tags
// on the fields of the structure are properly set.
func Unmarshal(rawVal interface{}, opts ...DecoderConfigOption) error {
	return q.Unmarshal(rawVal, opts...)
}
func (q *Qconf) Unmarshal(rawVal interface{}, opts ...DecoderConfigOption) error {
	err := decode(q.AllSettings(), defaultDecoderConfig(rawVal, opts...))

	if err != nil {
		return err
	}

	return nil
}

// defaultDecoderConfig returns default mapsstructure.DecoderConfig with suppot
// of time.Duration values & string slices
func defaultDecoderConfig(output interface{}, opts ...DecoderConfigOption) *mapstructure.DecoderConfig {
	q := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

// A wrapper around mapstructure.Decode that mimics the WeakDecode functionality
func decode(input interface{}, config *mapstructure.DecoderConfig) error {
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

// UnmarshalExact unmarshals the config into a Struct, erroring if a field is nonexistent
// in the destination struct.
func UnmarshalExact(rawVal interface{}, opts ...DecoderConfigOption) error {
	return q.UnmarshalExact(rawVal, opts...)
}
func (q *Qconf) UnmarshalExact(rawVal interface{}, opts ...DecoderConfigOption) error {
	config := defaultDecoderConfig(rawVal, opts...)
	config.ErrorUnused = true

	err := decode(q.AllSettings(), config)

	if err != nil {
		return err
	}

	return nil
}

// Given a key, find the value.
//
// Qconf will check to see if an alias exists first.
// Qconf will then check in the following order:
// flag, env, config file, key/value store.
// Lastly, if no value was found and flagDefault is true, and if the key
// corresponds to a flag, the flag's default value is returned.
//
// Note: this assumes a lower-cased key given.
func (q *Qconf) find(lcaseKey string, flagDefault bool) interface{} {
	var (
		val    interface{}
		path   = strings.Split(lcaseKey, q.keyDelim)
		nested = len(path) > 1
		//exists bool
	)

	// compute the path through the nested maps to the nested value
	if nested && q.isPathShadowedInDeepMap(path, castMapStringToMapInterface(q.aliases)) != "" {
		return nil
	}

	// if the requested key is an alias, then return the proper key
	lcaseKey = q.realKey(lcaseKey)
	path = strings.Split(lcaseKey, q.keyDelim)
	nested = len(path) > 1

	// Set() override first
	val = q.searchMap(q.override, path)
	if val != nil {
		return val
	}
	if nested && q.isPathShadowedInDeepMap(path, q.override) != "" {
		return nil
	}

	// PFlag override next
	//flag, exists := q.pflags[lcaseKey]
	//if exists && flag.HasChanged() {
	//	switch flag.ValueType() {
	//	case "int", "int8", "int16", "int32", "int64":
	//		return cast.ToInt(flag.ValueString())
	//	case "bool":
	//		return cast.ToBool(flag.ValueString())
	//	case "stringSlice":
	//		s := strings.TrimPrefix(flag.ValueString(), "[")
	//		s = strings.TrimSuffix(s, "]")
	//		res, _ := readAsCSV(s)
	//		return res
	//	case "intSlice":
	//		s := strings.TrimPrefix(flag.ValueString(), "[")
	//		s = strings.TrimSuffix(s, "]")
	//		res, _ := readAsCSV(s)
	//		return cast.ToIntSlice(res)
	//	default:
	//		return flag.ValueString()
	//	}
	//}
	//if nested && q.isPathShadowedInFlatMap(path, q.pflags) != "" {
	//	return nil
	//}

	// Env override next
	//if q.automaticEnvApplied {
	//	// even if it hasn't been registered, if automaticEnv is used,
	//	// check any Get request
	//	if val, ok := q.getEnv(q.mergeWithEnvPrefix(lcaseKey)); ok {
	//		return val
	//	}
	//	if nested && q.isPathShadowedInAutoEnv(path) != "" {
	//		return nil
	//	}
	//}
	//envkey, exists := q.env[lcaseKey]
	//if exists {
	//	if val, ok := q.getEnv(envkey); ok {
	//		return val
	//	}
	//}
	//if nested && q.isPathShadowedInFlatMap(path, q.env) != "" {
	//	return nil
	//}

	// Config file next
	val = q.searchMapWithPathPrefixes(q.config, path)
	if val != nil {
		return val
	}
	if nested && q.isPathShadowedInDeepMap(path, q.config) != "" {
		return nil
	}

	// K/V store next
	//val = q.searchMap(q.kvstore, path)
	//if val != nil {
	//	return val
	//}
	//if nested && q.isPathShadowedInDeepMap(path, q.kvstore) != "" {
	//	return nil
	//}
	//
	// Default next
	val = q.searchMap(q.defaults, path)
	if val != nil {
		return val
	}
	if nested && q.isPathShadowedInDeepMap(path, q.defaults) != "" {
		return nil
	}

	//if flagDefault {
	//	// last chance: if no value is found and a flag does exist for the key,
	//	// get the flag's default value even if the flag's value has not been set.
	//	if flag, exists := q.pflags[lcaseKey]; exists {
	//		switch flag.ValueType() {
	//		case "int", "int8", "int16", "int32", "int64":
	//			return cast.ToInt(flag.ValueString())
	//		case "bool":
	//			return cast.ToBool(flag.ValueString())
	//		case "stringSlice":
	//			s := strings.TrimPrefix(flag.ValueString(), "[")
	//			s = strings.TrimSuffix(s, "]")
	//			res, _ := readAsCSV(s)
	//			return res
	//		case "intSlice":
	//			s := strings.TrimPrefix(flag.ValueString(), "[")
	//			s = strings.TrimSuffix(s, "]")
	//			res, _ := readAsCSV(s)
	//			return cast.ToIntSlice(res)
	//		default:
	//			return flag.ValueString()
	//		}
	//	}
	//	// last item, no need to check shadowing
	//}

	return nil
}

//func readAsCSV(val string) ([]string, error) {
//	if val == "" {
//		return []string{}, nil
//	}
//	stringReader := strings.NewReader(val)
//	csvReader := csv.NewReader(stringReader)
//	return csvReader.Read()
//}

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key.
func IsSet(key string) bool { return q.IsSet(key) }
func (q *Qconf) IsSet(key string) bool {
	lcaseKey := strings.ToLower(key)
	val := q.find(lcaseKey, false)
	return val != nil
}

// RegisterAlias creates an alias that provides another accessor for the same key.
// This enables one to change a name without breaking the application.
func RegisterAlias(alias string, key string) { q.RegisterAlias(alias, key) }
func (q *Qconf) RegisterAlias(alias string, key string) {
	q.registerAlias(alias, strings.ToLower(key))
}

func (q *Qconf) registerAlias(alias string, key string) {
	alias = strings.ToLower(alias)
	if alias != key && alias != q.realKey(key) {
		_, exists := q.aliases[alias]

		if !exists {
			// if we alias something that exists in one of the maps to another
			// name, we'll never be able to get that value using the original
			// name, so move the config value to the new realkey.
			if val, ok := q.config[alias]; ok {
				delete(q.config, alias)
				q.config[key] = val
			}
			//if val, ok := q.kvstore[alias]; ok {
			//	delete(q.kvstore, alias)
			//	q.kvstore[key] = val
			//}
			if val, ok := q.defaults[alias]; ok {
				delete(q.defaults, alias)
				q.defaults[key] = val
			}
			if val, ok := q.override[alias]; ok {
				delete(q.override, alias)
				q.override[key] = val
			}
			q.aliases[alias] = key
		}
	} else {
		jww.WARN.Println("Creating circular reference alias", alias, key, q.realKey(key))
	}
}

func (q *Qconf) realKey(key string) string {
	newkey, exists := q.aliases[key]
	if exists {
		jww.DEBUG.Println("Alias", key, "to", newkey)
		return q.realKey(newkey)
	}
	return key
}

// InConfig checks to see if the given key (or an alias) is in the config file.
func InConfig(key string) bool { return q.InConfig(key) }
func (q *Qconf) InConfig(key string) bool {
	// if the requested key is an alias, then return the proper key
	key = q.realKey(key)

	_, exists := q.config[key]
	return exists
}

// SetDefault sets the default value for this key.
// SetDefault is case-insensitive for a key.
// Default only used when no value is provided by the user via flag, config or ENV.
func SetDefault(key string, value interface{}) { q.SetDefault(key, value) }
func (q *Qconf) SetDefault(key string, value interface{}) {
	// If alias passed in, then set the proper default
	key = q.realKey(strings.ToLower(key))
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, q.keyDelim)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(q.defaults, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

// Set sets the value for the key in the override register.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
func Set(key string, value interface{}) { q.Set(key, value) }
func (q *Qconf) Set(key string, value interface{}) {
	// If alias passed in, then set the proper override
	key = q.realKey(strings.ToLower(key))
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, q.keyDelim)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(q.override, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

// ReadInConfig will discover and load the configuration file from disk
// and key/value stores, searching in one of the defined paths.
func ReadInConfig() error { return q.ReadInConfig() }
func (q *Qconf) ReadInConfig() error {
	jww.INFO.Println("Attempting to read in config file")

	readfn := func(filename string, isRequired bool) error {
		exist, err := afero.Exists(q.fs, filename)
		if err != nil {
			return err
		}
		if !exist {
			if isRequired {
				return ConfigFileNotFoundError{filepath.Base(filename), filepath.Dir(filename)}
			}
			return nil
		}

		if err := q.MergeConfig(filename); err != nil {
			return err
		}
		return nil
	}

	for _, filename := range q.requiredFiles { // 必须要存在
		if err := readfn(filename, true); err != nil {
			return err
		}
	}

	for _, filename := range q.optionalFiles {
		if err := readfn(filename, false); err != nil {
			return err
		}
	}

	return nil
}

// MergeConfig merges a new configuration with an existing config.
func MergeConfig(filename string) error { return q.MergeConfig(filename) }
func (q *Qconf) MergeConfig(filename string) error {
	cfg := make(map[string]interface{})
	configType, err := getConfigType(filename)
	if err != nil {
		return err
	}
	file, err := afero.ReadFile(q.fs, filename)
	if err != nil {
		return err
	}
	if err := q.unmarshalReader(bytes.NewReader(file), cfg, configType); err != nil {
		return err
	}
	return q.MergeConfigMap(cfg)
}

// MergeConfigMap merges the configuration from the map given with an existing config.
// Note that the map given may be modified.
func MergeConfigMap(cfg map[string]interface{}) error { return q.MergeConfigMap(cfg) }
func (q *Qconf) MergeConfigMap(cfg map[string]interface{}) error {
	if q.config == nil {
		q.config = make(map[string]interface{})
	}
	insensitiviseMap(cfg)
	mergeMaps(cfg, q.config, nil)
	return nil
}

// WriteConfig writes the current configuration to a file.
//func WriteConfig() error { return q.WriteConfig() }
//func (q *Qconf) WriteConfig() error {
//	filename, err := q.getConfigFile()
//	if err != nil {
//		return err
//	}
//	return q.writeConfig(filename, true)
//}

// SafeWriteConfig writes current configuration to file only if the file does not exist.
//func SafeWriteConfig() error { return q.SafeWriteConfig() }
//func (q *Qconf) SafeWriteConfig() error {
//	if len(q.configPaths) < 1 {
//		return errors.New("missing configuration for 'configPath'")
//	}
//	return q.SafeWriteConfigAs(filepath.Join(q.configPaths[0], q.configName+"."+q.configType))
//}

// WriteConfigAs writes current configuration to a given filename.
func WriteConfigAs(filename string) error { return q.WriteConfigAs(filename) }
func (q *Qconf) WriteConfigAs(filename string) error {
	return q.writeConfig(filename, true)
}

// SafeWriteConfigAs writes current configuration to a given filename if it does not exist.
func SafeWriteConfigAs(filename string) error { return q.SafeWriteConfigAs(filename) }
func (q *Qconf) SafeWriteConfigAs(filename string) error {
	alreadyExists, err := afero.Exists(q.fs, filename)
	if alreadyExists && err == nil {
		return ConfigFileAlreadyExistsError(filename)
	}
	return q.writeConfig(filename, false)
}

func (q *Qconf) writeConfig(filename string, force bool) error {
	jww.INFO.Println("Attempting to write configuration to file.")
	ext := filepath.Ext(filename)
	if len(ext) <= 1 {
		return fmt.Errorf("filename: %s requires valid extension", filename)
	}
	configType := ext[1:]
	if !stringInSlice(configType, SupportedExts) {
		return UnsupportedConfigError(configType)
	}
	if q.config == nil {
		q.config = make(map[string]interface{})
	}
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	if !force {
		flags |= os.O_EXCL
	}
	f, err := q.fs.OpenFile(filename, flags, q.configPermissions)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := q.marshalWriter(f, configType); err != nil {
		return err
	}

	return f.Sync()
}

// Unmarshal a Reader into a map.
// Should probably be an unexported function.
//func unmarshalReader(in io.Reader, c map[string]interface{}, ext string) error {
//	return q.unmarshalReader(in, c, ext)
//}
func (q *Qconf) unmarshalReader(in io.Reader, c map[string]interface{}, ext string) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	switch strings.ToLower(ext) {
	case "yaml", "yml":
		if err := yaml.Unmarshal(buf.Bytes(), &c); err != nil {
			return ConfigParseError{err}
		}

	case "json":
		if err := json.Unmarshal(buf.Bytes(), &c); err != nil {
			return ConfigParseError{err}
		}

	case "toml":
		tree, err := toml.LoadReader(buf)
		if err != nil {
			return ConfigParseError{err}
		}
		tmap := tree.ToMap()
		for k, v := range tmap {
			c[k] = v
		}

		//case "dotenv", "env":
		//	env, err := gotenv.StrictParse(buf)
		//	if err != nil {
		//		return ConfigParseError{err}
		//	}
		//	for k, v := range env {
		//		c[k] = v
		//	}

		//case "properties", "props", "prop":
		//	q.properties = properties.NewProperties()
		//	var err error
		//	if q.properties, err = properties.Load(buf.Bytes(), properties.UTF8); err != nil {
		//		return ConfigParseError{err}
		//	}
		//	for _, key := range q.properties.Keys() {
		//		value, _ := q.properties.Get(key)
		//		// recursively build nested maps
		//		path := strings.Split(key, ".")
		//		lastKey := strings.ToLower(path[len(path)-1])
		//		deepestMap := deepSearch(c, path[0:len(path)-1])
		//		// set innermost value
		//		deepestMap[lastKey] = value
		//	}

	}

	insensitiviseMap(c)
	return nil
}

// Marshal a map into Writer.
func (q *Qconf) marshalWriter(f afero.File, configType string) error {
	c := q.AllSettings()
	switch configType {
	case "json":
		b, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			return ConfigMarshalError{err}
		}
		_, err = f.WriteString(string(b))
		if err != nil {
			return ConfigMarshalError{err}
		}

	//case "prop", "props", "properties":
	//	if q.properties == nil {
	//		q.properties = properties.NewProperties()
	//	}
	//	p := q.properties
	//	for _, key := range q.AllKeys() {
	//		_, _, err := p.Set(key, q.GetString(key))
	//		if err != nil {
	//			return ConfigMarshalError{err}
	//		}
	//	}
	//	_, err := p.WriteComment(f, "#", properties.UTF8)
	//	if err != nil {
	//		return ConfigMarshalError{err}
	//	}

	//case "dotenv", "env":
	//	lines := []string{}
	//	for _, key := range q.AllKeys() {
	//		envName := strings.ToUpper(strings.Replace(key, ".", "_", -1))
	//		val := q.Get(key)
	//		lines = append(lines, fmt.Sprintf("%v=%v", envName, val))
	//	}
	//	s := strings.Join(lines, "\n")
	//	if _, err := f.WriteString(s); err != nil {
	//		return ConfigMarshalError{err}
	//	}

	case "toml":
		t, err := toml.TreeFromMap(c)
		if err != nil {
			return ConfigMarshalError{err}
		}
		s := t.String()
		if _, err := f.WriteString(s); err != nil {
			return ConfigMarshalError{err}
		}

	case "yaml", "yml":
		b, err := yaml.Marshal(c)
		if err != nil {
			return ConfigMarshalError{err}
		}
		if _, err = f.WriteString(string(b)); err != nil {
			return ConfigMarshalError{err}
		}

	}
	return nil
}

func keyExists(k string, m map[string]interface{}) string {
	lk := strings.ToLower(k)
	for mk := range m {
		lmk := strings.ToLower(mk)
		if lmk == lk {
			return mk
		}
	}
	return ""
}

func castToMapStringInterface(
	src map[interface{}]interface{}) map[string]interface{} {
	tgt := map[string]interface{}{}
	for k, q := range src {
		tgt[fmt.Sprintf("%q", k)] = q
	}
	return tgt
}

func castMapStringToMapInterface(src map[string]string) map[string]interface{} {
	tgt := map[string]interface{}{}
	for k, q := range src {
		tgt[k] = q
	}
	return tgt
}

//func castMapFlagToMapInterface(src map[string]FlagValue) map[string]interface{} {
//	tgt := map[string]interface{}{}
//	for k, q := range src {
//		tgt[k] = q
//	}
//	return tgt
//}

// mergeMaps merges two maps. The `itgt` parameter is for handling go-yaml's
// insistence on parsing nested structures as `map[interface{}]interface{}`
// instead of using a `string` as the key for nest structures beyond one level
// deep. Both map types are supported as there is a go-yaml fork that uses
// `map[string]interface{}` instead.
func mergeMaps(
	src, tgt map[string]interface{}, itgt map[interface{}]interface{}) {
	for sk, sv := range src {
		tk := keyExists(sk, tgt)
		if tk == "" {
			jww.TRACE.Printf("tk=\"\", tgt[%s]=%q", sk, sv)
			tgt[sk] = sv
			if itgt != nil {
				itgt[sk] = sv
			}
			continue
		}

		tv, ok := tgt[tk]
		if !ok {
			jww.TRACE.Printf("tgt[%s] != ok, tgt[%s]=%q", tk, sk, sv)
			tgt[sk] = sv
			if itgt != nil {
				itgt[sk] = sv
			}
			continue
		}

		svType := reflect.TypeOf(sv)
		tvType := reflect.TypeOf(tv)
		if svType != tvType {
			jww.ERROR.Printf(
				"svType != tvType; key=%s, st=%q, tt=%q, sv=%q, tv=%q",
				sk, svType, tvType, sv, tv)
			continue
		}

		jww.TRACE.Printf("processing key=%s, st=%q, tt=%q, sv=%q, tv=%q",
			sk, svType, tvType, sv, tv)

		switch ttv := tv.(type) {
		case map[interface{}]interface{}:
			jww.TRACE.Printf("merging maps (must convert)")
			tsv := sv.(map[interface{}]interface{})
			ssv := castToMapStringInterface(tsv)
			stv := castToMapStringInterface(ttv)
			mergeMaps(ssv, stv, ttv)
		case map[string]interface{}:
			jww.TRACE.Printf("merging maps")
			mergeMaps(sv.(map[string]interface{}), ttv, nil)
		default:
			jww.TRACE.Printf("setting value")
			tgt[tk] = sv
			if itgt != nil {
				itgt[tk] = sv
			}
		}
	}
}

// AllKeys returns all keys holding a value, regardless of where they are set.
// Nested keys are returned with a q.keyDelim separator
func AllKeys() []string { return q.AllKeys() }
func (q *Qconf) AllKeys() []string {
	m := map[string]bool{}
	// add all paths, by order of descending priority to ensure correct shadowing
	m = q.flattenAndMergeMap(m, castMapStringToMapInterface(q.aliases), "")
	m = q.flattenAndMergeMap(m, q.override, "")
	//m = q.mergeFlatMap(m, castMapFlagToMapInterface(q.pflags))
	//m = q.mergeFlatMap(m, castMapStringToMapInterface(q.env))
	m = q.flattenAndMergeMap(m, q.config, "")
	//m = q.flattenAndMergeMap(m, q.kvstore, "")
	m = q.flattenAndMergeMap(m, q.defaults, "")

	// convert set of paths to list
	a := make([]string, 0, len(m))
	for x := range m {
		a = append(a, x)
	}
	return a
}

// flattenAndMergeMap recursively flattens the given map into a map[string]bool
// of key paths (used as a set, easier to manipulate than a []string):
// - each path is merged into a single key string, delimited with q.keyDelim
// - if a path is shadowed by an earlier value in the initial shadow map,
//   it is skipped.
// The resulting set of paths is merged to the given shadow set at the same time.
func (q *Qconf) flattenAndMergeMap(shadow map[string]bool, m map[string]interface{}, prefix string) map[string]bool {
	if shadow != nil && prefix != "" && shadow[prefix] {
		// prefix is shadowed => nothing more to flatten
		return shadow
	}
	if shadow == nil {
		shadow = make(map[string]bool)
	}

	var m2 map[string]interface{}
	if prefix != "" {
		prefix += q.keyDelim
	}
	for k, val := range m {
		fullKey := prefix + k
		switch val.(type) {
		case map[string]interface{}:
			m2 = val.(map[string]interface{})
		case map[interface{}]interface{}:
			m2 = cast.ToStringMap(val)
		default:
			// immediate value
			shadow[strings.ToLower(fullKey)] = true
			continue
		}
		// recursively merge to shadow map
		shadow = q.flattenAndMergeMap(shadow, m2, fullKey)
	}
	return shadow
}

// mergeFlatMap merges the given maps, excluding values of the second map
// shadowed by values from the first map.
func (q *Qconf) mergeFlatMap(shadow map[string]bool, m map[string]interface{}) map[string]bool {
	// scan keys
outer:
	for k := range m {
		path := strings.Split(k, q.keyDelim)
		// scan intermediate paths
		var parentKey string
		for i := 1; i < len(path); i++ {
			parentKey = strings.Join(path[0:i], q.keyDelim)
			if shadow[parentKey] {
				// path is shadowed, continue
				continue outer
			}
		}
		// add key
		shadow[strings.ToLower(k)] = true
	}
	return shadow
}

// AllSettings merges all settings and returns them as a map[string]interface{}.
func AllSettings() map[string]interface{} { return q.AllSettings() }
func (q *Qconf) AllSettings() map[string]interface{} {
	m := map[string]interface{}{}
	// start from the list of keys, and construct the map one value at a time
	for _, k := range q.AllKeys() {
		value := q.Get(k)
		if value == nil {
			// should not happen, since AllKeys() returns only keys holding a value,
			// check just in case anything changes
			continue
		}
		path := strings.Split(k, q.keyDelim)
		lastKey := strings.ToLower(path[len(path)-1])
		deepestMap := deepSearch(m, path[0:len(path)-1])
		// set innermost value
		deepestMap[lastKey] = value
	}
	return m
}

// SetFs sets the filesystem to use to read configuration.
func SetFs(fs afero.Fs) { q.SetFs(fs) }
func (q *Qconf) SetFs(fs afero.Fs) {
	q.fs = fs
}

// SetConfigPermissions sets the permissions for the config file.
func SetConfigPermissions(perm os.FileMode) { q.SetConfigPermissions(perm) }
func (q *Qconf) SetConfigPermissions(perm os.FileMode) {
	q.configPermissions = perm.Perm()
}

func getConfigType(filename string) (string, error) {
	ext := filepath.Ext(filename)
	if len(ext) <= 1 {
		return "", fmt.Errorf("filename: %s requires valid extension", filename)
	}
	configType := ext[1:]
	if !stringInSlice(configType, SupportedExts) {
		return "", UnsupportedConfigError(configType)
	}
	return configType, nil
}

// Debug prints all configuration registries for debugging
// purposes.
func Debug() { q.Debug() }
func (q *Qconf) Debug() {
	fmt.Printf("Aliases:\n%#q\n", q.aliases)
	fmt.Printf("Override:\n%#q\n", q.override)
	//fmt.Printf("PFlags:\n%#q\n", q.pflags)
	//fmt.Printf("Env:\n%#q\n", q.env)
	//fmt.Printf("Key/Value Store:\n%#q\n", q.kvstore)
	fmt.Printf("Config:\n%#q\n", q.config)
	fmt.Printf("Defaults:\n%#q\n", q.defaults)
}
