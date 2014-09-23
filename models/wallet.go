// wallet
package models

import (
	"labix.org/v2/mgo/bson"
)

type Wallet struct {
	Id      bson.ObjectId `bson:"_id,omitempty"`
	Payload string
}

func (w *Wallet) Find(wid string) error {
	if !bson.IsObjectIdHex(wid) {
		return nil
	}

	return findOne(cWallets, bson.M{"_id": bson.ObjectIdHex(wid)}, nil, w)
}

func (w *Wallet) Save() error {
	if len(w.Payload) == 0 {
		return nil
	}
	change := bson.M{
		"$set": bson.M{
			"payload": w.Payload,
		},
	}
	_, err := upsert(cWallets, bson.M{"_id": w.Id}, change, true)

	return err
}
