package cc

import (
	"errors"
	"fmt"
)

var (
	ErrEnvEnumValue = errors.New("无效的环境枚举值，有效的枚举值为：qa, pre, prd")
	ErrBackup       = errors.New("数据备份失败")
)

type ErrChecksum struct {
	checksum    string
	calCheckSum string
}

func (e ErrChecksum) Error() string {
	return fmt.Sprintf("checksum 失败，计算的checksum: %s与传入的checksum: %s 不一致", e.calCheckSum, e.checksum)
}
