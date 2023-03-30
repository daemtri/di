package validate

import (
	"github.com/go-playground/validator/v10"
)

var Default = validator.New()

// Struct 使用 https://github.com/go-playground/validator 验证结构体
// 结构体添加,如 `validate:"hexcolor|rgb|rgba`
func Struct(v interface{}) error {
	return Default.Struct(v)
}
