//go:build webtransport

package webtransportrelay

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

func writeProtoFrame(w io.Writer, msg proto.Message) error {
	if msg == nil {
		return fmt.Errorf("nil proto message")
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return writeFrame(w, b)
}

func readProtoFrame(r io.Reader, msg proto.Message, maxBytes int) error {
	if msg == nil {
		return fmt.Errorf("nil proto message")
	}
	b, err := readFrame(r, maxBytes)
	if err != nil {
		return err
	}
	return proto.Unmarshal(b, msg)
}

func writeFrame(w io.Writer, payload []byte) error {
	if len(payload) == 0 {
		return fmt.Errorf("empty frame")
	}
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))
	if _, err := w.Write(lenBuf[:]); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func readFrame(r io.Reader, maxBytes int) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf[:])
	if length == 0 {
		return nil, fmt.Errorf("frame length 0")
	}
	if maxBytes > 0 && int(length) > maxBytes {
		return nil, fmt.Errorf("frame too large: %d > %d", length, maxBytes)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
