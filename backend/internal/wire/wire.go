//go:build wireinject
// +build wireinject

package wire

import (
	"incidex/internal/config"

	"github.com/google/wire"
)

// InitializeApp creates a fully-wired App with all dependencies.
func InitializeApp(cfg *config.Config) (*App, error) {
	wire.Build(AllProviders)
	return nil, nil
}
