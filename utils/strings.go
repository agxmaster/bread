package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Join(s interface{}, sep string) string {
	sValue := reflect.ValueOf(s)

	strs := make([]string, 0)
	switch sValue.Type().Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < sValue.Len(); i++ {
			switch sValue.Index(i).Kind() {
			case reflect.Bool:
			case reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32,
				reflect.Int64, reflect.Uint64, reflect.Int, reflect.Uint, reflect.Uintptr:
				strs = append(strs, fmt.Sprintf("%d", sValue.Index(i).Interface()))
			case reflect.Float32, reflect.Float64:
			default:
			}
		}
	}
	return strings.Join(strs, sep)
}

func Split(s string, sep string) []string {
	if len(s) == 0 {
		return []string{}
	}
	return strings.Split(s, sep)
}

func SplitInts(s string, sep string) []int {
	if len(s) == 0 {
		return []int{}
	}
	n := make([]int, 0)
	for _, value := range Split(s, sep) {
		i, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		n = append(n, i)
	}
	return n
}

func SplitUInt64(s string, sep string) []uint64 {
	if len(s) == 0 {
		return []uint64{}
	}
	n := make([]uint64, 0)
	for _, value := range Split(s, sep) {
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			continue
		}
		n = append(n, i)
	}
	return n
}

func IsEmptyString(s *string) bool {
	return s == nil || len(*s) == 0
}
