package valid

import (
	"zkCache/pkg/response"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

type verification struct {
}

var Verification verification

func (this verification) Verify(value ...interface{}) error {
	uni := ut.New(zh.New())
	trans, _ := uni.GetTranslator("zh")
	//实例化验证器
	validate := validator.New()
	// 注册翻译器到校验器
	err := zh_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		panic(err)
	}

	for _, v := range value {
		err = validate.Struct(v)
		var data []string
		if err != nil {
			for _, errMsg := range err.(validator.ValidationErrors) {
				data = append(data, errMsg.Translate(trans))
			}
			return response.NewErr(response.PARAMETER_ERROR)
		}
	}
	return nil
}
