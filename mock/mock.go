package main

import (
	"encoding/binary"
	"encoding/json"
	"github.com/Lawliet-Chan/roller-go/config"
	"github.com/Lawliet-Chan/roller-go/roller"
	"github.com/Lawliet-Chan/roller-go/types"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gorilla/websocket"
	"github.com/scroll-tech/go-ethereum/common"
	"github.com/scroll-tech/go-ethereum/log"
	"net"
	"net/http"
)

func main() {
	cfg := &config.Config{
		RollerName:   "my-roller",
		Secret:       nil,
		ScrollUrl:    "ws://localhost:9000",
		ProverPath:   "/tmp/prover.sock",
		StackPath:    "stack",
		WsTimeoutSec: 10,
	}
	go mockIpcProver(cfg.ProverPath)
	go mockScroll()
	r := roller.NewRoller(cfg)
	r.Run()
	defer r.Close()
}

func mockScroll() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		up := websocket.Upgrader{}
		c, err := up.Upgrade(w, req, nil)
		if err != nil {
			panic(err)
		}
		payload, err := func(c *websocket.Conn) ([]byte, error) {
			for {
				mt, payload, err := c.ReadMessage()
				if err != nil {
					return nil, err
				}

				if mt == websocket.BinaryMessage {
					return payload, nil
				}
			}
		}(c)

		msg := &types.Msg{}
		if err = json.Unmarshal(payload, msg); err != nil {
			panic(err)
		}
		authMsg := &types.AuthMessage{}
		if err := json.Unmarshal(msg.Payload, authMsg); err != nil {
			panic(err)
		}

		// Verify signature
		hash, err := authMsg.Identity.Hash()
		if err != nil {
			panic(err)
		}
		if !secp256k1.VerifySignature(common.FromHex(authMsg.Identity.PublicKey), hash, common.FromHex(authMsg.Signature)[:64]) {
			panic("verify signer failed: " + err.Error())
		}
		log.Info("signature verification successful", "roller name", authMsg.Identity.Name)

		traces := &types.BlockTraces{
			ID:     1,
			Traces: nil,
		}
		msgByt, err := roller.MakeMsgByt(types.BlockTrace, traces)
		if err != nil {
			panic(err)
		}
		err = c.WriteMessage(websocket.BinaryMessage, msgByt)
		if err != nil {
			panic(err)
		}
	})
	http.ListenAndServe(":9000", nil)
}

func mockIpcProver(socket string) {
	lis, err := net.Listen("unix", socket)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}
		buf := make([]byte, 4)
		_, err = conn.Read(buf)
		if err != nil {
			panic(err)
		}
		bytesLen := binary.BigEndian.Uint32(buf)
		jsonBuf := make([]byte, bytesLen)
		_, err = conn.Read(jsonBuf)
		if err != nil {
			panic(err)
		}

		zkproof := &types.ZkProof{
			ID:         0,
			EvmProof:   nil,
			StateProof: nil,
		}
		proofByt, err := json.Marshal(zkproof)
		if err != nil {
			panic(err)
		}
		length := uint32(len(proofByt))
		lenByt := make([]byte, 4)
		binary.BigEndian.PutUint32(lenByt, length)

		_, err = conn.Write(append(lenByt, proofByt...))
		if err != nil {
			panic(err)
		}
	}
}
