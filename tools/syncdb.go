// syncdb
package main

import (
	"fmt"
	"github.com/conformal/btcjson"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcwire"
	"github.com/ginuerzh/cmdcoin/models"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const (
	Satoshi = 100000000
)

var (
	client *btcrpcclient.Client
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configs := coinConfigs()

	cfg := &btcrpcclient.ConnConfig{
		Host:         "localhost:8110",
		User:         configs["rpcuser"],
		Pass:         configs["rpcpassword"],
		DisableTLS:   true,
		HttpPostMode: true,
	}
	var err error
	client, err = btcrpcclient.New(cfg, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func coinConfigs() map[string]string {
	configs := map[string]string{}

	data, err := ioutil.ReadFile(os.Getenv("HOME") + "/.cmdcoin/cmdcoin.conf")
	if err != nil {
		return configs
	}

	entries := strings.Split(string(data), "\n")
	for _, entry := range entries {
		kv := strings.Split(entry, "=")
		if len(kv) == 2 {
			configs[kv[0]] = kv[1]
		}
	}
	return configs
}

func syncdb() {
	blocks, _ := client.GetBlockCount()
	lastblock := &models.Block{}
	if err := lastblock.Latest(); err != nil {
		log.Println(err)
	}
	blkHeight := lastblock.Height
	blkIndex := lastblock.Index

	lasttx := &models.Tx{}
	if err := lasttx.Last(); err != nil {
		log.Println(err)
	}
	txIndex := lasttx.Index
	if blkHeight < blocks {
		fmt.Printf("blocks: %d/%d\n", blkHeight, blocks)
	}
	for blkHeight < blocks {
		blkHeight++
		blkIndex++
		hash, err := client.GetBlockHash(blkHeight)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println("block", blkHeight, hash)
		blk, err := client.GetBlockVerbose(hash, true)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println(time.Unix(blk.Time, 0).Format("2006-01-02 15:04:05"))
		saveBlock(blk, blkIndex)

		for _, txid := range blk.Tx {
			txIndex++
			fmt.Println("txid:", txid, txIndex)
			hash, err := btcwire.NewShaHashFromStr(txid)
			if err != nil {
				log.Println(err)
				continue
			}

			rtx, err := client.GetRawTransactionVerbose(hash)
			if err != nil {
				log.Println(err)
				continue
			}

			saveTx(rtx, txIndex)

			syncOutputs(rtx, blkHeight)
		}
	}
}

func syncOutputs(rtx *btcjson.TxRawResult, height int64) {
	var vin []string

	for _, in := range rtx.Vin {
		if in.IsCoinBase() {
			continue // omit coinbase
		}

		op := &models.Output{}
		if err := op.ApplyRemove(in.Txid, in.Vout); err != nil { // spent, remove it
			log.Println(err)
			continue
		}
		if len(op.Address) > 0 {
			vin = append(vin, op.Address)
		}
	}

	saveOutputs(rtx, height, vin)
}

func saveOutputs(rtx *btcjson.TxRawResult, height int64, vin []string) {
	for _, out := range rtx.Vout {
		switch out.ScriptPubKey.Type {
		case "pubkeyhash", "pubkey", "scripthash":
			op := &models.Output{
				Txid:        rtx.Txid,
				BlockHeight: height,
				BlockHash:   rtx.BlockHash,
				Index:       out.N,
				Vin:         vin,
				Address:     out.ScriptPubKey.Addresses[0],
				Balance:     int64(out.Value * Satoshi),
				Script:      out.ScriptPubKey.Hex,
				Age:         rtx.Time,
			}

			if height > 0 {
				if err := op.Remove(); err != nil {
					log.Println(err)
				}
				op.Save()
				fmt.Println("confirmed output:", op.Address, op.Balance)
			} else {
				if exists, err := op.Exists(); err == nil && !exists {
					if err := op.Save(); err != nil {
						log.Println(err)
					}
					fmt.Println("unconfirmed output:", op.Address, op.Balance)
				}
			}

		default:
			log.Println("Unknown script type:", out.ScriptPubKey.Type)
		}
	}
}

func getUnconfirmedTx() {
	hashs, err := client.GetRawMempool()
	if err != nil {
		log.Println(err)
		return
	}
	for _, hash := range hashs {
		tx, err := client.GetRawTransactionVerbose(hash)
		if err != nil {
			log.Println(err)
			continue
		}

		mtx := &models.Tx{
			Hash:  tx.Txid,
			Block: tx.BlockHash,
		}
		exist, err := mtx.Exists()
		if err != nil {
			log.Println(err)
		}
		if err == nil && !exist {
			fmt.Println("Unconfirmed tx:", tx.Txid)
			saveTx(tx, 0)

			var vin []string
			for _, in := range tx.Vin {
				if in.IsCoinBase() {
					continue // omit coinbase
				}
				op := &models.Output{
					Txid:  in.Txid,
					Index: in.Vout,
				}
				op.SetHeight(-1)
				if len(op.Address) > 0 {
					vin = append(vin, op.Address)
				}
			}
			saveOutputs(tx, 0, vin)
		}
	}
}

func saveBlock(block *btcjson.BlockResult, index int64) {
	blk := &models.Block{
		Id:         block.Hash,
		Height:     block.Height,
		Index:      index,
		Version:    block.Version,
		Merkleroot: block.MerkleRoot,
		Txs:        block.Tx,
		Time:       block.Time,
		Nonce:      block.Nonce,
		Bits:       block.Bits,
		Difficulty: block.Difficulty,
		Prev:       block.PreviousHash,
		Next:       block.NextHash,
	}
	if err := blk.Save(); err != nil {
		log.Println(err)
	}
}

func saveTx(rtx *btcjson.TxRawResult, index int64) {
	tx := &models.Tx{
		Hash:    rtx.Txid,
		Block:   rtx.BlockHash,
		Version: rtx.Version,
		Time:    rtx.Time,
		Index:   index,
	}
	if tx.Time == 0 {
		tx.Time = time.Now().Unix()
	}
	for _, in := range rtx.Vin {
		if in.IsCoinBase() {
			continue
		}

		pretx := &models.Tx{}
		if err := pretx.Find(in.Txid); err != nil {
			log.Println(err)
			continue
		}

		vin := &models.Vin{
			Txid:     in.Txid,
			Sequence: in.Sequence,
			Script:   in.ScriptSig.Hex,
			PrevOut:  *pretx.Vout[int(in.Vout)],
		}
		tx.Vin = append(tx.Vin, vin)
	}
	for _, out := range rtx.Vout {
		if len(out.ScriptPubKey.Addresses) == 0 {
			continue
		}
		vout := &models.Vout{
			Value:      int64(out.Value * Satoshi),
			N:          out.N,
			Script:     out.ScriptPubKey.Hex,
			ScriptType: out.ScriptPubKey.Type,
			Address:    out.ScriptPubKey.Addresses[0],
		}
		tx.Vout = append(tx.Vout, vout)
	}
	if err := tx.Remove(tx.Hash, ""); err != nil {
		log.Println(err)
	}
	if err := tx.Save(); err != nil {
		log.Println(err)
	}
}

func main() {
	timer := time.NewTimer(time.Second * 0)
	for {
		select {
		case <-timer.C:
			syncdb()
			getUnconfirmedTx()
			timer = time.NewTimer(time.Second * 1)
		}
	}
}
