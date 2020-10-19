package service

import (
	"fmt"

	"github.com/dreamlu/micro/v2/plugin"
)

var (
	defaultManager = plugin.NewManager()
)

// Plugins lists the service plugins
func Plugins() []plugin.Plugin {
	return defaultManager.Plugins()
}

// Register registers an service plugin
func Register(pl plugin.Plugin) error {
	if plugin.IsRegistered(pl) {
		return fmt.Errorf("%s registered globally", pl.String())
	}
	return defaultManager.Register(pl)
}