// address
package controllers

import (
	"github.com/ginuerzh/cmdcoin/models"
	"github.com/martini-contrib/binding"
	"gopkg.in/go-martini/martini.v1"
	"log"
	"net/http"
	"strings"
)

func BindAddrApi(m *martini.ClassicMartini) {
	m.Get("/multiaddr", binding.Form(multiAddrForm{}), multiAddrHandler)
	m.Get("/unspent", binding.Form(unspentForm{}), unspentHandler)
}

type multiAddrForm struct {
	Addrs string `form:"addr"`
}

type address struct {
	Address     string `json:"address"`
	Confirmed   int64  `json:"confirmed"`
	Unconfirmed int64  `json:"unconfirmed"`
}

func multiAddrHandler(resp http.ResponseWriter, form multiAddrForm) {
	addrs := strings.Split(form.Addrs, "|")
	var addresses []address

	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		confirmed, unconfirmed, err := models.FinalBalance(addr)
		if err != nil {
			log.Println(err)
		}
		addresses = append(addresses, address{Address: addr, Confirmed: confirmed, Unconfirmed: unconfirmed})
	}

	respData := map[string]interface{}{"addresses": addresses}
	writeResponse(resp, respData)
}

type unspentForm struct {
	Addrs string `form:"addr"`
}

type unspent struct {
	//TxAge   int64  `json:"tx_age"`
	//TxIndex int64  `json:"tx_index"`
	TxHash string `json:"tx_hash"`
	TxN    uint32 `json:"tx_output_n"`
	Script string `json:"script"`
	Value  int64  `json:"value"`
}

func unspentHandler(resp http.ResponseWriter, form unspentForm) {
	addrs := strings.Split(form.Addrs, "|")
	var unspents []unspent

	outputs, err := models.AddrOutputs(addrs)
	if err != nil {
		log.Println(err)
	}
	for _, output := range outputs {
		if output.BlockHeight <= 0 {
			continue
		}
		us := unspent{
			TxHash: output.Txid,
			TxN:    output.Index,
			Script: output.Script,
			Value:  output.Balance,
		}
		unspents = append(unspents, us)
	}

	respData := map[string]interface{}{"unspent_outputs": unspents}
	writeResponse(resp, respData)
}
