// block
package models

import (
	"labix.org/v2/mgo/bson"
	"strconv"
)

func init() {
	ensureIndex(cBlocks, "height")
	ensureIndex(cBlocks, "-time")
}

type Block struct {
	Id         string `bson:"_id"`
	Height     int64
	Index      int64
	Version    int32
	Merkleroot string
	Txs        []string
	Time       int64
	Nonce      uint32
	Bits       string
	Difficulty float64
	Prev       string
	Next       string
}

func (blk *Block) Save() error {
	return save(cBlocks, blk, true)
}

func (blk *Block) Latest() error {
	return findOne(cBlocks, nil, []string{"-index"}, blk)
}

func (blk *Block) Find(s string) error {
	if len(s) == 0 {
		return nil
	}
	index, _ := strconv.ParseInt(s, 10, 64)
	query := bson.M{}
	if index > 0 {
		query["index"] = index
	} else {
		query["_id"] = s
	}
	return findOne(cBlocks, query, nil, blk)
}
