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
	Txid     string `json:"txid"`
	Sequence uint32 `json:"sequence"`
	Script   string `json:"script"`
	PrevOut  Vout   `json:"prev_out"`
}

type Vout struct {
	Value      int64  `json:"value"`
	N          uint32 `json:"n"`
	Script     string `json:"script"`
	ScriptType string `json:"type"`
	Address    string `json:"addr"`
}

type Tx struct {
	Id      bson.ObjectId `bson:"_id" json:"-"`
	Hash    string        `json:"hash"`
	Block   string        `json:"block"`
	Version int32         `json:"version"`
	Index   int64         `json:"tx_index"`
	Time    int64         `json:"time"`
	Vin     []*Vin        `json:"inputs"`
	Vout    []*Vout       `json:"outputs"`
}

func (tx *Tx) AddInput(input *Vin) {
	tx.Vin = append(tx.Vin, input)
}

func (tx *Tx) AddOutput(output *Vout) {
	tx.Vout = append(tx.Vout, output)
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

func AddrTxs(addrs []string) ([]Tx, error) {
	if len(addrs) == 0 {
		return nil, nil
	}

	var txs []Tx
	selector := bson.M{
		"$or": []bson.M{
			{"vin.prevout.address": bson.M{"$in": addrs}},
			{"vout.address": bson.M{"$in": addrs}},
		},
	}
	err := find(cTxs, selector, nil, 0, 20, []string{"-time"}, nil, &txs)
	return txs, err
}
