package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

var validate = validator.New()

func ValidateStruct(s any) error {
	if err := validate.Struct(s); err != nil {
		log.Err(err).Msg("Validation failed")
		return err
	}
	return nil
}
