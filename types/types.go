package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"golang.org/x/crypto/blake2s"

	"github.com/scroll-tech/go-ethereum/core/types"
)

// Type denotes the type of message being sent or received.
type Type uint16

const (
	// Error message.
	Error Type = iota
	// Register message, sent by a roller when a connection is established.
	Register
	// BlockTrace message, sent by a sequencer to a roller to notify them
	// they need to generate a proof.
	BlockTrace
	// Proof message, sent by a roller to a sequencer when they have finished
	// proof generation of a given set of block traces.
	Proof
)

// Msg is the top-level message container which contains the payload and the
// message identifier.
type Msg struct {
	// Message type
	Type Type `json:"type"`
	// Message payload
	Payload []byte `json:"payload"`
}

// AuthMessage is the first message exchanged from the Roller to the Sequencer.
// It effectively acts as a registration, and makes the Roller identification
// known to the Sequencer.
type AuthMessage struct {
	// Message fields
	Identity Identity `json:"message"`
	// Roller signature
	Signature string `json:"signature"`
}

// Identity contains all the fields to be signed by the roller.
type Identity struct {
	// Roller name
	Name string `json:"name"`
	// Time of message creation
	Timestamp int64 `json:"timestamp"`
	// Roller public key
	PublicKey string `json:"publicKey"`
}

// Hash returns the hash of the auth message, which should be the message used
// to construct the Signature.
func (i *Identity) Hash() ([]byte, error) {
	bs, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	hash := blake2s.Sum256(bs)
	return hash[:], nil
}

// BlockTraces is a wrapper type around types.BlockResult which adds an ID
// that identifies which proof generation session these block traces are
// associated to. This then allows the roller to add the ID back to their
// proof message once generated, and in turn helps the sequencer understand
// where to handle the proof.
type BlockTraces struct {
	ID     uint64             `json:"id"`
	Traces *types.BlockResult `json:"blockTraces"`
}

// ZkProof is a proof of correct computation for a list of execution traces.
// This is normally sent to the Sequencer by the Roller in order to provide
// the Sequencer with a rollup on request.
type ZkProof struct {
	ID         uint64 `json:"id"`
	EvmProof   []byte `json:"evmProof"`
	StateProof []byte `json:"stateProof"`
}

// Marshal the ZkProof into byte form so it can be sent to the Halo2 verifier.
func (z *ZkProof) Marshal(buf *bytes.Buffer) error {
	// Reserve 4 bytes for message length
	if _, err := buf.Write(make([]byte, 4)); err != nil {
		return err
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, z.ID)
	if _, err := buf.Write(b); err != nil {
		return err
	}

	evmLen := uint32(len(z.EvmProof))
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, evmLen)
	if _, err := buf.Write(b); err != nil {
		return err
	}

	if _, err := buf.Write(z.EvmProof); err != nil {
		return err
	}

	stateLen := uint32(len(z.StateProof))
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, stateLen)
	if _, err := buf.Write(b); err != nil {
		return err
	}

	if _, err := buf.Write(z.StateProof); err != nil {
		return err
	}

	// Now set length bytes
	msgLen := uint32(buf.Len() - 4)
	bytes := buf.Bytes()
	binary.LittleEndian.PutUint32(bytes[:4], msgLen)

	return nil
}
