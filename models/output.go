// output
package models

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	//"log"
)

func init() {
	ensureIndex(cOutputs, "txid", "index")
	ensureIndex(cOutputs, "-block_num")
	ensureIndex(cOutputs, "address")
	ensureIndex(cOutputs, "age")
}

type Output struct {
	Id          bson.ObjectId `bson:"_id,omitempty"`
	Txid        string
	Index       uint32
	BlockHeight int64    `bson:"block_height"`
	BlockHash   string   `bson:"block_hash"`
	Vin         []string `bson:",omitempty"`
	Address     string
	Balance     int64
	Script      string
	Age         int64
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

func (op *Output) ApplyRemove(txid string, index uint32) error {
	_, err := apply(cOutputs, bson.M{"txid": txid, "index": index}, mgo.Change{Remove: true}, op)
	return err
}

func (op *Output) applyUpdate(txid string, index uint32, update interface{}) error {
	_, err := apply(cOutputs, bson.M{"txid": txid, "index": index}, mgo.Change{Update: update}, op)
	return err
}

func (op *Output) SetHeight(height int64) error {
	change := bson.M{
		"$set": bson.M{
			"block_height": height,
		},
	}
	return op.applyUpdate(op.Txid, op.Index, change)
}

func AddrOutputs(addrs []string) ([]Output, error) {
	var outputs []Output
	selector := bson.M{
		"address": bson.M{
			"$in": addrs,
		},
		"block_height": bson.M{
			"$gte": 0,
		},
	}
	err := find(cOutputs, selector, nil, 0, 0, []string{"age"}, nil, &outputs)
	return outputs, err
}

func FinalBalance(address string, addrs []string) (confirmed int64, unconfirmed int64, err error) {
	outputs, err := AddrOutputs([]string{address})
	for _, op := range outputs {
		if op.BlockHeight == 0 && len(op.Vin) > 0 {
			/*
				find := false
				for _, addr := range addrs {
					if addr == op.Vin[0] {
						find = true
						break
					}
				}

				if !find {
					unconfirmed += op.Balance
					continue
				}
			*/

			if op.Vin[0] != address {
				unconfirmed += op.Balance
				continue
			}
		}
		confirmed += op.Balance
	}
	return
}
