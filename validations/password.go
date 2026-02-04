package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/minisource/go-common/common"
)

type ValidatorPasswordConfig struct {
	common.PasswordConfig
}

func (cfg ValidatorPasswordConfig) PasswordValidator(fld validator.FieldLevel) bool {
	value, ok := fld.Field().Interface().(string)
	if !ok {
		fld.Param()
		return false
	}

	return cfg.CheckPassword(value)
}
