package model

import (
	"crypto/sha256"
	"fmt"
	"github.com/hbahadorzadeh/key-master/service"
)

type TokenType string

type User struct {
	service.BasicData
	Email    string `json:"email" bson:"email,omitempty" validate:"required,email"`
	IsRemote bool   `json:"is_remote" bson:"is_remote" validate:"required"`
	Password string `json:"_" bson:"password"`

	//User info
	FirstName string `json:"first_name" bson:"first_name,omitempty" validate:"required"`
	LastName  string `json:"last_name" bson:"last_name,omitempty" validate:"required"`

	//Keys
	Keys []UserKey `json:"keys" bson:"keys"`
}

func (u *User) SetPassword(database *service.MongoDB, password string) error {
	u.Password = fmt.Sprintf("%x", sha256.Sum256(
		[]byte(fmt.Sprintf("%s%x",
			u.Email,
			sha256.Sum256([]byte(password))))))
	database.Update(u)
	return nil
}
