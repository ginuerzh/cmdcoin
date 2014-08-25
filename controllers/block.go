// block
package controllers

import (
	"github.com/ginuerzh/cmdcoin/models"
	"github.com/martini-contrib/binding"
	"gopkg.in/go-martini/martini.v1"
	"log"
	"net/http"
)

func BindBlockApi(m *martini.ClassicMartini) {
	m.Get("/rawblock", binding.Form(rawBlockForm{}), rawBlockHandler)
	m.Get("/latestblock", latestBlockHandler)
}

type rawBlockForm struct {
	Block string `form:"block"`
}

type block struct {
	Hash      string   `json:"hash"`
	Version   int32    `json:"ver"`
	PrevBlock string   `json:"prev_block"`
	MrklRoot  string   `json:"mrkl_root"`
	Time      int64    `json:"time"`
	Nonce     uint32   `json:"nonce"`
	Height    int64    `json:"height"`
	Tx        []string `json:"tx"`
}

func rawBlockHandler(resp http.ResponseWriter, form rawBlockForm) {
	blk := &models.Block{}
	if err := blk.Find(form.Block); err != nil {
		log.Println(err)
	}
	b := &block{
		Hash:      blk.Id,
		Version:   blk.Version,
		PrevBlock: blk.Prev,
		MrklRoot:  blk.Merkleroot,
		Time:      blk.Time,
		Nonce:     blk.Nonce,
		Height:    blk.Height,
		Tx:        blk.Txs,
	}
	writeResponse(resp, b)
}

type latestBlock struct {
	Hash   string   `json:"hash"`
	Time   int64    `json:"time"`
	Height int64    `json:"height"`
	Txs    []string `json:"txs"`
}

func latestBlockHandler(resp http.ResponseWriter) {
	latest := &models.Block{}
	err := latest.Latest()
	if err != nil {
		log.Println(err)
	}
	blk := &latestBlock{
		Hash:   latest.Id,
		Time:   latest.Time,
		Height: latest.Height,
		Txs:    latest.Txs,
	}

	writeResponse(resp, blk)
}
