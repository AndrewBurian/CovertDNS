/* ----------------------------------------------------------------------------
SOURCE FILE

Name:		server.go

Program:	CovertDNS

Developer:	Andrew Burian

Created On: 	2015-09-24

Functions:
	func server(dst io.Writer, secret string)
	func decodeAndSave(dst io.Writer, secret string)
	func serveResponse(w dns.ResponseWriter, r *dns.Msg)

Description:
	The main entry point for the CovertDNS program. Contains the main
	function.

Revisions:
	(none)

---------------------------------------------------------------------------- */
package main

import (
	"github.com/miekg/dns"
	"crypto/rand"
	"io"
	"strings"
	"encoding/base32"
	"crypto/rc4"
)

var messages chan chan byte

/* ----------------------------------------------------------------------------
FUNCTION

Name:		Server

Prototype:	func server(dst io.Writer, secret string)

Developer:	Andrew Burian

Created On:	2015-09-24

Parameters:
	dst io.Writer
		the writer to write received data to
	secret string
		the rc4 encryption key for the encrypted data

Return Values:
	(none)

Description:
	Runs the DNS server capable of receiving data encrypted
	with the provided secret, and writes received data to
	the provided writer.

Revisions:
	(none)
---------------------------------------------------------------------------- */
func server(dst io.Writer, secret string) {

	messages = make(chan chan byte, 1)

	// set the server to handle the response for anything
	dns.HandleFunc(".", serveResponse)

	// start the decoder
	go decodeAndSave(dst, secret)

	// run the dns server
	server := &dns.Server{Addr: ":53", Net: "udp"}
	err := server.ListenAndServe()

	if err != nil {
		panic(err)
	}

}

/* ----------------------------------------------------------------------------
FUNCTION

Name:		Decode and Save

Prototype:	func decodeAndSave(dst io.Writer, secret string)

Developer:	Andrew Burian

Created On:	2015-09-24

Parameters:
	dst io.Writer
		the writer to write received data to
	secret string
		the rc4 encryption key for the encrypted data

Return Values:
	(none)

Description:
	Endlessly receives a new channel from the global
	channel store and from that channel, reads in, decrypts
	and saves data to the destination writer

Revisions:
	(none)
---------------------------------------------------------------------------- */
func decodeAndSave(dst io.Writer, secret string) {

	// set up cipher
	cipher, err := rc4.NewCipher([]byte(secret))
	if err != nil {
		panic(err)
	}


	for {
		// get a new data channel
		datachan := <-messages
		data := make([]byte, 0)

		// read until the end of the channel
		for b := range datachan {
			data = append(data, b)
		}

		// add padding back on
		for len(data) % 8 != 0 {
			data = append(data, byte('='))
		}

		// decode
		text := make([]byte, base32.StdEncoding.DecodedLen(len(data)))
		n, err := base32.StdEncoding.Decode(text, data)
		if err != nil {
			panic(err)
		}

		// decrypt
		cipher.XORKeyStream(text[:n], text[:n])

		// save
		dst.Write(text)

	}

}

/* ----------------------------------------------------------------------------
FUNCTION

Name:		Serve Response

Prototype:	func serveResponse(w dns.ResponseWriter, r *dns.Msg)

Developer:	Andrew Burian

Created On:	2015-09-24

Parameters:
	w dns.ResponseWriter
		the writer that will respond to the dns query
	r dns.Msg
		the dns query to respond to

Return Values:
	(none)

Description:
	Handles a single dns request by generating and writing
	a response to the provided writer.
	Also strips out the contained data and sends it over a
	channel

Revisions:
	(none)
---------------------------------------------------------------------------- */
func serveResponse(w dns.ResponseWriter, r *dns.Msg) {

	// create a message to respond with
	respMsg := new(dns.Msg)

	// set this message to be a response to the first
	respMsg.SetReply(r)

	// create a new record to respond with
	respRec := new(dns.A)
	respRec.Hdr = dns.RR_Header {
		Name: r.Question[0].Name,
		Rrtype: dns.TypeA,
		Class: dns.ClassINET,
		Ttl: 0} // 0 TTL to avoid caching

	// pseudo random IP
	rand_bytes := make([]byte, 4)
	rand.Read(rand_bytes)
	respRec.A = rand_bytes

	respMsg.Answer = append(respMsg.Answer, respRec)

	w.WriteMsg(respMsg)

	// now that the response has been handled
	// process the covert data

	// get the data from the query
	data := (strings.Split(r.Question[0].Name, "."))[0]

	// create the new message channel
	datachan := make(chan byte, len(data))

	// send it to the main
	messages <- datachan

	// read all the bytes into the channel
	for _, b := range []byte(data) {
		datachan <- b
	}

	close(datachan)

}
