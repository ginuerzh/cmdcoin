// address
package controllers

import (
	"github.com/conformal/btcrpcclient"
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
	m.Get("/newaddr", newAddrHandler)
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
		confirmed, unconfirmed, err := models.FinalBalance(addr, addrs)
		if err != nil {
			log.Println(err)
		}
		addresses = append(addresses, address{Address: addr, Confirmed: confirmed, Unconfirmed: unconfirmed})
	}

	respData := map[string]interface{}{"addresses": addresses}
	writeResponse(resp, respData)
}

type unspentForm struct {
	Addr string `form:"addr"`
	Keys string `form:"keys"`
}

type unspent struct {
	//TxAge   int64  `json:"tx_age"`
	//TxIndex int64  `json:"tx_index"`
	TxHash  string `json:"tx_hash"`
	TxN     uint32 `json:"tx_output_n"`
	Script  string `json:"script"`
	Value   int64  `json:"value"`
	Address string `json:"address"`
}

func unspentHandler(resp http.ResponseWriter, form unspentForm) {
	//keys := strings.Split(form.Keys, "|")

	var unspents []unspent
	var amount int64

	outputs, err := models.AddrOutputs([]string{form.Addr})
	if err != nil {
		log.Println(err)
	}
	for _, output := range outputs {
		if output.BlockHeight == 0 && len(output.Vin) > 0 {
			/*
				find := false
				for _, addr := range keys {
					if addr == output.Vin[0] {
						find = true
						break
					}
				}
				if !find {
					continue
				}
			*/
			if output.Vin[0] != form.Addr {
				continue
			}
		}

		us := unspent{
			TxHash:  output.Txid,
			TxN:     output.Index,
			Script:  output.Script,
			Value:   output.Balance,
			Address: output.Address,
		}
		unspents = append(unspents, us)
		amount += us.Value
	}
	log.Println(amount)

	respData := map[string]interface{}{"unspent_outputs": unspents}
	writeResponse(resp, respData)
}

func newAddrHandler(resp http.ResponseWriter, client *btcrpcclient.Client) {
	status := "ok"

	addr, err := client.GetNewAddress()
	if err != nil {
		status = err.Error()
		writeResponse(resp, map[string]interface{}{"result": status})
		return
	}

	writeResponse(resp, map[string]interface{}{"addr": addr.String(), "result": status})
}
