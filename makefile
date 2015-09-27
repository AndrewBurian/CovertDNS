all: baddns

baddns: $(wildcard *.go)
	go build

