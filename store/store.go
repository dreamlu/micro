package store

import (
	"strings"
	"time"

	"github.com/micro/cli/v2"
	"github.com/dreamlu/go-micro/v2"
	"github.com/dreamlu/go-micro/v2/config/cmd"
	log "github.com/dreamlu/go-micro/v2/logger"
	"github.com/dreamlu/go-micro/v2/store"
	pb "github.com/dreamlu/go-micro/v2/store/service/proto"
	mcli "github.com/dreamlu/micro/v2/cli"
	"github.com/dreamlu/micro/v2/store/handler"
	"github.com/pkg/errors"
)

var (
	// Name of the store service
	Name = "go.micro.store"
	// Address is the store address
	Address = ":8002"
	// Backend is the implementation of the store
	Backend = "memory"
	// Nodes is passed to the underlying backend
	Nodes = []string{"localhost"}
	// Database is passed to the underlying backend if set.
	Database = "micro"
	// Table is passed to the underlying backend if set.
	Table = "store"
)

// run runs the micro server
func Run(ctx *cli.Context, srvOpts ...micro.Option) {
	log.Init(log.WithFields(map[string]interface{}{"service": "store"}))

	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	if len(ctx.String("server_name")) > 0 {
		Name = ctx.String("server_name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}
	if len(ctx.String("store")) > 0 {
		Backend = ctx.String("store")
	}
	if len(ctx.String("store_address")) > 0 {
		Nodes = strings.Split(ctx.String("store_address"), ",")
	}
	if len(ctx.String("store_database")) > 0 {
		Database = ctx.String("store_database")
	}
	if len(ctx.String("store_table")) > 0 {
		Table = ctx.String("store_table")
	}

	// Initialise service
	service := micro.NewService(
		micro.Name(Name),
		micro.RegisterTTL(time.Duration(ctx.Int("register_ttl"))*time.Second),
		micro.RegisterInterval(time.Duration(ctx.Int("register_interval"))*time.Second),
	)

	opts := []store.Option{store.Nodes(Nodes...)}
	if len(Database) > 0 {
		opts = append(opts, store.Database(Database))
	}
	if len(Table) > 0 {
		opts = append(opts, store.Table(Table))
	}

	// the store handler
	storeHandler := &handler.Store{
		Stores: make(map[string]store.Store),
	}

	// get from the existing list of stores
	newStore, ok := cmd.DefaultStores[Backend]
	if !ok {
		log.Fatalf("%s is not an implemented store", Backend)
	}

	log.Infof("Initialising the [%s] store with opts: nodes=%v database=%v table=%v", Backend, Nodes, Database, Table)

	// set the default store
	storeHandler.Default = newStore(opts...)

	// set the internal store
	storeHandler.Internal = newStore(
		store.Nodes(Nodes...),
		store.Database(Database),
		store.Table("internal"),
	)

	// set the new store initialiser
	storeHandler.New = func(database string, table string) (store.Store, error) {
		// return a new default store
		v := newStore(
			store.Nodes(Nodes...),
			store.Database(database),
			store.Table(table),
		)
		if err := v.Init(); err != nil {
			return nil, err
		}
		// Record the new database and table in the internal store
		if err := storeHandler.Internal.Write(&store.Record{
			Key:   "databases/" + database,
			Value: []byte{},
		}); err != nil {
			return nil, errors.Wrap(err, "micro store couldn't store new database in internal table")
		}
		if err := storeHandler.Internal.Write(&store.Record{
			Key:   "tables/" + database + "/" + table,
			Value: []byte{},
		}); err != nil {
			return nil, errors.Wrap(err, "micro store couldn't store new table in internal table")
		}
		return v, nil
	}

	pb.RegisterStoreHandler(service.Server(), storeHandler)

	// start the service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

// Commands is the cli interface for the store service
func Commands(options ...micro.Option) []*cli.Command {
	command := &cli.Command{
		Name:  "store",
		Usage: "Run the micro store service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Set the micro tunnel address :8002",
				EnvVars: []string{"MICRO_SERVER_ADDRESS"},
			},
		},
		Action: func(ctx *cli.Context) error {
			Run(ctx, options...)
			return nil
		},
		Subcommands: mcli.StoreCommands(),
	}

	for _, p := range Plugins() {
		if cmds := p.Commands(); len(cmds) > 0 {
			command.Subcommands = append(command.Subcommands, cmds...)
		}

		if flags := p.Flags(); len(flags) > 0 {
			command.Flags = append(command.Flags, flags...)
		}
	}

	return []*cli.Command{command}
}
