package mock

import (
	"encoding/binary"
	"encoding/json"
	"github.com/Lawliet-Chan/roller-go/config"
	"github.com/Lawliet-Chan/roller-go/roller"
	"github.com/Lawliet-Chan/roller-go/types"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/gorilla/websocket"
	"github.com/scroll-tech/go-ethereum/common"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestRoller(t *testing.T) {
	cfg := &config.Config{
		RollerName:       "my-roller",
		SecretKey:        "dcf2cbdd171a21c480aa7f53d77f31bb102282b3ff099c78e3118b37348c72f7",
		ScrollUrl:        "ws://localhost:9000",
		ProverSocketPath: "/tmp/prover.sock",
		StackPath:        "stack",
		WsTimeoutSec:     10,
	}
	go mockIpcProver(t, cfg.ProverSocketPath)
	go mockScroll(t)
	r := roller.NewRoller(cfg)
	go r.Run()
	time.Sleep(2 * time.Second)
	defer func() {
		_ = os.Remove(cfg.ProverSocketPath)
		r.Close()
	}()
}

func mockScroll(t *testing.T) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		up := websocket.Upgrader{}
		c, err := up.Upgrade(w, req, nil)
		if err != nil {
			t.Fatal(err)
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
			t.Fatal(err)
		}
		authMsg := &types.AuthMessage{}
		if err := json.Unmarshal(msg.Payload, authMsg); err != nil {
			t.Fatal(err)
		}

		// Verify signature
		hash, err := authMsg.Identity.Hash()
		if err != nil {
			t.Fatal(err)
		}
		if !secp256k1.VerifySignature(common.FromHex(authMsg.Identity.PublicKey), hash, common.FromHex(authMsg.Signature)[:64]) {
			t.Fatal("verify signer failed: " + err.Error())
		}
		println("signature verification successful. Roller: ", authMsg.Identity.Name)

		traces := &types.BlockTraces{
			ID:     16,
			Traces: nil,
		}
		msgByt, err := roller.MakeMsgByt(types.BlockTrace, traces)
		if err != nil {
			t.Fatal(err)
		}
		err = c.WriteMessage(websocket.BinaryMessage, msgByt)
		if err != nil {
			t.Fatal(err)
		}
	})
	http.ListenAndServe(":9000", nil)
}

func mockIpcProver(t *testing.T, socket string) {
	lis, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4)
		_, err = conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		bytesLen := binary.BigEndian.Uint32(buf)
		jsonBuf := make([]byte, bytesLen)
		_, err = conn.Read(jsonBuf)
		if err != nil {
			t.Fatal(err)
		}

		zkproof := &types.ZkProof{
			ID:         0,
			EvmProof:   nil,
			StateProof: nil,
		}
		proofByt, err := json.Marshal(zkproof)
		if err != nil {
			t.Fatal(err)
		}
		length := uint32(len(proofByt))
		lenByt := make([]byte, 4)
		binary.BigEndian.PutUint32(lenByt, length)

		_, err = conn.Write(append(lenByt, proofByt...))
		if err != nil {
			t.Fatal(err)
		}
	}
}
