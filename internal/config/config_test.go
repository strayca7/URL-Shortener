package config

import (
	"errors"
	"testing"

	"github.com/rs/zerolog/log"
)

func TestInitLogger(t *testing.T) {
	log.Info().Str("test", "init").Msg("测试")
	log.Err(errors.New("test error")).Msg("error")
}
