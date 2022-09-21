package cache

import (
	"encoding/json"

	"github.com/IPFS-NEXIVIL/orbit-grpc/server/orbit/models"
	"github.com/tidwall/buntdb"
)

type Cache struct {
	db     *buntdb.DB
	dbPath string
}

func NewCache(dbPath string) (*Cache, error) {
	var err error

	cache := new(Cache)
	cache.dbPath = dbPath
	cache.db, err = buntdb.Open(cache.dbPath)
	if err != nil {
		return nil, err
	}

	return cache, nil
}

func (cache *Cache) Close() {
	cache.db.Close()
}

func (cache *Cache) StoreArticle(data *models.Data) error {
	modelJson, jsonErr := json.Marshal(data)
	if jsonErr != nil {
		return jsonErr
	}

	err := cache.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(data.ID, string(modelJson), nil)
		return err
	})
	return err
}

func (cache *Cache) LoadData(data *models.Data) error {
	err := cache.db.View(func(tx *buntdb.Tx) error {
		value, err := tx.Get(data.ID)
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(value), data)
		return nil
	})
	return err
}
