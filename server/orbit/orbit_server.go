package orbit

import (
	"log"

	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/database"
	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/models"
)

type DBInfo struct {
	DB *database.Database
}

// type Paste struct {
// 	Content string `form:"content" binding:"required"`
// }

// type UploadFile struct {
// 	File *multipart.FileHeader `form:"file" binding:"required"`
// }

func (database DBInfo) saveAndGetDBData(content string, project string) models.Data {
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

func (database DBInfo) Save() {
	// // json data
	// type ContentRequestBody struct {
	// 	Content string `json:"content"`
	// 	Project string `json:"project"`
	// }

	// var requestBody ContentRequestBody

	newData := models.NewData()
	log.Println(newData)

	// c.BindJSON(&requestBody)

	// data := database.saveAndGetDBData(requestBody.Content, requestBody.Project)

	// c.String(http.StatusOK, "%s data %s save to orbit db success", data.Project, data.Content)
}

// func (database DBInfo) get() {
// 	// our orbit db
// 	db := database.DB

// 	// id := c.Param("id")

// 	nexivilData, err := db.GetDataByID(id)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	date := nexivilData.Date

// 	MillisecondsToDate := time.Unix(0, date*int64(time.Millisecond)).Format("2006-01-02 15:04:05")

// 	c.JSON(http.StatusOK, gin.H{
// 		"id":      nexivilData.ID,
// 		"project": nexivilData.Project,
// 		"date":    MillisecondsToDate,
// 		"content": nexivilData.Content,
// 	})
// }
