package database

import (
	"context"
	"fmt"
	"log"
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
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"

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

	IPFSNodes    []*core.IpfsNode
	IPFSCoreAPIs []icore.CoreAPI

	OrbitDBs    []orbitdb.OrbitDB
	DocStores   []orbitdb.DocumentStore
	EventStores []orbitdb.EventLogStore
	Events      event.Subscription
}

func (db *Database) init() error {
	var err error

	ctx := context.Background()

	dbPath1 := db.CachePath + "1"
	dbPath2 := db.CachePath + "2"

	log.Println("initializing NewOrbitDB ...")
	db.Logger.Debug("initializing NewOrbitDB ...")
	orbit1, err := orbitdb.NewOrbitDB(ctx, db.IPFSCoreAPIs[0], &orbitdb.NewOrbitDBOptions{
		Directory: &dbPath1,
		Logger:    db.Logger,
	})
	if err != nil {
		return err
	}

	orbit2, err := orbitdb.NewOrbitDB(ctx, db.IPFSCoreAPIs[1], &orbitdb.NewOrbitDBOptions{
		Directory: &dbPath2,
		Logger:    db.Logger,
	})
	if err != nil {
		return err
	}

	db.OrbitDBs = []orbitdb.OrbitDB{orbit1, orbit2}

	access := &accesscontroller.CreateAccessControllerOptions{
		Access: map[string][]string{
			"write": {
				"*",
			},
		},
	}

	address := "nexivil"

	var replication = true

	storetype := "docstore"
	log.Println("initializing OrbitDB.Docs ...")
	db.Logger.Debug("initializing OrbitDB.Docs ...")
	docstore1, err := db.OrbitDBs[0].Docs(ctx, address, &orbitdb.CreateDBOptions{
		AccessController:  access,
		StoreType:         &storetype,
		StoreSpecificOpts: documentstore.DefaultStoreOptsForMap("id"),
		Timeout:           time.Second * 600,
		Replicate:         &replication,
	})
	if err != nil {
		log.Fatalf("%s, %s", err, db.CachePath)
	}

	db.DocStores = []orbitdb.DocumentStore{docstore1}

	// Replica of main database (DocStore[0])
	docstore2, err := db.OrbitDBs[1].Docs(ctx, db.DocStores[0].Address().String(), &orbitdb.CreateDBOptions{
		AccessController:  access,
		StoreType:         &storetype,
		StoreSpecificOpts: documentstore.DefaultStoreOptsForMap("id"),
		Timeout:           time.Second * 600,
		Replicate:         &replication,
	})
	if err != nil {
		log.Fatalf("%s, %s", err, db.CachePath)
	}

	db.DocStores = append(db.DocStores, docstore2)

	log.Printf("main orbit db docstore: %s", db.DocStores[0].Address().String())
	log.Printf("replication 1 of main orbit db docstore: %s", db.DocStores[1].Address().String())

	eventstore1, err := db.OrbitDBs[0].Log(ctx, "replicate-automatically", &orbitdb.CreateDBOptions{
		Directory:        &dbPath1,
		AccessController: access,
	})
	if err != nil {
		log.Fatal(err)
	}

	db.EventStores = []orbitdb.EventLogStore{eventstore1}

	// Replica of main EventStore database (EventStore[0])
	eventstore2, err := db.OrbitDBs[0].Log(ctx, db.EventStores[0].Address().String(), &orbitdb.CreateDBOptions{
		Directory:        &dbPath1,
		AccessController: access,
	})
	if err != nil {
		log.Fatal(err)
	}

	db.EventStores = append(db.EventStores, eventstore2)

	// add message to log
	for i := 0; i < 10; i++ {
		_, err := db.EventStores[0].Add(ctx, []byte(fmt.Sprintf("hello%d", i)))
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("subscribing to EventBus ...")
	db.Logger.Debug("subscribing to EventBus ...")
	db.Events, err = db.DocStores[0].EventBus().Subscribe(new(stores.EventReady))
	if err != nil {
		return nil
	}

	db.Events, err = db.DocStores[0].EventBus().Subscribe(new(stores.EventReplicated))
	if err != nil {
		return nil
	}

	db.Events, err = db.DocStores[1].EventBus().Subscribe(new(stores.EventReady))
	if err != nil {
		return nil
	}

	return nil
}

func (db *Database) GetOwnID() string {
	return db.OrbitDBs[0].Identity().ID
}

func (db *Database) GetOwnPubKey() crypto.PubKey {
	pubKey, err := db.OrbitDBs[0].Identity().GetPublicKey()
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
			err := db.IPFSCoreAPIs[0].Swarm().Connect(db.ctx, *peerInfo)
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

	mocknet := mocknet.New(ctx)

	// Create two IPFS nodes for data replication
	db.Logger.Debug("creating IPFS node ...")

	ipfsNode1, ipfsCoreAPI1, err := createNode(ctx, defaultPath, mocknet)
	if err != nil {
		return nil, err
	}
	ipfsNode2, ipfsCoreAPI2, err := createNode(ctx, defaultPath, mocknet)
	if err != nil {
		return nil, err
	}

	db.IPFSNodes = []*core.IpfsNode{ipfsNode1, ipfsNode2}
	db.IPFSCoreAPIs = []icore.CoreAPI{ipfsCoreAPI1, ipfsCoreAPI2}

	log.Println(db.IPFSNodes)
	log.Println(db.IPFSCoreAPIs)

	log.Printf("node1 is %s", db.IPFSNodes[0].Identity.String())
	log.Printf("node2 is %s", db.IPFSNodes[1].Identity.String())

	_, err = mocknet.LinkPeers(db.IPFSNodes[0].Identity, db.IPFSNodes[1].Identity)
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
					db.URI = db.DocStores[0].Address().String()
					onReady(db.URI)
					continue
				}
			}
		}
	}()

	err = db.DocStores[0].Load(db.ctx, -1)
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
	db.DocStores[0].Close()
	db.OrbitDBs[0].Close()
}

