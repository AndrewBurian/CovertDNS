package main

import (
	"flag"
	"fmt"
	"os"
	"io"
)

func main() {

	// set up program flags
	help := flag.Bool("help", false, "Display usage")
	send := flag.Bool("send", false, "Send data in client mode")
	recv := flag.Bool("recv", false, "Receive data in server mode")
	filename := flag.String("file", "", "The file to send from or receive to")
	remote := flag.String("remote", "", "Remote address to send to in client mode")
	secret := flag.String("secret", "supersecret", "RC4 encryption key for transmitted data")

	flag.Parse()

	// sanity checks
	if !*send && !*recv {
		*help = true
		fmt.Println("Need to specify either client or server mode")
	}

	if *send && *remote == "" {
		*help = true
		fmt.Println("Must specify remote when in server mode")
	}

	// print usage
	if *help {
		flag.Usage()
		return
	}

	// set the source to either the file or stdin
	var source io.Reader
	var dest io.Writer

	if *filename == "" {
		// if there's no file
		if *send {
			// and we're client, use stdin
			source = os.Stdin
		} else if *recv {
			// and we're server, use stdout
			dest = os.Stdout
		}
	} else {
		file, err := os.Open(*filename)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		if *send {
			source = file
		} else {
			dest = file
		}
	}


	// run client
	if *send {
		fmt.Println("Sending file " + *filename)
		client(source, *remote, *secret)
	}

	// run server
	if *recv {
		fmt.Println("Receiving into file " + *filename)
		server(dest, *secret)
	}
}
