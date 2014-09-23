// wallet
package controllers

import (
	"github.com/ginuerzh/cmdcoin/models"
	"github.com/martini-contrib/binding"
	"gopkg.in/go-martini/martini.v1"
	"labix.org/v2/mgo/bson"
	"net/http"
)

func BindWalletApi(m *martini.ClassicMartini) {
	m.Get("/wallet", binding.Form(walletForm{}), getWalletHandler)
	m.Post("/wallet", binding.Form(walletForm{}), saveWalletHandler)
}

type walletForm struct {
	Id      string `form:"wallet_id" json:"wallet_id"`
	Payload string `form:"payload" json:"payload"`
}

func getWalletHandler(resp http.ResponseWriter, form walletForm) {
	wallet := &models.Wallet{}
	wallet.Find(form.Id)

	form.Payload = wallet.Payload
	writeResponse(resp, form)
}

func saveWalletHandler(resp http.ResponseWriter, form walletForm) {
	wallet := &models.Wallet{
		Payload: form.Payload,
	}
	if !bson.IsObjectIdHex(form.Id) {
		wallet.Id = bson.NewObjectId()
	} else {
		wallet.Id = bson.ObjectIdHex(form.Id)
	}
	err := wallet.Save()

	status := map[string]string{"wallet_id": wallet.Id.Hex(), "status": "ok"}
	if err != nil {
		status["status"] = err.Error()
	}

	writeResponse(resp, status)
}
