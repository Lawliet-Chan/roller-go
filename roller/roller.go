package roller

import (
	"encoding/json"
	"github.com/Lawliet-Chan/roller-go/config"
	"github.com/Lawliet-Chan/roller-go/prover"
	"github.com/Lawliet-Chan/roller-go/stack"
	"github.com/Lawliet-Chan/roller-go/types"
	"github.com/gorilla/websocket"
	"github.com/scroll-tech/go-ethereum/common"
	"github.com/scroll-tech/go-ethereum/log"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"net"
	"time"
)

type Roller struct {
	cfg    *config.Config
	conn   *websocket.Conn
	stack  *stack.Stack
	prover *prover.Prover
}

func NewRoller(cfg *config.Config) *Roller {
	conn, _, err := websocket.DefaultDialer.Dial(cfg.ScrollUrl, nil)
	if err != nil {
		log.Crit("websocket connects failed", "error", err)
	}

	return &Roller{
		cfg:    cfg,
		conn:   conn,
		stack:  stack.NewStack(cfg.StackPath),
		prover: prover.NewProver(cfg.ProverPath),
	}
}

func (r *Roller) Run() {
	err := r.Register()
	if err != nil {
		log.Crit("register to scroll failed", "error", err)
	}
	go r.HandleScroll()
	r.Prove()
}

func (r *Roller) Register() error {
	prvkey := secp256k1.GenPrivKeySecp256k1(r.cfg.Secret)
	pubkey := prvkey.PubKey().Bytes()
	authMsg := &types.AuthMessage{
		Identity: types.Identity{
			Name:      "testRoller",
			Timestamp: time.Now().UnixMilli(),
			PublicKey: common.Bytes2Hex(pubkey),
		},
		Signature: "",
	}

	hash, err := authMsg.Identity.Hash()
	if err != nil {
		return err
	}

	sig, err := prvkey.Sign(hash)
	if err != nil {
		return err
	}
	authMsg.Signature = common.Bytes2Hex(sig)

	msgByt, err := makeMsgByt(types.Register, authMsg)
	if err != nil {
		return err
	}

	return r.conn.WriteMessage(websocket.BinaryMessage, msgByt)
}

func (r *Roller) HandleScroll() {
	for {
		r.conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(r.cfg.WsTimeoutSec)))
		r.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(r.cfg.WsTimeoutSec)))
		err := r.handleScroll()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				panic(err)
			}
			log.Error("handle scroll failed", "error", err)
		}
	}
}

func (r *Roller) Prove() {
	for {
		r.conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(r.cfg.WsTimeoutSec)))
		err := r.prove()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				panic(err)
			}
			log.Error("prove failed", "error", err)
		}
	}
}

func (r *Roller) handleScroll() error {
	mt, msg, err := r.conn.ReadMessage()
	if err != nil {
		return err
	}

	switch mt {
	case websocket.BinaryMessage:
		err = r.persistTrace(msg)
		if err != nil {
			return err
		}
	case websocket.PingMessage:
		log.Debug("receive heartbeat!")
		err = r.conn.WriteMessage(websocket.PongMessage, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Roller) prove() error {
	traces, err := r.stack.Pop()
	if err != nil {
		return err
	}
	proof, err := r.prover.Prove(traces)
	if err != nil {
		return err
	}
	msgByt, err := makeMsgByt(types.Proof, proof)
	if err != nil {
		return err
	}
	return r.conn.WriteMessage(websocket.BinaryMessage, msgByt)
}

func (r *Roller) Close() {
	r.conn.Close()
}

func (r *Roller) persistTrace(byt []byte) error {
	var msg = &types.Msg{}
	err := json.Unmarshal(byt, msg)
	if err != nil {
		return err
	}
	var traces = &types.BlockTraces{}
	err = json.Unmarshal(msg.Payload, traces)
	if err != nil {
		return err
	}
	return r.stack.Append(traces)
}

func makeMsgByt(msgTyp types.Type, payloadVal interface{}) ([]byte, error) {
	payload, err := json.Marshal(payloadVal)
	if err != nil {
		return nil, err
	}
	msg := &types.Msg{
		Type:    msgTyp,
		Payload: payload,
	}
	return json.Marshal(msg)
}
