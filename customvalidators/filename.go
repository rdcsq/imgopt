package customvalidators

import (
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

func Filename(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() != reflect.String {
		return false
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9-_./]+$", field.String())
	return matched
}
