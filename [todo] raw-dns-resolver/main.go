package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

func main() {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	tv := unix.NsecToTimeval(10 * 1e9)
	err = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &tv)
	if err != nil {
		panic(err)
	}

	query, err := newARecordQuery("example.com")
	if err != nil {
		panic(err)
	}

	fmt.Println(query.Bytes())

	err = unix.Sendto(fd, query.Bytes(), 0, &unix.SockaddrInet4{
		Port: 53,
		Addr: [4]byte{1, 1, 1, 1},
	})
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 4096)
	n, _, err := unix.Recvfrom(fd, buf, 0)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf[:n])
}

const (
	maxUint16 = 1<<16 - 1
)

type Question struct {
	QNAME  []byte
	QTYPE  uint16
	QCLASS uint16
}

func (q *Question) Len() int {
	return len(q.QNAME) + 4
}

type Packet struct {
	// Header
	ID        uint16
	Flags     uint16
	Questions []Question
}

func (p *Packet) Valid() error {
	if len(p.Questions) > maxUint16 {
		return errors.New("error questions exceeded maximum length")
	}

	return nil
}

func newPacket(id uint16, flags uint16, questions []Question) (*Packet, error) {
	p := &Packet{
		ID:        id,
		Flags:     flags,
		Questions: questions,
	}

	err := p.Valid()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func newARecordQuery(hostname string) (*Packet, error) {
	qName, err := encodeHostname(hostname)
	if err != nil {
		return nil, err
	}

	packet, err := newPacket(1, 0x0100, []Question{
		{
			QNAME:  qName,
			QTYPE:  1,
			QCLASS: 1,
		},
	})
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func (p *Packet) Bytes() []byte {
	qdCount := len(p.Questions)

	buf := new(bytes.Buffer)
	buf.Write([]byte{
		byte(p.ID >> 8),
		byte(p.ID),
		byte(p.Flags >> 8),
		byte(p.Flags),
		byte(qdCount >> 8),
		byte(qdCount),
		0, 0, 0, 0, 0, 0,
	})

	for _, q := range p.Questions {
		buf.Write(q.QNAME)
		buf.Write([]byte{
			byte(q.QTYPE >> 8),
			byte(q.QTYPE),
			byte(q.QCLASS >> 8),
			byte(q.QCLASS),
		})
	}

	return buf.Bytes()
}

// encodeHostname encodes the given hostname into length-prefixed labels.
func encodeHostname(hostname string) ([]byte, error) {
	parts := strings.Split(hostname, ".")

	var retval []byte
	for _, part := range parts {
		if len(part) >= (1 << 8) {
			return nil, fmt.Errorf("error invalid host name: part %s exceeded maximum length", part)
		}
		retval = append(retval, byte(len(part)))
		retval = append(retval, []byte(part)...)
	}
	retval = append(retval, 0)

	return retval, nil
}
