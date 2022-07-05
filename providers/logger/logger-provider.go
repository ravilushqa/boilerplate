package loggerprovider

import (
	"go.uber.org/zap"
)

// New creates logger
func New(level string) (*zap.Logger, error) {
	lcfg := zap.NewProductionConfig()
	lcfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	atom := zap.NewAtomicLevel()
	_ = atom.UnmarshalText([]byte(level))

	lcfg.Level = atom

	return lcfg.Build(zap.Hooks())
}
