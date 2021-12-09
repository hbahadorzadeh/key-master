package model

import (
	"github.com/hbahadorzadeh/key-master/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Secret struct {
	service.BasicData

	Public []byte `json:"public_key" bson:"public_key"`
	EncryptedPrivate []byte `json:"encrypted_private_key" bson:"encrypted_private_key" validate:"required"`
	EncryptedKeys []EncryptedKey `json:"encrypted_keys" bson:"encrypted_keys"`
}

type EncryptedKey struct {
	Key []byte `json:"key" bson:"key"`
	Owner primitive.ObjectID `json:"owner" bson:"owner"`
}