func (db *Database) SubmitData(data *models.Data) error {
	entity, err := StructToMap(*data)
	if err != nil {
		return err
	}
	entity["type"] = "data"

	_, err = db.DocStores[0].Put(db.ctx, entity)
	return err
}

func (db *Database) GetDataByID(id string) (models.Data, error) {
	entity, err := db.DocStores[0].Get(db.ctx, id, &iface.DocumentStoreGetOptions{CaseInsensitive: false})
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

// func (db *Database) ListData() ([]*models.Data, error) {
// 	var data []*models.Data
// 	var dataMap map[string]*models.Data

// 	dataMap = make(map[string]*models.Data)

// 	_, err := db.Store.Query(db.ctx, func(e interface{}) (bool, error) {
// 		entity := e.(map[string]interface{})
// 		if entity["type"] == "data" {
// 			var oneData models.Data
// 			err := mapstructure.Decode(entity, &oneData)
// 			if err == nil {
// 				// TODO: Not sure why mapstructure won't convert this field and simply
// 				//       leave it ""
// 				if entity["in-reply-to-id"] != nil {
// 					oneData.InReplyToID = entity["in-reply-to-id"].(string)
// 				}
// 				db.Cache.LoadData(&oneData)
// 				data = append(data, &oneData)
// 				dataMap[oneData.ID] = data[(len(data) - 1)]
// 			}
// 			return true, err
// 		}
// 		return false, nil
// 	})
// 	if err != nil {
// 		return data, err
// 	}

// 	sort.SliceStable(data, func(i, j int) bool {
// 		return data[i].Date > data[j].Date
// 	})

// 	// var dataRoots []*models.Data
// 	// for i := 0; i < len(data); i++ {
// 	// 	if data[i].InReplyToID != "" {
// 	// 		inReplyTo := data[i].InReplyToID
// 	// 		if _, exist := dataMap[inReplyTo]; exist == true {

// 	// 			(*dataMap[inReplyTo]).Replies =
// 	// 				append((*dataMap[inReplyTo]).Replies, data[i])
// 	// 			(*dataMap[inReplyTo]).LatestReply = data[i].Date
// 	// 			continue
// 	// 		}
// 	// 	}
// 	// 	dataRoots = append(dataRoots, data[i])
// 	// }

// 	// sort.SliceStable(dataRoots, func(i, j int) bool {
// 	// 	iLatest := dataRoots[i].LatestReply
// 	// 	if iLatest <= 0 {
// 	// 		iLatest = dataRoots[i].Date
// 	// 	}

// 	// 	jLatest := dataRoots[j].LatestReply
// 	// 	if jLatest <= 0 {
// 	// 		jLatest = dataRoots[j].Date
// 	// 	}

// 	// 	return iLatest > jLatest
// 	// })

// 	return data, nil
// }
