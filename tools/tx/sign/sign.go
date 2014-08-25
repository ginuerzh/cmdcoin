// sign
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/conformal/btcec"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"log"
)

const (
	/*
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
	*/
	txHex       = "0100000001fda04ce364a4bbbbcfbf151ffad412c6d62e6a1b87d41c7a7c2e29d01b120871000000006c49304602210081815d8405ca5774563d4c36435c4b8634e28ab71891ce38242edd2a03763d51022100bd9af692063168c4fa4e9a94cd490e4aee74d9ccf04e060bae923829dd2d5003012103c616c029a335f895add6942fdc92f80ff7a51eecb3f668de2965bd37e786930affffffff0200e1f505000000001976a9149c77061d85cc39bf3d9717b6d545fabc277cf52088ac00e9280d180000001976a9147cdf96ae5588a0baa3e8e79c20ea3daf6460fd0388ac00000000"
	txFee       = 100000
	privKey     = "5JRqiCuujPYZL7gxqVoGnn7aFwoNZPTLJf5ksAMksqqjaXvKo7W"
	sendToAddr  = "1PqF4Mmaq9LZ2JznppsamCQCfLbCxnsKGh"
	outputIndex = 0

	SIGHASH_ALL = 1
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func dsha256(data []byte) []byte {
	return btcwire.DoubleSha256(data)
}

func parseTx(txHex string) (*btcwire.MsgTx, error) {
	msgTx := btcwire.NewMsgTx()
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}
	if err = msgTx.BtcDecode(bytes.NewBuffer(txBytes), 1); err != nil {
		return nil, err
	}
	return msgTx, nil
}

func makeScriptPubKey(toAddr string) ([]byte, error) {
	addr, err := btcutil.DecodeAddress(toAddr, &btcnet.MainNetParams)
	if err != nil {
		return nil, err
	}
	log.Println("script addr:", hex.EncodeToString(addr.ScriptAddress()))
	builder := btcscript.NewScriptBuilder()
	builder.AddOp(btcscript.OP_DUP).AddOp(btcscript.OP_HASH160)
	builder.AddData(addr.ScriptAddress())
	builder.AddOp(btcscript.OP_EQUALVERIFY).AddOp(btcscript.OP_CHECKSIG)
	//script := "76" + "a9" + "14" + hex.EncodeToString(addr.ScriptAddress()) + "88" + "ac"

	return builder.Script(), nil
}

func makeTx(prevMsgTx *btcwire.MsgTx, index uint32) (*btcwire.MsgTx, error) {
	msgTx := btcwire.NewMsgTx()
	hash, err := prevMsgTx.TxSha()
	if err != nil {
		return nil, err
	}
	log.Println("prev tx hash:", hash)
	txIn := btcwire.NewTxIn(btcwire.NewOutPoint(&hash, index), prevMsgTx.TxOut[index].PkScript)
	log.Println("prev script:", hex.EncodeToString(prevMsgTx.TxOut[index].PkScript))
	msgTx.AddTxIn(txIn)

	script, err := makeScriptPubKey(sendToAddr)
	if err != nil {
		return nil, err
	}
	txOut := btcwire.NewTxOut(prevMsgTx.TxOut[index].Value-txFee, script)
	log.Println("pay value:", prevMsgTx.TxOut[index].Value, txFee, txOut.Value)
	msgTx.AddTxOut(txOut)

	return msgTx, nil
}

func signScript(tx *btcwire.MsgTx, idx int, subscript []byte, privKey string) ([]byte, error) {
	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return nil, err
	}

	return btcscript.SignatureScript(tx, idx, subscript, SIGHASH_ALL, wif.PrivKey.ToECDSA(), wif.CompressPubKey)
}

func sigTx(msgTx *btcwire.MsgTx, privKey string) (*btcec.PublicKey, *btcec.Signature, error) {
	sha, err := msgTx.TxSha()
	if err != nil {
		return nil, nil, err
	}

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return nil, nil, err
	}

	priv, pub := btcec.PrivKeyFromBytes(btcec.S256(), wif.PrivKey.Serialize())
	log.Println("privkey:", hex.EncodeToString(priv.Serialize()))

	sig, err := priv.Sign(sha.Bytes())
	if err != nil {
		return pub, nil, err
	}
	return pub, sig, nil
}

func signScript2(tx *btcwire.MsgTx, subscript []byte, privKey string) ([]byte, error) {
	pubKey, sig, err := sigTx(tx, privKey)
	if err != nil {
		return nil, err
	}
	sigData := append(sig.Serialize(), SIGHASH_ALL) // append hash type SIGHASH_ALL(1) to sign data

	log.Println("sig data:", hex.EncodeToString(sigData))

	addr, err := btcutil.NewAddressPubKey(pubKey.SerializeUncompressed(), &btcnet.MainNetParams)
	if err != nil {
		return nil, err
	}
	log.Println("from addr:", addr.EncodeAddress())

	srcAddr, _ := btcutil.DecodeAddress(addr.EncodeAddress(), &btcnet.MainNetParams)

	if !bytes.Equal(srcAddr.ScriptAddress(), subscript[3:len(subscript)-2]) {
		return nil, errors.New("The supplied private key cannot be used to redeem output")
	}

	builder := btcscript.NewScriptBuilder()
	builder.AddData(sigData)
	builder.AddData(pubKey.SerializeUncompressed())

	return builder.Script(), nil
}

func main() {
	prevMsgTx := btcwire.NewMsgTx()
	b, err := hex.DecodeString(txHex)
	if err != nil {
		log.Fatal(err)
	}
	if err := prevMsgTx.BtcDecode(bytes.NewBuffer(b), 1); err != nil {
		log.Fatal(err)
	}
	msgTx, err := makeTx(prevMsgTx, outputIndex)
	if err != nil {
		log.Fatal(err)
	}

	buffer := &bytes.Buffer{}
	if err := msgTx.BtcEncode(buffer, 1); err != nil {
		log.Fatal(err)
	}
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, SIGHASH_ALL) // append hash type SIGHASH_ALL(1)
	b = append(buffer.Bytes(), b...)
	log.Println("tx:", hex.EncodeToString(b))
	hashScriptless := dsha256(b)
	log.Println("hash_scriptless:", hex.EncodeToString(hashScriptless))

	finalTx := msgTx.Copy()

	scriptSig, err := signScript(msgTx, outputIndex, prevMsgTx.TxOut[outputIndex].PkScript, privKey)

	//scriptSig, err := signScript2(msgTx, prevMsgTx.TxOut[outputIndex].PkScript, privKey)
	if err != nil {
		log.Fatal(err)
	}

	finalTx.TxIn[0].SignatureScript = scriptSig
	log.Println("scriptSig:", hex.EncodeToString(scriptSig))

	buffer = &bytes.Buffer{}
	if err := finalTx.BtcEncode(buffer, 1); err != nil {
		log.Fatal(err)
	}

	log.Println("final tx:", hex.EncodeToString(buffer.Bytes()))
	sha, _ := finalTx.TxSha()
	log.Println("final tx hash:", sha)

}
