// transaction
package controllers

import (
	"bytes"
	"encoding/hex"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/ginuerzh/cmdcoin/models"
	"github.com/martini-contrib/binding"
	"gopkg.in/go-martini/martini.v1"
	"log"
	"net/http"
	"strings"
)

func BindTxApi(m *martini.ClassicMartini) {
	m.Get("/rawtx", binding.Form(rawTxForm{}), rawTxHandler)
	m.Get("/addr_txs", binding.Form(addrTxsForm{}), addrTxsHandler)
	m.Post("/pushtx", binding.Form(pushTxForm{}), pushTxHandler)
	m.Get("/unconfirmed_tx", unconfirmedTxHandler)
	m.Post("/send", binding.Form(sendCoinForm{}), sendCoinHandler)
}

type rawTxForm struct {
	Tx string `form:"tx"`
}

func rawTxHandler(resp http.ResponseWriter, form rawTxForm) {
	tx := &models.Tx{}
	if err := tx.Find(form.Tx); err != nil {
		log.Println(err)
	}

	writeResponse(resp, tx)
}

type addrTxsForm struct {
	Addrs string `form:"addr"`
}

func addrTxsHandler(resp http.ResponseWriter, form addrTxsForm) {
	addrs := strings.Split(form.Addrs, "|")
	//log.Println(addrs)
	txs, err := models.AddrTxs(addrs)
	if err != nil {
		log.Println(err)
	}
	writeResponse(resp, txs)
}

type pushTxForm struct {
	RawTx string `form:"rawtx"`
}

func pushTxHandler(resp http.ResponseWriter, client *btcrpcclient.Client, form pushTxForm) {
	msgTx := btcwire.NewMsgTx()
	//log.Println("pushtx:", form.RawTx)
	b, err := hex.DecodeString(form.RawTx)
	if err != nil {
		writeResponse(resp, map[string]string{"error": err.Error()})
		return
	}
	if err := msgTx.BtcDecode(bytes.NewBuffer(b), 1); err != nil {
		writeResponse(resp, map[string]string{"error": err.Error()})
		return
	}
	hash, err := client.SendRawTransaction(msgTx, false)
	if err != nil {
		writeResponse(resp, map[string]string{"error": err.Error()})
		return
	}
	writeResponse(resp, map[string]string{"txid": hash.String()})
}

func unconfirmedTxHandler(resp http.ResponseWriter, client *btcrpcclient.Client) {
	txs, err := models.UnconfirmedTxs()
	if err != nil {
		log.Println(err)
	}

	writeResponse(resp, map[string]interface{}{"txs": txs})
}

type sendCoinForm struct {
	ToAddr string `form:"to"`
	Amount int64  `form:"amount"`
}

func sendCoinHandler(resp http.ResponseWriter, client *btcrpcclient.Client, form sendCoinForm) {
	status := "ok"
	addr, err := btcutil.DecodeAddress(form.ToAddr, &btcnet.MainNetParams)
	if err != nil {
		status = err.Error()
		writeResponse(resp, map[string]interface{}{"result": status})
		return
	}
	sha, err := client.SendToAddress(addr, btcutil.Amount(form.Amount))
	if err != nil {
		status = err.Error()
		writeResponse(resp, map[string]interface{}{"result": status})
		return
	}

	writeResponse(resp, map[string]interface{}{"txid": sha.String(), "result": status})
}
