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
		"content": "대통령의 임기연장 또는 중임변경을 위한 헌법개정은 그 헌법개정 제안 당시의 대통령에 대하여는 효력이 없다. 국교는 인정되지 아니하며, 종교와 정치는 분리된다.\r\n"
	  },
	  {
		"id": 4,
		"project_name": "brown",
		"content": "감사원의 조직·직무범위·감사위원의 자격·감사대상공무원의 범위 기타 필요한 사항은 법률로 정한다. 국가는 지역간의 균형있는 발전을 위하여 지역경제를 육성할 의무를 진다.\r\n"
	  },
	  {
		"id": 5,
		"project_name": "brown",
		"content": "대통령이 궐위되거나 사고로 인하여 직무를 수행할 수 없을 때에는 국무총리, 법률이 정한 국무위원의 순서로 그 권한을 대행한다. 모든 국민은 인간으로서의 존엄과 가치를 가지며, 행복을 추구할 권리를 가진다. 국가는 개인이 가지는 불가침의 기본적 인권을 확인하고 이를 보장할 의무를 진다.\r\n"
	  },
	  {
		"id": 6,
		"project_name": "brown",
		"content": "국회는 국무총리 또는 국무위원의 해임을 대통령에게 건의할 수 있다. 대통령의 국법상 행위는 문서로써 하며, 이 문서에는 국무총리와 관계 국무위원이 부서한다. 군사에 관한 것도 또한 같다.\r\n"
	  },
	  {
		"id": 7,
		"project_name": "green",
		"content": "모든 국민은 학문과 예술의 자유를 가진다. 국가는 농·어민과 중소기업의 자조조직을 육성하여야 하며, 그 자율적 활동과 발전을 보장한다. 헌법개정은 국회재적의원 과반수 또는 대통령의 발의로 제안된다.\r\n"
	  },
	  {
		"id": 8,
		"project_name": "blue",
		"content": "법관은 헌법과 법률에 의하여 그 양심에 따라 독립하여 심판한다. 위원은 정당에 가입하거나 정치에 관여할 수 없다. 언론·출판은 타인의 명예나 권리 또는 공중도덕이나 사회윤리를 침해하여서는 아니된다. 언론·출판이 타인의 명예나 권리를 침해한 때에는 피해자는 이에 대한 피해의 배상을 청구할 수 있다.\r\n"
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
