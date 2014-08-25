// tx_test
package tx

import (
	"encoding/hex"
	"testing"
)

const (
	txHex = "01000000" + // version
		"01" + // n inputs
		"26c07ece0bce7cda0ccd14d99e205f118cde27e83dd75da7b141fe487b5528fb" + //prev txid
		"00000000" + // output index
		"8b" + // length of scriptSig
		"48304502202b7e37831273d74c8b5b1956c23e79acd660635a8d1063d413c50b218eb6bc8a022100a10a3a7b5aaa0f07827207daf81f718f51eeac96695cf1ef9f2020f21a0de02f01410452684bce6797a0a50d028e9632be0c2a7e5031b710972c2a3285520fb29fcd4ecfb5fc2bf86a1e7578e4f8a305eeb341d1c6fc0173e5837e2d3c7b178aade078" +
		"ffffffff" + // sequence
		"02" + // n outputs
		"b06c191e01000000" + // amount of ouput 1
		"19" + // length of output 1 script
		"76a9143564a74f9ddb4372301c49154605573d7d1a88fe88ac" +
		"00e1f50500000000" + // amount of ouput 2
		"19" + // length of output 2 script
		"76a914010966776006953d5567439e5e39f86a0d273bee88ac" +
		"00000000" // lock time

	privKey    = "18E14A7B6A307F426A94F8114701E7C8E774E7F9A47E2C2035DB29A206321725"
	sendToAddr = "1runeksijzfVxyrpiyCY2LCBvYsSiFsCm"
)

func TestDecodeTx(t *testing.T) {
	b, err := hex.DecodeString(txHex)
	if err != nil {
		t.Fatal(err)
	}

	tx := &Tx{}
	if err = tx.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", tx)

	b, err = tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(b) == txHex)
	t.Log(tx.Hash())
}

func TestSignTx(t *testing.T) {
	b, err := hex.DecodeString(txHex)
	if err != nil {
		t.Fatal(err)
	}

	tx := &Tx{}
	if err = tx.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}

	sigMsg, err := tx.Sign(privKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(sigMsg)
}
