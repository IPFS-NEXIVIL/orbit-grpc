package orbit

import (
	"log"

	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/database"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/models"
)

type DBInfo struct {
	DB *database.Database
}

func (database DBInfo) SaveAndGetDBData(content string, project string) models.Data {
	// our orbit db
	db := database.DB

	// new data create
	newData := models.NewData()
	log.Println(newData)

	newData.Content = content
	newData.Project = project

	// insert `content` data to orbit db
	db.SubmitData(newData)

	nexivilData, err := db.GetDataByID(newData.ID)
	if err != nil {
		log.Fatal(err)
	}

	return nexivilData
}
