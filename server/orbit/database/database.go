package database

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	orbitdb "berty.tech/go-orbit-db"
	"berty.tech/go-orbit-db/accesscontroller"
	"berty.tech/go-orbit-db/iface"
	"berty.tech/go-orbit-db/stores"
	"berty.tech/go-orbit-db/stores/documentstore"
	config "github.com/ipfs/go-ipfs-config"
	"github.com/ipfs/go-ipfs/core"
	icore "github.com/ipfs/interface-go-ipfs-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/cache"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/models"
)

type Database struct {
	ctx              context.Context
	ConnectionString string
	URI              string
	CachePath        string
	Cache            *cache.Cache

	Logger *zap.Logger

	IPFSNode    *core.IpfsNode
	IPFSCoreAPI icore.CoreAPI

	OrbitDB orbitdb.OrbitDB
	Store   orbitdb.DocumentStore
	Events  event.Subscription
}

func (db *Database) init() error {
	var err error

	ctx := context.Background()

	log.Println("initializing NewOrbitDB ...")
	db.Logger.Debug("initializing NewOrbitDB ...")
	db.OrbitDB, err = orbitdb.NewOrbitDB(ctx, db.IPFSCoreAPI, &orbitdb.NewOrbitDBOptions{
		Directory: &db.CachePath,
		Logger:    db.Logger,
	})
	if err != nil {
		return err
	}

	ac := &accesscontroller.CreateAccessControllerOptions{
		Access: map[string][]string{
			"write": {
				"*",
			},
		},
	}

	address := "nexivil"

	storetype := "docstore"
	log.Println("initializing OrbitDB.Docs ...")
	db.Logger.Debug("initializing OrbitDB.Docs ...")
	db.Store, err = db.OrbitDB.Docs(ctx, address, &orbitdb.CreateDBOptions{
		AccessController:  ac,
		StoreType:         &storetype,
		StoreSpecificOpts: documentstore.DefaultStoreOptsForMap("id"),
		Timeout:           time.Second * 600,
	})
	if err != nil {
		log.Fatalf("%s, %s", err, db.CachePath)
	}

	log.Printf("%s", db.Store.Address().String())

	log.Println("subscribing to EventBus ...")
	db.Logger.Debug("subscribing to EventBus ...")
	db.Events, err = db.Store.EventBus().Subscribe(new(stores.EventReady))
	if err != nil {
		return nil
	}

	return nil
}

func (db *Database) GetOwnID() string {
	return db.OrbitDB.Identity().ID
}

func (db *Database) GetOwnPubKey() crypto.PubKey {
	pubKey, err := db.OrbitDB.Identity().GetPublicKey()
	if err != nil {
		return nil
	}

	return pubKey
}

func (db *Database) connectToPeers() error {
	var wg sync.WaitGroup

	peerInfos, err := config.DefaultBootstrapPeers()
	if err != nil {
		return err
	}

	wg.Add(len(peerInfos))
	for _, peerInfo := range peerInfos {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			err := db.IPFSCoreAPI.Swarm().Connect(db.ctx, *peerInfo)
			if err != nil {
				db.Logger.Error("failed to connect", zap.String("peerID", peerInfo.ID.String()), zap.Error(err))
				log.Printf("failed to connect to %s: %s", peerInfo.ID, err)
			} else {
				db.Logger.Debug("connected!", zap.String("peerID", peerInfo.ID.String()))
				log.Printf("connect to %s", peerInfo.ID)
			}
		}(&peerInfo)
	}
	wg.Wait()
	return nil
}

