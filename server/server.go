package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/IPFS-NEXIVIL/orbit-grpc/server/contentpb"
	"google.golang.org/grpc"
)

var (
	port int = 8080
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
func (s *server) loadContents(filePath string) {
	var data []byte
	if filePath != "" {
		var err error
		data, err = ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to load default contents: %v", err)
		}
	} else {
		data = exampleData
	}
	if err := json.Unmarshal(data, &s.savedContents); err != nil {
		log.Fatalf("Failed to load default contents: %v", err)
	}
}

func newServer() *server {
	s := &server{}
	s.loadContents("")
	return s
}

func main() {
	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	contentpb.RegisterNexivilServer(grpcServer, newServer())
	log.Printf("server listening at %v", lis.Addr())
	grpcServer.Serve(lis)
}

// exampleData is a copy of testdata/route_guide_db.json. It's to avoid
// specifying file path with `go run`.
var exampleData = []byte(`[
	{
		"id": 3,
		"project_name": "blue",
		"content": "Ex exercitation officia exercitation exercitation Lorem. Ea laboris occaecat aliquip nulla minim duis. Anim exercitation irure commodo irure non. Lorem aliquip minim sit est sint id qui quis sit ipsum reprehenderit Lorem aute cupidatat.\r\n"
	  },
	  {
		"id": 4,
		"project_name": "brown",
		"content": "Esse veniam ullamco pariatur aute cillum fugiat id sit ea irure anim. Culpa magna magna qui magna qui ex laboris sint sunt qui excepteur id. Adipisicing magna pariatur pariatur deserunt nulla labore laboris commodo cupidatat adipisicing. Esse ad ullamco est dolor enim ex irure quis minim pariatur qui nulla. Ex aliqua sit ut ea. Pariatur fugiat nostrud occaecat adipisicing culpa incididunt ea.\r\n"
	  },
	  {
		"id": 5,
		"project_name": "brown",
		"content": "Elit consequat incididunt nisi cillum aliquip do consectetur magna sunt irure mollit. Laborum aliquip ea do laboris reprehenderit ut aliqua cupidatat. Exercitation commodo do est do Lorem adipisicing nulla commodo aliquip deserunt non exercitation ad eu. In occaecat anim dolor exercitation ea irure magna. Proident officia magna adipisicing ut occaecat. Ad adipisicing dolore ea do cupidatat magna ea qui eiusmod nulla consectetur labore.\r\n"
	  },
	  {
		"id": 6,
		"project_name": "brown",
		"content": "Veniam labore ullamco in incididunt. Irure mollit laborum pariatur consequat id esse velit anim nulla occaecat. Eu velit sint ullamco consectetur proident enim voluptate.\r\n"
	  },
	  {
		"id": 7,
		"project_name": "green",
		"content": "Veniam dolor laboris mollit eiusmod cupidatat in culpa ex quis aliqua pariatur eiusmod. Anim tempor cupidatat mollit culpa. Non proident dolore anim Lorem reprehenderit excepteur consectetur nostrud ullamco aliquip cupidatat incididunt.\r\n"
	  },
	  {
		"id": 8,
		"project_name": "blue",
		"content": "Labore dolor nostrud labore voluptate aliqua. Veniam fugiat eu ea exercitation id eiusmod ad commodo commodo mollit irure Lorem. Esse non ut ex eu occaecat irure ad esse.\r\n"
	  },
	  {
		"id": 9,
		"project_name": "blue",
		"content": "Aliqua consectetur anim aliqua et sit. Ea ad deserunt cupidatat minim deserunt quis elit nisi officia. Eiusmod ad consequat elit eu ad. Anim aliquip ad ad non. In magna voluptate cupidatat aute anim pariatur anim officia proident cupidatat consectetur id. Adipisicing veniam non cillum voluptate irure culpa exercitation exercitation duis. Dolore sit voluptate pariatur Lorem laboris exercitation esse nostrud nulla in aliquip sit.\r\n"
	  },
	  {
		"id": 10,
		"project_name": "green",
		"content": "Sunt sunt quis tempor pariatur incididunt. Est duis sunt reprehenderit non aliquip laborum occaecat culpa proident. Aute tempor deserunt ad anim eu proident voluptate. Voluptate sunt ea irure est amet et. Non nisi commodo enim officia nulla aliqua cillum amet cillum fugiat enim. Aliqua fugiat culpa Lorem ex et minim laborum exercitation est irure eiusmod ut. Adipisicing exercitation esse pariatur nulla.\r\n"
	  },
	  {
		"id": 11,
		"project_name": "green",
		"content": "Minim nostrud in irure ipsum. Cillum est enim et pariatur irure nulla sit ullamco aliquip eu cupidatat. Magna sunt consectetur quis cupidatat do. Culpa cupidatat culpa nulla deserunt. Aliquip voluptate ea reprehenderit consequat esse ex ut aute elit minim irure aliqua pariatur.\r\n"
	  },
	  {
		"id": 12,
		"project_name": "green",
		"content": "Reprehenderit veniam consequat est non eiusmod cupidatat Lorem nostrud. Ipsum irure exercitation proident et minim veniam labore magna fugiat ad eiusmod ex id. Laboris proident minim ut eu tempor dolor. Occaecat amet laboris consequat Lorem consequat culpa laborum proident magna. Excepteur exercitation exercitation officia nulla irure consequat magna.\r\n"
	  },
	  {
		"id": 13,
		"project_name": "blue",
		"content": "Et deserunt et ex amet aliquip consequat deserunt aliquip ad consequat. Tempor et est reprehenderit ad culpa incididunt esse. Esse officia voluptate velit fugiat nulla qui laborum ullamco exercitation consequat eu incididunt. Consequat sunt ipsum sint dolor veniam non ullamco minim in ipsum tempor nostrud laborum. Eu eiusmod dolor eu cillum mollit velit nisi esse nulla sunt veniam aliquip fugiat. Consectetur Lorem consectetur qui laborum mollit sunt commodo duis officia proident do est est pariatur. Fugiat ut consectetur anim proident adipisicing cillum commodo irure.\r\n"
	  },
	  {
		"id": 14,
		"project_name": "brown",
		"content": "Consectetur qui incididunt id cillum mollit irure elit qui. Magna nisi nostrud consequat adipisicing velit aliquip aute id aliqua. Nostrud laboris enim veniam eu est consectetur labore consectetur proident enim fugiat cillum. Est ullamco magna proident fugiat.\r\n"
	  }
]`)
