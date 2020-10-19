// Package plugins includes the plugins we want to load
package plugins

import (
	"github.com/dreamlu/go-micro/v2/config/cmd"

	// import specific plugins
	ckStore "github.com/dreamlu/go-micro/v2/store/cockroach"
	fileStore "github.com/dreamlu/go-micro/v2/store/file"
	memStore "github.com/dreamlu/go-micro/v2/store/memory"
	// we only use CF internally for certs
	cfStore "github.com/dreamlu/micro/v2/internal/plugins/store/cloudflare"
)

func init() {
	// TODO: make it so we only have to import them
	cmd.DefaultStores["cloudflare"] = cfStore.NewStore
	cmd.DefaultStores["cockroach"] = ckStore.NewStore
	cmd.DefaultStores["file"] = fileStore.NewStore
	cmd.DefaultStores["memory"] = memStore.NewStore
}
