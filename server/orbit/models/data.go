package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Data struct {
	ID          string  `mapstructure:"_id" json:"-" validate:"uuid_rfc4122"`
	InReplyToID string  `mapstructure:"in-reply-to-id" json:"-" validate:"omitempty,uuid_rfc4122"`
	Project     string  `mapstructure:"project" json:"-" validate:"required,string"`
	Replies     []*Data `mapstructure:"-" json:"-" validate:"-"`
	LatestReply int64   `mapstructure:"-" json:"-" validate:"-"`
	Date        int64   `mapstructure:"date" json:"-" validate:"required,number"`
	Content     string  `mapstructure:"content" json:"-" validate:"required,string"`
}

func NewData() *Data {
	data := new(Data)

	id, _ := uuid.NewUUID()
	data.ID = id.String()

	data.Date = time.Now().UnixNano() / int64(time.Millisecond)

	return data
}

func (data *Data) IsValid() (bool, error) {
	validate := validator.New()
	errs := validate.Struct(data)
	if errs != nil {
		return false, errs
	}

	return true, nil
}
