package model

type UserKey struct {
	EncryptedPrivateKey []byte `json:"encrypted_private_key" bson:"encrypted_private_key"`
	PublicKey           []byte `json:"public_key" bson:"public_key"`
}
