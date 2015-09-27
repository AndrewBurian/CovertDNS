package main

import (
	"github.com/miekg/dns"
	"crypto/rand"
	"io"
	"strings"
	"encoding/base64"
	"crypto/rc4"
	//"fmt"
)

var messages chan chan byte

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

		//fmt.Println("Starting new decoding sequence")

		// read until the end of the channel
		for b := range datachan {
			data = append(data, b)
		}

		// add padding back on
		for len(data) % 4 != 0 {
			data = append(data, byte('='))
		}

		//fmt.Println("Sequence:")
		//fmt.Println(string(data))

		// decode
		text := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
		n, err := base64.StdEncoding.Decode(text, data)
		if err != nil {
			panic(err)
		}

		//fmt.Printf("Decrypting %v bytes\n", n)

		// decrypt
		cipher.XORKeyStream(text[:n], text[:n])

		// save
		dst.Write(text)

		//fmt.Println("Decoding done")

	}

}

func serveResponse(w dns.ResponseWriter, r *dns.Msg) {

	//fmt.Println("Got new query")
	//fmt.Println(r.Question[0].Name)

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

	//fmt.Println("Responded to query")

	// now that the response has been handled
	// process the covert data

	// get the data from the query
	data := (strings.Split(r.Question[0].Name, "."))[0]

	//fmt.Println("Sending encoded message")
	//fmt.Println(data)

	// create the new message channel
	datachan := make(chan byte, len(data))

	// send it to the main
	messages <- datachan

	// read all the bytes into the channel
	for _, b := range []byte(data) {
		datachan <- b
	}

	close(datachan)

	//fmt.Println("Done sended encoded message")

}
