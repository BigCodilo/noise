package main

import (
	"flag"
	"noise/p2p"
)

func main() {
	hostFlag := flag.String("h", "127.0.0.1", "host to listen for peers on")
	portFlag := flag.Uint("p", 3001, "port to listen for peers on")
	spamAmount := flag.Int("spam_amount", 0, "amount of spamming msgs")
	goroutinesAmount := flag.Int("go_amount", 1, "amount of goroutines")
	nodeName := flag.String("node_name", "node", "name of node")
	flag.Parse()

	node := p2p.NewNode(*hostFlag, *portFlag, *nodeName)
	node.StartNode(*goroutinesAmount, *spamAmount)
}