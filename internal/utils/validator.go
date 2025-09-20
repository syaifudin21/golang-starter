package utils

import (
	"exam/internal/i18n"
	"strings"

	"github.com/go-playground/validator/v10"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

func ValidateStruct(structToValidate interface{}, lang string) (map[string]string, bool) {
	validate := validator.New()
	err := validate.Struct(structToValidate)
	if err != nil {
		localizer := i18n.GetLocalizer(lang)
		errors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			field := strings.ToLower(err.Field())
			switch err.Tag() {
			case "required":
				errors[field], _ = localizer.Localize(&goi18n.LocalizeConfig{
					MessageID: "required",
				})
			case "email":
				errors[field], _ = localizer.Localize(&goi18n.LocalizeConfig{
					MessageID: "email",
				})
			case "min":
				errors[field], _ = localizer.Localize(&goi18n.LocalizeConfig{
					MessageID: "min",
					TemplateData: map[string]interface{}{
						"Min": err.Param(),
					},
				})
			default:
				errors[field] = err.Error()
			}
		}
		return errors, false
	}
	return nil, true
}