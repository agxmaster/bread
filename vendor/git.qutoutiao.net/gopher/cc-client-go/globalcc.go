package cc

type registeredCC struct {
	center       *ConfigCenter
	isRegistered bool
}

var (
	globalCC = registeredCC{NewNullConfigCenter("", "", ""), false}
)

// SetGlobalCC 设置全局cc对象
func SetGlobalCC(center *ConfigCenter) {
	globalCC = registeredCC{center, true}
}

func GlobalCC() *ConfigCenter {
	return globalCC.center
}

// IsGlobalCCRegistered 配置中心是否初始化成功
func IsGlobalTracerRegistered() bool {
	return globalCC.isRegistered
}

// GetString 获取某个key的值
func GetString(key, defaultVal string) string {
	return GlobalCC().GetString(key, defaultVal)
}

func GetInt(key string, defaultVal int) int {
	return GlobalCC().GetInt(key, defaultVal)
}

func GetFloat(key string, defaultVal float64) float64 {
	return GlobalCC().GetFloat(key, defaultVal)
}

func GetBool(key string, defaultVal bool) bool {
	return GlobalCC().GetBool(key, defaultVal)
}

func GetAll() map[string]string {
	return GlobalCC().GetAll()
}

func IsKeyExists(key string) bool {
	return GlobalCC().IsKeyExists(key)
}

func Close() error {
	return GlobalCC().Close()
}
