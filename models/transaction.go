// transaction
package models

import (
	"labix.org/v2/mgo/bson"
	"strconv"
)

func init() {
	ensureIndex(cTxs, "block")
	ensureIndex(cTxs, "-index")
	ensureIndex(cTxs, "-time")
}

type Vin struct {
	Txid     string
	Vout     uint32
	Coinbase string
}

type Vout struct {
	Value      int64
	N          uint32
	Script     string
	ScriptType string
	Address    string
}

type Tx struct {
	Id      bson.ObjectId `bson:"_id"`
	Hash    string
	Block   string
	Version int32
	Index   int64
	Time    int64
	Vin     []*Vin
	Vout    []*Vout
}

func (tx *Tx) Exists() (bool, error) {
	return exists(cTxs, bson.M{"hash": tx.Hash, "block": tx.Block})
}

func (tx *Tx) Save() error {
	tx.Id = bson.NewObjectId()
	return save(cTxs, tx, true)
}

func (tx *Tx) Upsert() error {
	change := bson.M{
		"index": tx.Index,
		"block": tx.Block,
	}
	_, err := upsert(cTxs, bson.M{"hash": tx.Hash}, change, true)
	return err
}

func (tx *Tx) Find(s string) error {
	if len(s) == 0 {
		return nil
	}
	index, _ := strconv.ParseInt(s, 10, 64)
	query := bson.M{}
	if index > 0 {
		query["index"] = index
	} else {
		query["hash"] = s
	}

	return findOne(cTxs, query, nil, tx)
}

func (tx *Tx) Remove(txid, block string) error {
	return remove(cTxs, bson.M{"hash": txid, "block": block}, true)
}

func (tx *Tx) Last() error {
	return findOne(cTxs, nil, []string{"-index"}, tx)
}

func UnconfirmedTxs() ([]Tx, error) {
	var txs []Tx
	err := find(cTxs, bson.M{"block": "", "index": 0}, nil, 0, 0, []string{"-time"}, nil, &txs)
	return txs, err
}
