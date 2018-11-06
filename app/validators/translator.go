package validators

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "gopkg.in/go-playground/validator.v9"
)

var translator ut.Translator

func initTranslator() {
	en := en.New()
	uni := ut.New(en, en)
	translator, _ = uni.GetTranslator("en")

	for _, trans := range []struct {
		tag             string
		translation     string
		customRegisFunc validator.RegisterTranslationsFunc
		customTransFunc validator.TranslationFunc
	}{
		{
			tag:         "required",
			translation: "This field is required.",
		},
		{
			tag:         "sql_exists",
			translation: "Object does not exist.",
		},
		{
			tag:         "sql_select",
			translation: "Object does not exist.",
		},
		{
			tag:         "sql_notexists",
			translation: "Object already exists.",
		},
	} {
		if trans.customTransFunc == nil {
			trans.customTransFunc = translationFunc
		}
		if trans.customRegisFunc == nil {
			trans.customRegisFunc = registrationFunc(trans.tag, trans.translation, true)
		}

		validate.RegisterTranslation(trans.tag, translator, trans.customRegisFunc, trans.customTransFunc) // nolint: errcheck
	}
}

func registrationFunc(tag string, translation string, override bool) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) (err error) {
		if err = ut.Add(tag, translation, override); err != nil {
			return
		}
		return
	}
}

func translationFunc(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T(fe.Tag())
	return t
}

// TranslateErrors returns translated error map.
func TranslateErrors(errs []validator.FieldError) map[string][]string {
	ret := make(map[string][]string)
	for _, fe := range errs {
		fname := fe.Field()
		ret[fname] = append(ret[fname], fe.Translate(translator))
	}
	return ret
}
