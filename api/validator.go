package api

import (
	"github.com/dubass83/simplebank/util"
	"github.com/go-playground/validator/v10"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	currency, ok := fl.Field().Interface().(string)
	if ok {
		return util.IfSupportedCurrency(currency)
	}
	return false
}
