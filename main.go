// main
package main

import (
	"flag"
	"github.com/conformal/btcrpcclient"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/ginuerzh/cmdcoin/controllers"
	"gopkg.in/go-martini/martini.v1"
	"log"
	"net/http"
)

var (
	staticDir  string
	listenAddr string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&staticDir, "static", "public", "static files directory")
	flag.StringVar(&listenAddr, "l", ":8087", "addr on listen")
	flag.Parse()
}

func main() {
	m := martini.Classic()
	m.Use(controllers.DumpReqBodyHandler)
	m.Map(btcRpcClient())

	controllers.BindBlockApi(m)
	controllers.BindTxApi(m)
	controllers.BindAddrApi(m)
	controllers.BindWalletApi(m)

	//http.ListenAndServe(listenAddr, m)
	gracehttp.Serve(&http.Server{Addr: listenAddr, Handler: m})
}

func btcRpcClient() *btcrpcclient.Client {
	cfg := &btcrpcclient.ConnConfig{
		Host:         "localhost:8110",
		User:         "btcrpc",
		Pass:         "pbtcrpc",
		DisableTLS:   true,
		HttpPostMode: true,
	}
	client, err := btcrpcclient.New(cfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
