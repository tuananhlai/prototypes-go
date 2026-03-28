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

	packet, err := newPacketParser(buf[:n]).parse()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", packet)
}

type packetParser struct {
	rawPacket []byte
	cur       int
}

func newPacketParser(b []byte) *packetParser {
	return &packetParser{
		rawPacket: b,
	}
}

func (p *packetParser) parse() (packet, error) {
	id, err := p.readUint16()
	if err != nil {
		return packet{}, err
	}
	flags, err := p.readUint16()
	if err != nil {
		return packet{}, err
	}
	qdCount, err := p.readUint16()
	if err != nil {
		return packet{}, err
	}
	anCount, err := p.readUint16()
	if err != nil {
		return packet{}, err
	}
	nsCount, err := p.readUint16()
	if err != nil {
		return packet{}, err
	}
	arCount, err := p.readUint16()
	if err != nil {
		return packet{}, err
	}

	questions := make([]question, 0, qdCount)
	for range qdCount {
		question, err := p.parseQuestion()
		if err != nil {
			return packet{}, err
		}

		questions = append(questions, question)
	}

	answers := make([]rr, 0, anCount)
	for range anCount {
		answer, err := p.parseRR()
		if err != nil {
			return packet{}, err
		}

		answers = append(answers, answer)
	}

	authority := make([]rr, 0, nsCount)
	for range nsCount {
		record, err := p.parseRR()
		if err != nil {
			return packet{}, err
		}

		authority = append(authority, record)
	}

	additional := make([]rr, 0, arCount)
	for range arCount {
		record, err := p.parseRR()
		if err != nil {
			return packet{}, err
		}

		additional = append(additional, record)
	}

	return packet{
		id:         id,
		flags:      flags,
		questions:  questions,
		answers:    answers,
		authority:  authority,
		additional: additional,
	}, nil
}

func (p *packetParser) parseQuestion() (question, error) {
	name, err := p.readName()
	if err != nil {
		return question{}, err
	}

	// TODO: validate
	typ, err := p.readUint16()
	if err != nil {
		return question{}, err
	}

	// TODO: validate
	clazz, err := p.readUint16()
	if err != nil {
		return question{}, err
	}

	return question{
		name:  name,
		typ:   typ,
		clazz: clazz,
	}, nil
}

func (p *packetParser) parseRR() (rr, error) {
	name, err := p.readName()
	if err != nil {
		return rr{}, fmt.Errorf("error reading name: %v", err)
	}
	typ, err := p.readUint16()
	if err != nil {
		return rr{}, fmt.Errorf("error reading type: %v", err)
	}
	clazz, err := p.readUint16()
	if err != nil {
		return rr{}, fmt.Errorf("error reading class: %v", err)
	}
	ttl, err := p.readUint32()
	if err != nil {
		return rr{}, fmt.Errorf("error reading ttl: %v", err)
	}
	rdLength, err := p.readUint16()
	if err != nil {
		return rr{}, fmt.Errorf("error reading record length: %v", err)
	}
	rData, err := p.readBytes(int(rdLength))
	if err != nil {
		return rr{}, fmt.Errorf("error reading record: %v", err)
	}

	return rr{
		name:  name,
		typ:   typ,
		clazz: clazz,
		ttl:   ttl,
		rData: rData,
	}, nil
}

func (p *packetParser) readUint32() (uint32, error) {
	b, err := p.readBytes(4)
	if err != nil {
		return 0, err
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), nil
}

func (p *packetParser) readUint16() (uint16, error) {
	b, err := p.readBytes(2)
	if err != nil {
		return 0, err
	}
	return uint16(b[0])<<8 | uint16(b[1]), nil
}

func (p *packetParser) readName() ([]byte, error) {
	name, consumed, err := p.readNameAt(p.cur, map[int]bool{})
	if err != nil {
		return nil, err
	}
	p.cur += consumed
	return name, nil
}

func (p *packetParser) readNameAt(pos int, seen map[int]bool) ([]byte, int, error) {
	if pos >= len(p.rawPacket) {
		return nil, 0, fmt.Errorf("name offset out of bounds: %d", pos)
	}
	if seen[pos] {
		return nil, 0, fmt.Errorf("compression loop detected at offset %d", pos)
	}
	seen[pos] = true

	buf := new(bytes.Buffer)
	length := p.rawPacket[pos]

	switch {
	case length == 0:
		_ = buf.WriteByte(0)
		return buf.Bytes(), 1, nil
	case length&0xC0 == 0xC0:
		if pos+1 >= len(p.rawPacket) {
			return nil, 0, fmt.Errorf("unexpected EOF while reading compression pointer")
		}

		offset := int(length&0x3F)<<8 | int(p.rawPacket[pos+1])
		pointedName, _, err := p.readNameAt(offset, seen)
		if err != nil {
			return nil, 0, err
		}
		_, _ = buf.Write(pointedName)
		return buf.Bytes(), 2, nil
	case length&0xC0 != 0:
		return nil, 0, fmt.Errorf("invalid label length byte: 0x%02x", length)
	}

	labelLen := int(length)
	labelStart := pos + 1
	labelEnd := labelStart + labelLen
	if labelEnd > len(p.rawPacket) {
		return nil, 0, fmt.Errorf("unexpected EOF while reading label")
	}

	_ = buf.WriteByte(length)
	_, _ = buf.Write(p.rawPacket[labelStart:labelEnd])
	nextNamePart, consumed, err := p.readNameAt(labelEnd, seen)
	if err != nil {
		return nil, 0, err
	}
	_, _ = buf.Write(nextNamePart)
	return buf.Bytes(), 1 + labelLen + consumed, nil
}

func (p *packetParser) readBytes(n int) ([]byte, error) {
	if p.cur+n > len(p.rawPacket) {
		return nil, fmt.Errorf("unexpected EOF: need %d bytes at offset %d", n, p.cur)
	}

	b := p.rawPacket[p.cur : p.cur+n]
	p.cur += n
	return b, nil
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

type packet struct {
	id         uint16
	flags      uint16
	questions  []question
	answers    []rr
	authority  []rr
	additional []rr
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

// func decodeLengthPrefixedLabel(v []byte) (string, error) {
// 	var parts []string

// 	cur := 0
// 	for {
// 		if cur >= len(v) {
// 			return "", errors.New("")
// 		}

// 		partLen := v[cur]
// 		if partLen == 0 {
// 			break
// 		}

// 		partStart, partEnd := cur+1, cur+1+int(partLen)
// 		if partEnd > len(v) {
// 			return "", fmt.Errorf("error invalid part length at offset %v", cur)
// 		}

// 		parts = append(parts, string(v[partStart:partEnd]))
// 		cur = partEnd
// 	}

// 	return strings.Join(parts, "."), nil
// }
