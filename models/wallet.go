// wallet
package models

import (
	"labix.org/v2/mgo/bson"
)

type Wallet struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"-"`
	WalletId string        `bson:"wallet_id" json:"wallet_id"`
	//Version  int           `json:"version"`
	Payload string `json:"payload"`
}

func (w *Wallet) Find(wid string) error {
	if len(wid) == 0 {
		return nil
	}

	return findOne(cWallets, bson.M{"wallet_id": wid}, nil, w)
}

func (w *Wallet) Save() error {
	if len(w.WalletId) == 0 || len(w.Payload) == 0 {
		return nil
	}
	change := bson.M{
		"$set": bson.M{
			"payload": w.Payload,
		},
	}
	_, err := upsert(cWallets, bson.M{"wallet_id": w.WalletId}, change, true)

	return err
}
