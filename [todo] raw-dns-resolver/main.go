package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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

type rr struct {
	name  []byte
	typ   uint16
	clazz uint16
	ttl   uint32
	rData []byte
}

type question struct {
	name  []byte
	typ   uint16
	clazz uint16
}

func (q *question) Len() int {
	return len(q.name) + 4
}

type packet struct {
	id         uint16
	flags      uint16
	questions  []question
	answers    []rr
	authority  []rr
	additional []rr
}

func parsePacket(r io.Reader) (*packet, error) {
	id, err := readUint16(r)
	if err != nil {
		return nil, err
	}
	flags, err := readUint16(r)
	if err != nil {
		return nil, err
	}
	qdCount, err := readUint16(r)
	if err != nil {
		return nil, err
	}
	anCount, err := readUint16(r)
	if err != nil {
		return nil, err
	}
	nsCount, err := readUint16(r)
	if err != nil {
		return nil, err
	}
	arCount, err := readUint16(r)
	if err != nil {
		return nil, err
	}

}

func readUint16(r io.Reader) (uint16, error) {
	var v uint16
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

func parseQuestion(r io.Reader) (question, error) {

}

func readLengthPrefixedLabels(r io.Reader) ([]byte, error) {
}

func (p *packet) valid() error {
	if len(p.questions) > maxUint16 {
		return errors.New("error questions exceeded maximum length")
	}
	if len(p.answers) > maxUint16 {
		return errors.New("error answers exceeded maximum length")
	}
	if len(p.authority) > maxUint16 {
		return errors.New("error authority exceeded maximum length")
	}
	if len(p.additional) > maxUint16 {
		return errors.New("error additional exceeded maximum length")
	}
	return nil
}

func newPacket(id uint16, flags uint16, questions []question) (*packet, error) {
	p := &packet{
		id:        id,
		flags:     flags,
		questions: questions,
	}

	err := p.valid()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func newARecordQuery(hostname string) (*packet, error) {
	qName, err := encodeHostname(hostname)
	if err != nil {
		return nil, err
	}

	packet, err := newPacket(1, 0x0100, []question{
		{
			name:  qName,
			typ:   1,
			clazz: 1,
		},
	})
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func (p *packet) Bytes() []byte {
	qdCount := len(p.questions)
	anCount := len(p.answers)
	nsCount := len(p.authority)
	arCount := len(p.additional)

	buf := new(bytes.Buffer)
	buf.Write([]byte{
		byte(p.id >> 8),
		byte(p.id),
		byte(p.flags >> 8),
		byte(p.flags),
		byte(qdCount >> 8),
		byte(qdCount),
		byte(anCount >> 8),
		byte(anCount),
		byte(nsCount >> 8),
		byte(nsCount),
		byte(arCount >> 8),
		byte(arCount),
	})

	for _, q := range p.questions {
		buf.Write(q.name)
		buf.Write([]byte{
			byte(q.typ >> 8),
			byte(q.typ),
			byte(q.clazz >> 8),
			byte(q.clazz),
		})
	}

	for _, rrs := range [][]rr{p.answers, p.authority, p.additional} {
		for _, rr := range rrs {
			buf.Write(rr.name)
			buf.Write([]byte{
				byte(rr.typ >> 8),
				byte(rr.typ),
				byte(rr.clazz >> 8),
				byte(rr.clazz),
				byte(rr.ttl >> 24),
				byte(rr.ttl >> 16),
				byte(rr.ttl >> 8),
				byte(rr.ttl),
				byte(len(rr.rData) >> 8),
				byte(len(rr.rData)),
			})
			buf.Write(rr.rData)
		}
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
