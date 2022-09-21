package orbit

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"time"

	"berty.tech/go-orbit-db/iface"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/cache"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/config"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/database"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func NewLogger(filename string) (*zap.Logger, error) {
	if runtime.GOOS == "windows" {
		zap.RegisterSink("winfile", func(u *url.URL) (zap.Sink, error) {
			return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		})
	}

	cfg := zap.NewDevelopmentConfig()
	if runtime.GOOS == "windows" {
		cfg.OutputPaths = []string{
			"stdout",
			"winfile:///" + filename,
		}
	} else {
		cfg.OutputPaths = []string{
			filename,
		}
	}

	return cfg.Build()
}

func StartOrbitDB() *database.Database {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("loading configuration ...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Panicln(err)
	}
	if !cfg.WasSetup() {
		cfg.Setup()
	}

	log.Println("initializing logger ...")
	logger, err := NewLogger(cfg.Logfile)
	if err != nil {
		log.Panicln(err)
	}

	log.Println("initializing cache ...")
	cch, err := cache.NewCache(cfg.ProgramCachePath)
	if err != nil {
		log.Panicln(err)
	}
	defer cch.Close()

	log.Println("initializing database ...")
	db, err := database.NewDatabase(ctx, cfg.ConnectionString, cfg.DatabaseCachePath, cch, logger)
	if err != nil {
		log.Panicln(err)
	}
	defer db.Disconnect()

	log.Println("connecting database ...")
	err = db.Connect(func(address string) {
	})
	if err != nil {
		log.Panicln(err)
	}

	go func() {
		for {
			_, err := db.IPFSCoreAPI.Swarm().Peers(context.Background())
			if err != nil {
				log.Panicln(err)
			}
			time.Sleep(time.Second * 5)
		}
	}()

	go func() {
		var input string
		for {
			fmt.Scanln(&input)

			switch input {
			case "q":
				return
			case "g":
				fmt.Scanln(&input)
				docs, err := db.Store.Get(ctx, input, &iface.DocumentStoreGetOptions{CaseInsensitive: false})
				if err != nil {

					log.Println(err)
				} else {
					log.Println(docs)
				}
			case "p":
				id, _ := uuid.NewUUID()
				log.Print(id.String())
				_, err = db.Store.Put(ctx, map[string]interface{}{"id": id.String(), "hello": "world"})
				if err != nil {
					log.Println(err)
					log.Println("Error")
				} else {
					log.Println(id)
				}
			case "l":
				docs, err := db.Store.Query(ctx, func(e interface{}) (bool, error) {
					return true, nil
				})
				if err != nil {
					log.Println(err)
				} else {
					log.Println(docs)
				}
			}

		}
	}()

	return db
}
