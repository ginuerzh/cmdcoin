// output
package models

import (
	"labix.org/v2/mgo/bson"
)

func init() {
	ensureIndex(cOutputs, "txid", "index")
	ensureIndex(cOutputs, "-block_num")
	ensureIndex(cOutputs, "address")
}

type Output struct {
	Id          bson.ObjectId `bson:"_id,omitempty"`
	Txid        string
	Index       uint32
	BlockHeight int64  `bson:"block_height"`
	BlockHash   string `bson:"block_hash"`
	Address     string
	Balance     int64
	Script      string
}

func (op *Output) Exists() (bool, error) {
	return exists(cOutputs, bson.M{"txid": op.Txid, "index": op.Index})
}

func (op *Output) Save() error {
	op.Id = bson.NewObjectId()
	return save(cOutputs, op, true)
}

func (op *Output) Remove() error {
	return remove(cOutputs, bson.M{"txid": op.Txid, "index": op.Index}, true)
}

func (op *Output) SetHeight(height int64) error {
	change := bson.M{
		"$set": bson.M{
			"block_height": height,
		},
	}
	return update(cOutputs, bson.M{"txid": op.Txid, "index": op.Index}, change, true)
}

func AddrOutputs(addrs []string) ([]Output, error) {
	var outputs []Output
	selector := bson.M{
		"address": bson.M{
			"$in": addrs,
		},
	}
	err := find(cOutputs, selector, nil, 0, 0, nil, nil, &outputs)
	return outputs, err
}

func FinalBalance(address string) (confirmed int64, unconfirmed int64, err error) {

	outputs, err := AddrOutputs([]string{address})
	for i, _ := range outputs {
		if outputs[i].BlockHeight > 0 {
			confirmed += outputs[i].Balance
		} else if outputs[i].BlockHeight < 0 {
			unconfirmed += outputs[i].Balance
		}
	}
	return
}