func NewDatabase(
	ctx context.Context,
	dbConnectionString string,
	dbCache string,
	cch *cache.Cache,
	logger *zap.Logger,
) (*Database, error) {
	var err error

	db := new(Database)
	db.ctx = ctx
	db.ConnectionString = dbConnectionString
	db.CachePath = dbCache
	db.Cache = cch
	db.Logger = logger

	db.Logger.Debug("getting config root path ...")
	defaultPath, err := config.PathRoot()
	if err != nil {
		return nil, err
	}

	db.Logger.Debug("setting up plugins ...")
	if err := setupPlugins(defaultPath); err != nil {
		return nil, err
	}

	db.Logger.Debug("creating IPFS node ...")
	db.IPFSNode, db.IPFSCoreAPI, err = createNode(ctx, defaultPath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *Database) Connect(onReady func(address string)) error {
	var err error

	log.Println("connecting to peers ...")
	db.Logger.Info("connecting to peers ...")
	err = db.connectToPeers()
	if err != nil {
		db.Logger.Error("failed to connect: %s", zap.Error(err))
	} else {
		db.Logger.Debug("connected to peer!")
	}

	log.Println("initializing database connection ...")
	db.Logger.Info("initializing database connection ...")
	err = db.init()
	if err != nil {
		db.Logger.Error("%s", zap.Error(err))
		return err
	}

	log.Println("running ...")
	db.Logger.Info("running ...")

	go func() {
		for {
			for ev := range db.Events.Out() {
				db.Logger.Debug("got event", zap.Any("event", ev))
				switch ev.(type) {
				case stores.EventReady:
					db.URI = db.Store.Address().String()
					onReady(db.URI)
					continue
				}
			}
		}
	}()

	err = db.Store.Load(db.ctx, -1)
	if err != nil {
		db.Logger.Error("%s", zap.Error(err))
		// TODO: clean up
		return err
	}

	log.Println("connect done")
	db.Logger.Debug("connect done")
	return nil
}

func (db *Database) Disconnect() {
	db.Events.Close()
	db.Store.Close()
	db.OrbitDB.Close()
}

func (db *Database) SubmitData(data *models.Data) error {
	entity, err := StructToMap(*data)
	if err != nil {
		return err
	}
	entity["type"] = "data"

	_, err = db.Store.Put(db.ctx, entity)
	return err
}

func (db *Database) GetDataByID(id string) (models.Data, error) {
	entity, err := db.Store.Get(db.ctx, id, &iface.DocumentStoreGetOptions{CaseInsensitive: false})
	if err != nil {
		return models.Data{}, err
	}

	var data models.Data
	err = mapstructure.Decode(entity[0], &data)
	if err != nil {
		return models.Data{}, err
	}

	return data, nil
}

func (db *Database) ListData() ([]*models.Data, error) {
	var data []*models.Data
	var dataMap map[string]*models.Data

	dataMap = make(map[string]*models.Data)

	_, err := db.Store.Query(db.ctx, func(e interface{}) (bool, error) {
		entity := e.(map[string]interface{})
		if entity["type"] == "data" {
			var oneData models.Data
			err := mapstructure.Decode(entity, &oneData)
			if err == nil {
				// TODO: Not sure why mapstructure won't convert this field and simply
				//       leave it ""
				if entity["in-reply-to-id"] != nil {
					oneData.InReplyToID = entity["in-reply-to-id"].(string)
				}
				db.Cache.LoadData(&oneData)
				data = append(data, &oneData)
				dataMap[oneData.ID] = data[(len(data) - 1)]
			}
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return data, err
	}

	sort.SliceStable(data, func(i, j int) bool {
		return data[i].Date > data[j].Date
	})

	// var dataRoots []*models.Data
	// for i := 0; i < len(data); i++ {
	// 	if data[i].InReplyToID != "" {
	// 		inReplyTo := data[i].InReplyToID
	// 		if _, exist := dataMap[inReplyTo]; exist == true {

	// 			(*dataMap[inReplyTo]).Replies =
	// 				append((*dataMap[inReplyTo]).Replies, data[i])
	// 			(*dataMap[inReplyTo]).LatestReply = data[i].Date
	// 			continue
	// 		}
	// 	}
	// 	dataRoots = append(dataRoots, data[i])
	// }

	// sort.SliceStable(dataRoots, func(i, j int) bool {
	// 	iLatest := dataRoots[i].LatestReply
	// 	if iLatest <= 0 {
	// 		iLatest = dataRoots[i].Date
	// 	}

	// 	jLatest := dataRoots[j].LatestReply
	// 	if jLatest <= 0 {
	// 		jLatest = dataRoots[j].Date
	// 	}

	// 	return iLatest > jLatest
	// })

	return data, nil
}
