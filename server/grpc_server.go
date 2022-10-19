package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"runtime"
	"time"

	"berty.tech/go-orbit-db/iface"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/contentpb"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/cache"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/config"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/database"
	corepath "github.com/ipfs/interface-go-ipfs-core/path"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	port int = 8001
)

type server struct {
	contentpb.UnimplementedNexivilServer
	hotDataId string

	DB *database.Database
}

func (s *server) NexivilContent(req *contentpb.ContentRequest, stream contentpb.Nexivil_NexivilContentServer) error {
	for {
		time.Sleep(time.Second * 5)

		switch {

		// Stream data none when there is no hot data
		case s.hotDataId == "":
			err := stream.Send(&contentpb.ContentResponse{Id: "None", Date: "None", ProjectName: "None", Content: "None"})
			if err != nil {
				log.Println("Error sending metric message ", err)
			}

		case s.hotDataId != "":
			orbitData, err := s.DB.GetDataByID(s.hotDataId)
			if err != nil {
				log.Println("Failed to get Orbit Data", err)
			}

			id := orbitData.ID
			date := orbitData.Date
			strDate := time.Unix(0, date*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
			projectName := orbitData.Project
			content := orbitData.Content

			err = stream.Send(&contentpb.ContentResponse{Id: id, Date: strDate, ProjectName: projectName, Content: content})
			if err != nil {
				log.Println("Error sending metric message ", err)
			}
		}
	}
}

// // loadFeatures loads features from a JSON file.
// func (s *server) loadContents(filePath string) {
// 	var data []byte
// 	var dataa []*models.Data
// 	if filePath != "" {
// 		var err error
// 		data, err = ioutil.ReadFile(filePath)
// 		if err != nil {
// 			log.Fatalf("Failed to load default contents: %v", err)
// 		}
// 	} else {
// 		// get all data from orbit db
// 		orbitData, err := s.DB.GetDataByID(s.hotDataId)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		log.Printf("orbit data: %v", orbitData)

// 		dataa = orbitData
// 		data, _ = json.Marshal(dataa)
// 	}
// 	if err := json.Unmarshal(data, &s.savedContents); err != nil {
// 		log.Fatalf("Failed to load default contents: %v", err)
// 	}
// }

func newServer(db *database.Database) *server {
	s := &server{}
	s.DB = db
	s.hotDataId = ""
	return s
}

// Orbit Logger
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

	// pin orbit db

	path := corepath.IpfsPath(db.Store.Address().GetRoot())

	db.IPFSCoreAPI.Pin().Add(ctx, path)

	// Check Pin Status

	pin := db.IPFSCoreAPI.Pin()

	_, ok, err := pin.IsPinned(ctx, corepath.IpfsPath(db.Store.Address().GetRoot()))

	if err != nil {

		log.Panicln(err)

	}

	log.Println("Check Pin Status", ok)

	go func() {
		for {
			_, err := db.IPFSCoreAPI.Swarm().Peers(context.Background())
			if err != nil {
				log.Panicln(err)
			}
			time.Sleep(time.Second * 5)
		}
	}()

	// server
	var nexServer *server

	// Communicate with Orbit DB through command input to cli
	go func() {
		var input string
		var project string
		var content string
		for {
			fmt.Scanln(&input)

			switch input {
			case "q":
				return
			// Load specific data with id
			case "g":
				fmt.Scanln(&input)
				docs, err := db.Store.Get(ctx, input, &iface.DocumentStoreGetOptions{CaseInsensitive: false})
				if err != nil {

					log.Println(err)
				} else {
					log.Println(docs)
				}
			// Putting data
			case "p":
				fmt.Scanln(&project)
				fmt.Scanln(&content)

				database := orbit.DBInfo{
					DB: db,
				}

				data := database.SaveAndGetDBData(content, project)

				log.Printf("%s data %s save to orbit db success", data.Project, data.Content)

				nexServer.hotDataId = data.ID

			// Load all data
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
	nexServer = newServer(db)
	contentpb.RegisterNexivilServer(grpcServer, nexServer)
	log.Printf("server listening at %v", lis.Addr())
	grpcServer.Serve(lis)
}
