package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"berty.tech/go-orbit-db/iface"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/contentpb"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/cache"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/config"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/database"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	port int = 8001
)

type server struct {
	contentpb.UnimplementedNexivilServer
	savedContents []*contentpb.ContentResponse
}

// ListContents lists all contents contained within the given bounding project
func (s *server) ListContents(req *contentpb.ContentRequest, stream contentpb.Nexivil_ListContentsServer) error {
	for _, content := range s.savedContents {
		if content.ProjectName == req.ProjectName {
			if err := stream.Send(content); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadFeatures loads features from a JSON file.
func (s *server) loadContents(db *database.Database, filePath string) {
	var data []byte
	var dataa []*models.Data
	if filePath != "" {
		var err error
		data, err = ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to load default contents: %v", err)
		}
	} else {
		// Initialize and start the orbit db

		// example = exampleData

		// get all data from orbit db
		orbitData, err := db.ListData()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("orbit data: %v", orbitData)

		dataa = orbitData
		data, _ = json.Marshal(dataa)
	}
	if err := json.Unmarshal(data, &s.savedContents); err != nil {
		log.Fatalf("Failed to load default contents: %v", err)
	}
}

func newServer(db *database.Database) *server {
	s := &server{}
	s.loadContents(db, "")
	return s
}

// Orbit Logger
func NewLogger(filename string) (*zap.Logger, error) {
	// if runtime.GOOS == "windows" {
	// 	zap.RegisterSink("winfile", func(u *url.URL) (zap.Sink, error) {
	// 		return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	// 	})
	// }

	// cfg := zap.NewDevelopmentConfig()
	// if runtime.GOOS == "windows" {
	// 	cfg.OutputPaths = []string{
	// 		"stdout",
	// 		"winfile:///" + filename,
	// 	}
	// } else {
	// 	cfg.OutputPaths = []string{
	// 		filename,
	// 	}
	// }

	// return cfg.Build()
	zap.RegisterSink("winfile", func(u *url.URL) (zap.Sink, error) {

		return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

	})

	cfg := zap.NewDevelopmentConfig()

	cfg.OutputPaths = []string{

		"stdout",

		"winfile:///" + filename,
	}

	return cfg.Build()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("loading configuration ...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Panicln(err)
	}
	log.Println(cfg.WasSetup())
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
		var project string
		var content string
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
				fmt.Scanln(&project)
				fmt.Scanln(&content)
				id, _ := uuid.NewUUID()
				_, err = db.Store.Put(ctx, map[string]interface{}{"id": id.String(), "project": project, "content": content})
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

	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	contentpb.RegisterNexivilServer(grpcServer, newServer(db))
	log.Printf("server listening at %v", lis.Addr())
	grpcServer.Serve(lis)

}

// // exampleData is a copy of testdata/route_guide_db.json. It's to avoid
// // specifying file path with `go run`.
// var exampleData = []byte(`[
// 	{
// 	  "id": "0a9dd1e1-e566-4a9d-9841-9384b34dd04f",
// 	  "project": "blue",
// 	  "content": "Cupidatat anim esse exercitation et ut ex. Dolor Lorem Lorem occaecat sit nostrud velit laboris cillum. Tempor voluptate dolor ullamco cupidatat occaecat mollit Lorem laboris. Ex qui eu laborum magna amet deserunt Lorem laborum eu consequat adipisicing voluptate commodo.\r\n",
// 	  "date": "2015-01-11 02:09:35"
// 	},
// 	{
// 	  "id": "5f0d12cd-abe4-4cc7-a34d-c37c5867dd0e",
// 	  "project": "green",
// 	  "content": "Eu nulla exercitation laborum fugiat veniam ipsum est tempor. Reprehenderit excepteur officia enim consequat. Pariatur veniam nulla labore sunt aute. Aute duis pariatur proident mollit officia magna pariatur.\r\n",
// 	  "date": "2015-05-29 11:42:02"
// 	},
// 	{
// 	  "id": "e95b8ab9-1298-407b-9f2c-2d549a6cc2ba",
// 	  "project": "green",
// 	  "content": "Dolore amet non duis nisi veniam quis proident ad cillum sit enim sint excepteur exercitation. Nisi deserunt sint labore esse deserunt nulla duis deserunt voluptate quis voluptate laborum non. Ullamco dolore quis ipsum nisi consequat do voluptate consequat mollit nisi excepteur. Ex aliquip et veniam laboris eiusmod nisi duis nostrud incididunt sint exercitation ad eiusmod non. Mollit do sit nisi sunt consectetur anim quis do Lorem. In deserunt culpa deserunt ea ad id. Nisi tempor culpa Lorem nisi eiusmod et nisi non sunt.\r\n",
// 	  "date": "2014-09-24 05:43:38"
// 	},
// 	{
// 	  "id": "f84a78c8-1457-4c44-9cb3-84cf608dff82",
// 	  "project": "blue",
// 	  "content": "Aliqua adipisicing velit nostrud et occaecat ad cupidatat cillum consectetur do officia qui. In sunt et exercitation consequat dolore quis esse aliqua magna non. Incididunt aliquip nulla ex consectetur elit reprehenderit sunt nulla do ex veniam commodo mollit. Ullamco enim officia id occaecat sunt ullamco fugiat esse aliquip et deserunt. Nostrud cupidatat eiusmod sunt eu ad anim. Magna dolor incididunt aliquip sit.\r\n",
// 	  "date": "2016-03-11 07:50:56"
// 	},
// 	{
// 	  "id": "d957c756-034f-46dd-adbc-c09c03b00a8f",
// 	  "project": "blue",
// 	  "content": "Non nostrud anim excepteur commodo laboris sint. Aliquip irure nisi eiusmod nulla aliquip pariatur non. Commodo cupidatat non elit cillum quis voluptate. Dolor laborum elit reprehenderit proident eu occaecat. Non reprehenderit in voluptate eiusmod irure aute elit tempor incididunt sit consequat cillum in. Consectetur esse dolore est laborum ea est.\r\n",
// 	  "date": "2017-07-26 08:39:01"
// 	},
// 	{
// 	  "id": "b9746626-c4da-4939-b075-09ba32bff6fe",
// 	  "project": "blue",
// 	  "content": "Incididunt officia quis sunt esse. Culpa excepteur cupidatat minim veniam est velit non. Ex mollit nulla commodo non incididunt sint tempor commodo aute consequat voluptate. Ipsum esse est nisi do eu ad non cupidatat laborum et et minim tempor eiusmod. Elit veniam deserunt velit qui amet enim consectetur. Aute veniam pariatur adipisicing cupidatat irure consequat. Nisi exercitation qui commodo velit laboris veniam amet duis.\r\n",
// 	  "date": "2016-09-11 06:53:22"
// 	},
// 	{
// 	  "id": "9927172d-ba87-47d8-8f22-c52d39b20fef",
// 	  "project": "green",
// 	  "content": "Reprehenderit laboris labore exercitation veniam do anim pariatur proident eu. Culpa enim consectetur consectetur quis in. Consequat qui et sint dolore qui Lorem esse nostrud esse duis do ex magna voluptate. Est dolore Lorem anim proident cillum voluptate ullamco qui nostrud nulla magna. In mollit ipsum officia culpa cupidatat pariatur laborum. Veniam mollit aute elit mollit non sit Lorem exercitation laborum proident ullamco culpa aute. Ea esse Lorem dolor fugiat do sint id.\r\n",
// 	  "date": "2015-02-11 10:51:15"
// 	},
// 	{
// 	  "id": "dfd1fbbc-fd09-43d6-8891-fdc33dcb7325",
// 	  "project": "brown",
// 	  "content": "Magna proident ea duis ea anim laboris non fugiat laborum ex. Commodo eu consequat aute excepteur. Adipisicing nisi excepteur veniam dolor consequat pariatur. Sit dolore mollit in anim pariatur eiusmod laboris enim nulla dolore aliqua exercitation.\r\n",
// 	  "date": "2016-04-24 01:49:15"
// 	},
// 	{
// 	  "id": "495d85f1-8890-444a-9de0-fc16314efac3",
// 	  "project": "green",
// 	  "content": "Elit eiusmod sit est ex anim exercitation laborum ad laboris. Proident ullamco minim voluptate do tempor. Nisi anim ad sit non magna nisi anim. Culpa reprehenderit anim in ex nisi ex fugiat sit officia fugiat labore anim qui laborum.\r\n",
// 	  "date": "2020-03-24 09:17:30"
// 	},
// 	{
// 	  "id": "50a309fe-9acd-4c23-93db-9a049c06fe71",
// 	  "project": "green",
// 	  "content": "Ullamco velit aliquip non labore. Qui anim aliqua consequat est reprehenderit ut dolore culpa aliqua consectetur sunt incididunt ad. Incididunt pariatur veniam tempor exercitation ea. Qui nulla ad esse velit. Id laboris ex do reprehenderit proident laborum veniam. Sit do eiusmod labore do do ipsum qui consectetur incididunt.\r\n",
// 	  "date": "2016-03-19 08:57:23"
// 	},
// 	{
// 	  "id": "ce8ea92b-f740-42b7-89c5-6af804d0703e",
// 	  "project": "blue",
// 	  "content": "Ea consectetur aliqua pariatur minim quis do laborum cillum qui est non. Esse ipsum officia magna do fugiat consequat. Ad incididunt nulla sit ut quis aute ex cillum ex. Elit officia officia amet et irure ut in aliqua. Nostrud irure voluptate consequat non aliquip velit voluptate est in elit culpa. Ea ut sit pariatur fugiat culpa excepteur cillum aliquip id. Non velit veniam velit mollit ullamco pariatur aute deserunt enim consequat laboris dolore.\r\n",
// 	  "date": "2017-06-16 06:56:11"
// 	},
// 	{
// 	  "id": "d4909dba-9794-40ab-a2d8-a686ce735a4d",
// 	  "project": "green",
// 	  "content": "Est ullamco mollit Lorem quis non. Velit amet aliqua aliqua irure anim ipsum nisi labore. Tempor ex ad voluptate dolor deserunt. Veniam culpa labore dolor sint anim Lorem. Commodo eiusmod irure cupidatat occaecat ad commodo. Est deserunt ex id excepteur et labore. Lorem mollit culpa ipsum et ipsum deserunt mollit fugiat incididunt amet sint irure.\r\n",
// 	  "date": "2021-01-24 09:55:20"
// 	}
//   ]`)
