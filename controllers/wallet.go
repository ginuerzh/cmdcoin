// wallet
package controllers

import (
	"github.com/ginuerzh/cmdcoin/models"
	"github.com/martini-contrib/binding"
	"gopkg.in/go-martini/martini.v1"
	"net/http"
)

func BindWalletApi(m *martini.ClassicMartini) {
	m.Get("/wallet", binding.Form(walletGetForm{}), getWalletHandler)
	m.Post("/wallet", binding.Form(walletSaveForm{}), saveWalletHandler)
}

type walletGetForm struct {
	Id string `form:"wallet_id"`
}

func getWalletHandler(resp http.ResponseWriter, form walletGetForm) {
	wallet := &models.Wallet{WalletId: form.Id}
	if len(form.Id) > 0 {
		wallet.Find(form.Id)
	}

	writeResponse(resp, wallet)
}

type walletSaveForm struct {
	Id      string `form:"wallet_id"`
	Payload string `form:"payload"`
}

func saveWalletHandler(resp http.ResponseWriter, form walletSaveForm) {
	wallet := &models.Wallet{
		WalletId: form.Id,
		Payload:  form.Payload,
	}
	err := wallet.Save()

	status := map[string]string{"status": "ok"}
	if err != nil {
		status["status"] = err.Error()
	}

	writeResponse(resp, status)
}
