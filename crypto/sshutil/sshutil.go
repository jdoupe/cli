package sshutil

import (
	"crypto"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/binary"
	"math/big"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
)

// PublicKey returns the Go's crypto.PublicKey of an ssh.PublicKey.
func PublicKey(key ssh.PublicKey) (crypto.PublicKey, error) {
	_, in, ok := parseString(key.Marshal())
	if !ok {
		return nil, errors.New("public key is invalid")
	}

	switch key.Type() {
	case ssh.KeyAlgoRSA:
		return parseRSA(in)
	case ssh.KeyAlgoECDSA256, ssh.KeyAlgoECDSA384, ssh.KeyAlgoECDSA521:
		return parseECDSA(in)
	case ssh.KeyAlgoED25519:
		return parseED25519(in)
	case ssh.KeyAlgoDSA:
		return parseDSA(in)
	default:
		return nil, errors.Errorf("public key %s is not supported", key.Type())
	}
}

func parseString(in []byte) (out, rest []byte, ok bool) {
	if len(in) < 4 {
		return
	}
	length := binary.BigEndian.Uint32(in)
	in = in[4:]
	if uint32(len(in)) < length {
		return
	}
	out = in[:length]
	rest = in[length:]
	ok = true
	return
}

// parseDSA parses an DSA key according to RFC 4253, section 6.6.
func parseDSA(in []byte) (*dsa.PublicKey, error) {
	var w struct {
		P, Q, G, Y *big.Int
		Rest       []byte `ssh:"rest"`
	}
	if err := ssh.Unmarshal(in, &w); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling public key")
	}

	param := dsa.Parameters{
		P: w.P,
		Q: w.Q,
		G: w.G,
	}
	return &dsa.PublicKey{
		Parameters: param,
		Y:          w.Y,
	}, nil
}

// parseRSA parses an RSA key according to RFC 4253, section 6.6.
func parseRSA(in []byte) (*rsa.PublicKey, error) {
	var w struct {
		E    *big.Int
		N    *big.Int
		Rest []byte `ssh:"rest"`
	}
	if err := ssh.Unmarshal(in, &w); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling public key")
	}
	if w.E.BitLen() > 24 {
		return nil, errors.New("invalid public key: exponent too large")
	}
	e := w.E.Int64()
	if e < 3 || e&1 == 0 {
		return nil, errors.New("invalid public key: incorrect exponent")
	}

	var key rsa.PublicKey
	key.E = int(e)
	key.N = w.N
	return &key, nil
}

// parseECDSA parses an ECDSA key according to RFC 5656, section 3.1.
func parseECDSA(in []byte) (*ecdsa.PublicKey, error) {
	var w struct {
		Curve    string
		KeyBytes []byte
		Rest     []byte `ssh:"rest"`
	}

	if err := ssh.Unmarshal(in, &w); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling public key")
	}

	key := new(ecdsa.PublicKey)

	switch w.Curve {
	case "nistp256":
		key.Curve = elliptic.P256()
	case "nistp384":
		key.Curve = elliptic.P384()
	case "nistp521":
		key.Curve = elliptic.P521()
	default:
		return nil, errors.Errorf("unsupported curve %s", w.Curve)
	}

	key.X, key.Y = elliptic.Unmarshal(key.Curve, w.KeyBytes)
	if key.X == nil || key.Y == nil {
		return nil, errors.New("invalid curve point")
	}

	return key, nil
}

func parseED25519(in []byte) (ed25519.PublicKey, error) {
	var w struct {
		KeyBytes []byte
		Rest     []byte `ssh:"rest"`
	}

	if err := ssh.Unmarshal(in, &w); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling public key")
	}

	return ed25519.PublicKey(w.KeyBytes), nil
}