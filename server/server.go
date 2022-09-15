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
    "project_name": "blue",
    "content": "Eu eiusmod occaecat amet officia cillum ipsum enim. Nisi excepteur consectetur duis reprehenderit aute exercitation deserunt labore quis. Nisi duis nisi cupidatat et Lorem velit enim.\r\n"
  },
  {
    "project_name": "blue",
    "content": "Ipsum incididunt Lorem sit cillum quis. Lorem quis nulla veniam deserunt consectetur irure mollit do ea culpa commodo. Deserunt do nisi ipsum voluptate amet velit. Qui ullamco deserunt officia aute voluptate veniam sunt.\r\n"
  },
  {
    "project_name": "brown",
    "content": "Veniam dolore tempor eiusmod cupidatat deserunt aliquip quis ad sit in exercitation. Non sint elit nulla reprehenderit voluptate ut amet. Dolor occaecat incididunt cupidatat adipisicing culpa in commodo elit laborum.\r\n"
  },
  {
    "project_name": "brown",
    "content": "Irure non dolor ipsum tempor id est. Sint duis sunt qui anim cillum cillum et enim nisi. Velit ipsum reprehenderit voluptate fugiat ullamco laborum ullamco nostrud laborum duis aliqua dolor reprehenderit do. Culpa occaecat aliqua duis labore commodo dolore velit nisi qui. Sint Lorem laboris aute anim anim. In ipsum nisi enim ullamco minim velit amet cillum proident proident pariatur in laborum. Tempor laboris consequat Lorem excepteur officia quis pariatur ex.\r\n"
  },
  {
    "project_name": "brown",
    "content": "Elit cillum ut esse occaecat anim occaecat tempor sunt proident nisi. Eu voluptate sint minim ea magna amet dolor mollit magna incididunt. Id ut mollit duis sit. Incididunt commodo anim adipisicing eu duis ad do excepteur tempor cillum consequat pariatur. Proident occaecat sint officia eiusmod est enim cillum qui sunt et ea nulla incididunt duis. Minim anim deserunt aliquip reprehenderit elit duis.\r\n"
  },
  {
    "project_name": "brown",
    "content": "Amet qui quis do eiusmod proident ut nostrud. Mollit exercitation consequat pariatur aute exercitation do cupidatat. Cillum nostrud ad id nisi culpa. Dolor Lorem minim commodo deserunt nulla officia anim est adipisicing aute duis velit exercitation. Nisi quis tempor exercitation veniam nostrud do ex voluptate ut.\r\n"
  },
  {
    "project_name": "brown",
    "content": "Cupidatat ex nisi commodo qui in aute enim aliqua incididunt culpa eiusmod est officia mollit. Est velit adipisicing consectetur irure Lorem. Esse veniam aliqua occaecat consectetur deserunt elit in eiusmod non in id sunt. Culpa proident nostrud ipsum aliqua anim cillum. Cupidatat nulla ea dolore ea enim amet sunt sit. Fugiat pariatur esse commodo esse sunt officia.\r\n"
  },
  {
    "project_name": "blue",
    "content": "Ut nostrud reprehenderit ullamco culpa labore enim nostrud. Anim adipisicing deserunt excepteur nulla incididunt qui velit qui adipisicing Lorem duis nulla quis excepteur. Enim eu cupidatat ullamco sint aliqua. Irure ipsum id laborum dolor et dolore magna quis excepteur aliquip.\r\n"
  },
  {
    "project_name": "green",
    "content": "Nisi sint est magna aliquip aliqua eiusmod eu mollit eu nulla nulla consequat. Aliqua exercitation ipsum occaecat eu eiusmod ut. Occaecat sunt officia duis aute sit. Non officia quis est laborum fugiat est nisi proident tempor irure duis pariatur. Veniam quis tempor eiusmod ea labore.\r\n"
  },
  {
    "project_name": "green",
    "content": "Aliqua incididunt esse nisi aliquip dolor dolore laborum consectetur commodo quis est aliquip ullamco. Reprehenderit veniam elit elit minim mollit cupidatat nulla veniam non id labore nostrud in reprehenderit. Incididunt duis eu dolore occaecat consequat cupidatat elit ut deserunt adipisicing sint. Aliquip fugiat dolore adipisicing cillum aute laboris irure. Sit eiusmod duis occaecat nostrud anim velit consequat ut do id nostrud officia. Velit exercitation eu veniam magna labore minim est ullamco laborum aliqua pariatur cupidatat ut.\r\n"
  }
]`)
