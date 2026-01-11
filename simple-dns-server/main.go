package main

import (
	"context"
	"encoding/binary"
	"log"

	"codeberg.org/miekg/dns"
)

const (
	addr = ":8080"
)

func main() {
	mux := dns.NewServeMux()

	mux.HandleFunc(".", func(ctx context.Context, w dns.ResponseWriter, m *dns.Msg) {
		log.Println("DNS query received", m)
		if len(m.Question) == 0 {
			return
		}

		q := m.Question[0]
		qHeader := q.Header()
		log.Printf("DNS query received for %s (not handled)", qHeader.Name)

		// Return NXDOMAIN (domain not found)
		msg := m.Copy()
		msg.Response = true // Mark as response, not query
		msg.Rcode = dns.RcodeNameError

		// Pack and write the response
		if err := msg.Pack(); err != nil {
			log.Printf("error packing DNS response: %v", err)
			return
		}

		// For TCP, DNS messages must be prefixed with a 2-byte length field
		length := make([]byte, 2)
		binary.BigEndian.PutUint16(length, uint16(len(msg.Data)))

		// Write length prefix first, then the message data
		if _, err := w.Write(length); err != nil {
			log.Printf("error writing DNS response length: %v", err)
			return
		}
		if _, err := w.Write(msg.Data); err != nil {
			log.Printf("error writing DNS response: %v", err)
			return
		}
	})

	server := &dns.Server{
		Addr:    addr,
		Net:     "tcp",
		Handler: mux,
	}

	log.Printf("starting DNS server on %s (UDP)", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error starting DNS server: %v", err)
	}
}
