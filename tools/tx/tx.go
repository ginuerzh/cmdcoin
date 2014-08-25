// tx
package tx

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/conformal/btcec"
)

type VarInt uint64

func (vi VarInt) Len() int {
	if vi < 0xfd {
		return 1
	} else if vi < 0xffff {
		return 3
	} else if vi < 0xffffffff {
		return 5
	} else {
		return 9
	}
}

func (vi VarInt) Value() uint64 {
	return uint64(vi)
}

func (vi VarInt) MarshalBinary() (data []byte, err error) {
	if vi < 0xfd {
		data = []byte{byte(vi)}
	} else if vi < 0xffff {
		data = make([]byte, 3)
		data[0] = 0xfd
		binary.LittleEndian.PutUint16(data[1:], uint16(vi))
	} else if vi < 0xffffffff {
		data = make([]byte, 5)
		data[0] = 0xfe
		binary.LittleEndian.PutUint32(data[1:], uint32(vi))
	} else {
		data = make([]byte, 9)
		data[0] = 0xff
		binary.LittleEndian.PutUint64(data[1:], uint64(vi))
	}
	return
}

func (vi *VarInt) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	h := data[0]
	if h < 0xfd {
		*vi = VarInt(h)
	} else if h == 0xfd {
		*vi = VarInt(binary.LittleEndian.Uint16(data[1:3]))
	} else if h == 0xfe {
		*vi = VarInt(binary.LittleEndian.Uint32(data[1:5]))
	} else if h == 0xff {
		*vi = VarInt(binary.LittleEndian.Uint64(data[1:9]))
	} else {
		return errors.New("Unknown binary format")
	}
	return nil
}

type Input struct {
	Txid      string
	Index     uint32
	SigLen    VarInt
	ScriptSig string
	Sequence  uint32
}

func (in Input) Len() int {
	return 32 + 4 + in.SigLen.Len() + int(in.SigLen.Value()) + 4
}

func (in Input) MarshalBinary() (data []byte, err error) {
	b, err := hex.DecodeString(in.Txid)
	if err != nil {
		return
	}
	data = append(data, b...)

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, in.Index)
	data = append(data, b...)

	b, err = in.SigLen.MarshalBinary()
	if err != nil {
		return
	}
	data = append(data, b...)

	b, err = hex.DecodeString(in.ScriptSig)
	if err != nil {
		return
	}
	data = append(data, b...)

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, in.Sequence)
	data = append(data, b...)

	return
}

func (in *Input) UnmarshalBinary(data []byte) error {
	pos := 0
	in.Txid = hex.EncodeToString(data[pos : pos+32])

	pos += 32
	in.Index = binary.LittleEndian.Uint32(data[pos : pos+4])

	pos += 4
	if err := in.SigLen.UnmarshalBinary(data[pos : pos+9]); err != nil {
		return err
	}

	pos += in.SigLen.Len()
	in.ScriptSig = hex.EncodeToString(data[pos : pos+int(in.SigLen.Value())])

	pos += int(in.SigLen.Value())
	in.Sequence = binary.LittleEndian.Uint32(data[pos : pos+4])

	return nil
}

type Output struct {
	Value     int64
	ScriptLen VarInt
	Script    string
}

func (op Output) Len() int {
	return 8 + op.ScriptLen.Len() + int(op.ScriptLen.Value())
}

func (op Output) MarshalBinary() (data []byte, err error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(op.Value))
	data = append(data, b...)

	b, err = op.ScriptLen.MarshalBinary()
	if err != nil {
		return
	}
	data = append(data, b...)

	b, err = hex.DecodeString(op.Script)
	if err != nil {
		return
	}
	data = append(data, b...)

	return
}

func (op *Output) UnmarshalBinary(data []byte) error {
	pos := 0
	op.Value = int64(binary.LittleEndian.Uint64(data[pos : pos+8]))

	pos += 8
	if err := op.ScriptLen.UnmarshalBinary(data[pos : pos+9]); err != nil {
		return err
	}

	pos += op.ScriptLen.Len()
	op.Script = hex.EncodeToString(data[pos : pos+int(op.ScriptLen.Value())])

	return nil
}

type Tx struct {
	Version  uint32
	NInputs  VarInt
	Inputs   []Input
	NOutputs VarInt
	Outputs  []Output
	LockTime uint32
	//HashType  uint32
}

func (tx Tx) Len() int {
	n := 4 + tx.NInputs.Len()
	for i, _ := range tx.Inputs {
		n += tx.Inputs[i].Len()
	}
	n += tx.NOutputs.Len()
	for i, _ := range tx.Outputs {
		n += tx.Outputs[i].Len()
	}
	n += 4

	return n
}

func (tx Tx) MarshalBinary() (data []byte, err error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, tx.Version)
	data = append(data, b...)

	b, err = tx.NInputs.MarshalBinary()
	if err != nil {
		return
	}
	data = append(data, b...)

	for _, input := range tx.Inputs {
		b, err = input.MarshalBinary()
		if err != nil {
			return
		}
		data = append(data, b...)
	}

	b, err = tx.NOutputs.MarshalBinary()
	if err != nil {
		return
	}
	data = append(data, b...)

	for _, output := range tx.Outputs {
		b, err = output.MarshalBinary()
		if err != nil {
			return
		}
		data = append(data, b...)
	}

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, tx.LockTime)
	data = append(data, b...)

	return
}

func (tx *Tx) UnmarshalBinary(data []byte) error {
	pos := 0
	tx.Version = binary.LittleEndian.Uint32(data[pos : pos+4])

	pos += 4
	if err := tx.NInputs.UnmarshalBinary(data[pos : pos+9]); err != nil {
		return err
	}

	pos += tx.NInputs.Len()
	for i := 0; i < int(tx.NInputs.Value()); i++ {
		input := Input{}
		if err := input.UnmarshalBinary(data[pos:]); err != nil {
			return err
		}
		tx.Inputs = append(tx.Inputs, input)
		pos += input.Len()
	}

	if err := tx.NOutputs.UnmarshalBinary(data[pos : pos+9]); err != nil {
		return err
	}

	pos += tx.NOutputs.Len()
	for i := 0; i < int(tx.NOutputs.Value()); i++ {
		output := Output{}
		if err := output.UnmarshalBinary(data[pos:]); err != nil {
			return err
		}
		tx.Outputs = append(tx.Outputs, output)
		pos += output.Len()
	}

	tx.LockTime = binary.LittleEndian.Uint32(data[pos : pos+4])

	return nil
}

func (tx Tx) Hash() string {
	b, err := tx.MarshalBinary()
	if err != nil {
		return ""
	}

	sha := sha256.Sum256(b)
	b = sha[:]
	sha = sha256.Sum256(b)
	b = sha[:]
	return hex.EncodeToString(b)
}

func (tx *Tx) Sign(privKey string) (string, error) {
	pkBytes, err := hex.DecodeString(privKey)
	if err != nil {
		return "", err
	}
	priv, pub := btcec.PrivKeyFromBytes(btcec.S256(), pkBytes)

	hash, err := hex.DecodeString(tx.Hash())
	if err != nil {
		return "", err
	}
	sig, err := priv.Sign(hash)
	if err != nil {
		return "", err
	}

	fmt.Println("pubkey uncompressed:", hex.EncodeToString(pub.SerializeUncompressed()))
	fmt.Println("pubkey compressed:", hex.EncodeToString(pub.SerializeCompressed()))

	return hex.EncodeToString(sig.Serialize()), nil
}
