package prover

import (
	"encoding/binary"
	"encoding/json"
	"github.com/Lawliet-Chan/roller-go/types"
	"github.com/scroll-tech/go-ethereum/log"
	"net"
)

type Prover struct {
	conn net.Conn
}

func NewProver(path string) *Prover {
	conn, err := net.Dial("unix", path)
	if err != nil {
		log.Crit("init prover failed", "error", err)
	}
	return &Prover{conn: conn}
}

func (p *Prover) Prove(traces *types.BlockTraces) (*types.ZkProof, error) {
	tracesByt, err := encodeTraces(traces)
	if err != nil {
		return nil, err
	}
	_, err = p.conn.Write(tracesByt)
	if err != nil {
		return nil, err
	}

	return p.getZkProof()
}

func (p *Prover) getZkProof() (proof *types.ZkProof, err error) {
	lenByt := make([]byte, 4)
	_, err = p.conn.Read(lenByt)
	if err != nil {
		return
	}
	length := binary.BigEndian.Uint32(lenByt)
	proofByt := make([]byte, length)
	_, err = p.conn.Read(proofByt)
	if err != nil {
		return
	}
	proof = &types.ZkProof{}
	err = json.Unmarshal(proofByt, proof)
	return
}

func encodeTraces(traces *types.BlockTraces) ([]byte, error) {
	bytes, err := json.Marshal(traces)
	if err != nil {
		return nil, err
	}
	length := uint32(len(bytes))
	lenByt := make([]byte, 4)
	binary.BigEndian.PutUint32(lenByt, length)
	return append(lenByt, bytes...), nil
}
