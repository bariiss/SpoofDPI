package packet

import (
	"encoding/binary"
	"fmt"
	"io"
)

type TLSMessageType byte

const (
	TLSMaxPayloadLen uint16         = 16384 // 16 KB
	TLSHeaderLen                    = 5
	TLSHandshake     TLSMessageType = 0x16
)

type TLSMessage struct {
	Header     TLSHeader
	Raw        []byte //Header + Payload
	RawHeader  []byte
	RawPayload []byte
}

type TLSHeader struct {
	Type         TLSMessageType
	ProtoVersion uint16 // major | minor
	PayloadLen   uint16
}

// ReadTLSMessage reads a TLS message from the provided io.Reader.
func ReadTLSMessage(r io.Reader) (*TLSMessage, error) {
	var rawHeader [TLSHeaderLen]byte

	if _, err := io.ReadFull(r, rawHeader[:]); err != nil {
		return nil, err
	}

	header := TLSHeader{
		Type:         TLSMessageType(rawHeader[0]),
		ProtoVersion: binary.BigEndian.Uint16(rawHeader[1:3]),
		PayloadLen:   binary.BigEndian.Uint16(rawHeader[3:5]),
	}

	if header.PayloadLen > TLSMaxPayloadLen {
		return nil, fmt.Errorf(
			"invalid TLS header. Type: %x, ProtoVersion: %x, PayloadLen: %x",
			header.Type,
			header.ProtoVersion,
			header.PayloadLen,
		)
	}

	raw := make([]byte, TLSHeaderLen+int(header.PayloadLen))
	copy(raw[:TLSHeaderLen], rawHeader[:])

	if _, err := io.ReadFull(r, raw[TLSHeaderLen:]); err != nil {
		return nil, err
	}

	return &TLSMessage{
		Header:     header,
		Raw:        raw,
		RawHeader:  raw[:TLSHeaderLen],
		RawPayload: raw[TLSHeaderLen:],
	}, nil
}

// IsClientHello checks if the TLS message is a Client Hello message.
// According to RFC 8446 Section 4:
// TLS handshake message type 0x01 means "client_hello".
func (m *TLSMessage) IsClientHello() bool {
	if len(m.Raw) <= TLSHeaderLen {
		return false
	}
	if m.Header.Type != TLSHandshake {
		return false
	}
	return m.Raw[5] == 0x01
}
