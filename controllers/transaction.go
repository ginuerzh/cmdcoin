// transaction
package controllers

import (
	"github.com/conformal/btcrpcclient"
	//"github.com/conformal/btcutil"
	"bytes"
	"encoding/hex"
	"github.com/conformal/btcwire"
	"github.com/ginuerzh/cmdcoin/models"
	"github.com/martini-contrib/binding"
	"gopkg.in/go-martini/martini.v1"
	"log"
	"net/http"
)

func BindTxApi(m *martini.ClassicMartini) {
	m.Get("/rawtx", binding.Form(rawTxForm{}), rawTxHandler)
	m.Post("/pushtx", binding.Form(pushTxForm{}), pushTxHandler)
	m.Get("/unconfirmed_tx", unconfirmedTxHandler)
}

type rawTxForm struct {
	Tx string `form:"tx"`
}

type Tx struct {
	Hash    string    `json:"hash"`
	Version int32     `json:"ver"`
	Index   int64     `json:"tx_index"`
	Inputs  []*Input  `json:"inputs"`
	Outputs []*Output `json:"outputs"`
}

type Input struct {
	PreOut   *Output `json:"prev_out"`
	Coinbase string  `json:"coinbase"`
}

type Output struct {
	Value   int64  `json:"value"`
	N       uint32 `json:"n"`
	Addr    string `json:"addr"`
	Script  string `json:"script"`
	TxIndex int64  `json:"tx_index"`
}

func NewTx(ver int32, hash string, index int64) *Tx {
	return &Tx{
		Hash:    hash,
		Version: ver,
		Index:   index,
	}
}

func (tx *Tx) AddInput(input *Input) {
	tx.Inputs = append(tx.Inputs, input)
}

func (tx *Tx) AddOutput(output *Output) {
	tx.Outputs = append(tx.Outputs, output)
}

func transformTx(mtx *models.Tx) *Tx {
	tx := NewTx(mtx.Version, mtx.Hash, mtx.Index)

	for _, in := range mtx.Vin {
		pretx := &models.Tx{}
		pretx.Find(in.Txid)
		var preOut *Output
		if len(pretx.Vout) > 0 {
			vout := pretx.Vout[int(in.Vout)]
			preOut = &Output{
				N:       in.Vout,
				Value:   vout.Value,
				Addr:    vout.Address,
				Script:  vout.Script,
				TxIndex: pretx.Index,
			}
		}
		tx.AddInput(&Input{preOut, in.Coinbase})
	}
	for _, out := range mtx.Vout {
		op := &Output{
			N:       out.N,
			Value:   out.Value,
			Addr:    out.Address,
			Script:  out.Script,
			TxIndex: mtx.Index,
		}
		tx.AddOutput(op)
	}

	return tx
}

func rawTxHandler(resp http.ResponseWriter, form rawTxForm) {
	mtx := &models.Tx{}
	if err := mtx.Find(form.Tx); err != nil {
		log.Println(err)
	}

	writeResponse(resp, transformTx(mtx))
}

type pushTxForm struct {
	RawTx string `form:"rawtx"`
}

func pushTxHandler(resp http.ResponseWriter, client *btcrpcclient.Client, form pushTxForm) {
	msgTx := btcwire.NewMsgTx()
	b, err := hex.DecodeString(form.RawTx)
	if err != nil {
		writeResponse(resp, err.Error())
		return
	}
	if err := msgTx.BtcDecode(bytes.NewBuffer(b), 1); err != nil {
		writeResponse(resp, err.Error())
		return
	}
	hash, err := client.SendRawTransaction(msgTx, false)
	if err != nil {
		writeResponse(resp, err.Error())
		return
	}
	writeResponse(resp, hash.String())
}

func unconfirmedTxHandler(resp http.ResponseWriter, client *btcrpcclient.Client) {
	mtxs, err := models.UnconfirmedTxs()
	if err != nil {
		log.Println(err)
	}
	txs := []*Tx{}
	for _, mtx := range mtxs {
		txs = append(txs, transformTx(&mtx))
	}
	writeResponse(resp, map[string]interface{}{"txs": txs})
}
