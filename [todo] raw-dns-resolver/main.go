package main

import (
	"encoding/binary"
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

	query, err := NewARecordQuery("example.com")
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

type Packet struct {
	Header Header
	QNAME  []byte
	QTYPE  uint16
	QCLASS uint16
}

func NewARecordQuery(hostname string) (Packet, error) {
	qName, err := encodeHostname(hostname)
	if err != nil {
		return Packet{}, nil
	}

	return Packet{
		Header: Header{
			ID:    1,
			Flags: 0x0100,
			// Without QDCOUNT, DNS server will not send a response.
			// TODO: find out why.
			QDCOUNT: 1,
		},
		QNAME:  qName,
		QTYPE:  1,
		QCLASS: 1,
	}, nil
}

func (p Packet) Bytes() []byte {
	qTypeStart := 12 + len(p.QNAME)

	packet := make([]byte, 12+len(p.QNAME)+4)
	binary.BigEndian.PutUint16(packet, p.Header.ID)
	binary.BigEndian.PutUint16(packet[2:], p.Header.Flags)
	binary.BigEndian.PutUint16(packet[4:], p.Header.QDCOUNT)
	binary.BigEndian.PutUint16(packet[6:], p.Header.ANCOUNT)
	binary.BigEndian.PutUint16(packet[8:], p.Header.NSCOUNT)
	binary.BigEndian.PutUint16(packet[10:], p.Header.ARCOUNT)

	copy(packet[12:], p.QNAME)

	binary.BigEndian.PutUint16(packet[qTypeStart:], p.QTYPE)
	binary.BigEndian.PutUint16(packet[qTypeStart+2:], p.QCLASS)

	return packet
}

type Header struct {
	ID      uint16
	Flags   uint16
	QDCOUNT uint16
	ANCOUNT uint16
	NSCOUNT uint16
	ARCOUNT uint16
}

type Question struct {
	QNAME  []byte
	QTYPE  uint16
	QCLASS uint16
}

func (q Question) Bytes() []byte {
	return append(q.QNAME,
		byte(q.QTYPE>>8),
		byte(q.QTYPE),
		byte(q.QCLASS>>8),
		byte(q.QCLASS),
	)
}

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
