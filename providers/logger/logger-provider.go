package loggerprovider

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates logger
func New(environment, level string) (*zap.Logger, error) {
	lcfg := zap.NewProductionConfig()
	atom := zap.NewAtomicLevel()
	_ = atom.UnmarshalText([]byte(level))
	lcfg.Level = atom

	if environment == "development" {
		lcfg = zap.NewDevelopmentConfig()
		lcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return lcfg.Build(zap.Hooks())
}
