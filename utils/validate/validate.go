package validate

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/zh"

	zhTranslations "gopkg.in/go-playground/validator.v9/translations/zh"

	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
)

var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate = validator.New()
)

func init() {
	// register Zh translations
	zht := zh.New()
	uni = ut.New(zht, zht)
	trans, _ = uni.GetTranslator("zh")
	err := zhTranslations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		panic(err)
	}
	// 注册tagNameFunc 用于获取validator 的json字段
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func StructParam(val interface{}) error {
	err := validate.Struct(val)
	if err != nil {
		if errNew, ok := err.(*validator.InvalidValidationError); ok {
			panic(errNew)
		}
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(e.Translate(trans))
		}
		return err
	}

	return err
}
