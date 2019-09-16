package main

import (
	"flag"
	"fmt"
	"noise/proto"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/cipher/aead"
	"github.com/perlin-network/noise/handshake/ecdh"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/payload"
	"github.com/perlin-network/noise/protocol"
	"github.com/perlin-network/noise/skademlia"
	"github.com/pkg/errors"
	"strconv"
	"time"
)



/** DEFINE MESSAGES **/
var (
	opcodeChat noise.Opcode
	_          noise.Message = (*Msg)(nil)
)

type Msg proto.Msg


func (Msg) Read(reader payload.Reader) (noise.Message, error) {
	text, err := reader.ReadString()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read chat msg")
	}

	return Msg{Text: text}, nil
}

func (m Msg) Write() []byte {
	return payload.NewWriter(nil).WriteString(fmt.Sprintf("Autor: %v, Text: %v, Date: %v", m.Autor, m.Text, m.Date)).Bytes()
}

/** ENTRY POINT **/
func setup(node *noise.Node) {
	opcodeChat = noise.RegisterMessage(noise.NextAvailableOpcode(), (*Msg)(nil))

	node.OnPeerInit(func(node *noise.Node, peer *noise.Peer) error {
		peer.OnConnError(func(node *noise.Node, peer *noise.Peer, err error) error {
			log.Info().Msgf("Got an error: %v", err)

			return nil
		})

		peer.OnDisconnect(func(node *noise.Node, peer *noise.Peer) error {
			log.Info().Msgf("Peer %v has disconnected.", peer.RemoteIP().String()+":"+strconv.Itoa(int(peer.RemotePort())))

			return nil
		})

		go func() {
			for {
				msg := <-peer.Receive(opcodeChat)
				log.Info().Msgf("[%s]: %s", protocol.PeerID(peer), msg.(Msg).Text)
			}
		}()

		return nil
	})
}

func main() {
	hostFlag := flag.String("h", "127.0.0.1", "host to listen for peers on")
	portFlag := flag.Uint("p", 3001, "port to listen for peers on")
	spamAmount := flag.Int("spam_amount", 0, "amount of spamming msgs")
	goroutinesAmount := flag.Int("go_amount", 1, "amount of goroutines")

	flag.Parse()

	params := noise.DefaultParams()
	//params.NAT = nat.NewPMP()
	params.Keys = skademlia.RandomKeys()
	params.Host = *hostFlag
	params.Port = uint16(*portFlag)

	node, err := noise.NewNode(params)
	if err != nil {
		panic(err)
	}
	defer node.Kill()

	p := protocol.New()
	p.Register(ecdh.New())
	p.Register(aead.New())
	p.Register(skademlia.New())
	p.Enforce(node)

	setup(node)
	go node.Listen()

	log.Info().Msgf("Listening for peers on port %d.", node.ExternalPort())

	if len(flag.Args()) > 0 {
		for _, address := range flag.Args() {
			peer, err := node.Dial(address)
			if err != nil {
				log.Error().Msg("Cannot connect to " + address)
				continue
			}
			skademlia.WaitUntilAuthenticated(peer)
		}
		peers := skademlia.FindNode(node, protocol.NodeID(node).(skademlia.ID), skademlia.BucketSize(), 8)
		log.Info().Msgf("Bootstrapped with peers: %+v", peers)
	}

	for i := 0; i < *goroutinesAmount; i++{
		go SpamMsgs(node, *spamAmount / *goroutinesAmount)
	}

	for{
		time.Sleep(time.Second * 10)
	}
}

func SpamMsgs(node *noise.Node, amount int){
	start := time.Now()
	var msgProto Msg
	for i := 0; i < amount; i++{
		msgProto = Msg{
			Autor:                "Vladislav",
			Text:                 "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
				"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
				"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris " +
				"nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in " +
				"reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. " +
				"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui " +
				"officia deserunt mollit anim id est laborum",
			Date:                 time.Now().String(),
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		}

		skademlia.BroadcastAsync(node, msgProto)
		//time.Sleep(100 * time.Millisecond)
	}
	end := time.Now()
	result := end.Sub(start).String()
	log.Info().Msg("Tatal time for " + strconv.Itoa(amount) + " messages: " + result)
	msgSize := strconv.Itoa(len(msgProto.Date) + len(msgProto.Autor) + len(msgProto.Text))
	log.Info().Msg("Message size:  " + msgSize)
}