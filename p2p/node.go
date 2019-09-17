package p2p

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/cipher/aead"
	"github.com/perlin-network/noise/handshake/ecdh"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/protocol"
	"github.com/perlin-network/noise/skademlia"
	"os"
	"strconv"
	"strings"
	"time"
)

/** DEFINE MESSAGES111 **/
var (
	opcodeChat noise.Opcode
	_          noise.Message = (*Msg)(nil)
)

/** ENTRY POINT **/
func (myNode *Node) setup() {
	opcodeChat = noise.RegisterMessage(noise.NextAvailableOpcode(), (*Msg)(nil))

	myNode.Node.OnPeerInit(func(node *noise.Node, peer *noise.Peer) error {
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

func (myNode *Node) StartNode(goroutinesAmount, spamAmount int) {

	params := noise.DefaultParams()
	//params.NAT = nat.NewPMP()
	params.Keys = skademlia.RandomKeys()
	params.Host = myNode.IPAddr
	params.Port = uint16(myNode.Port)

	var err error

	myNode.Node, err = noise.NewNode(params)
	if err != nil {
		panic(err)
	}
	defer myNode.Node.Kill()

	p := protocol.New()
	p.Register(ecdh.New())
	p.Register(aead.New())
	p.Register(skademlia.New())
	p.Enforce(myNode.Node)

	myNode.setup()
	go myNode.Node.Listen()

	log.Info().Msgf("Listening for peers on port %d.", myNode.Node.ExternalPort())

	if len(flag.Args()) > 0 {
		for _, address := range flag.Args() {
			peer, err := myNode.Node.Dial(address)
			if err != nil {
				log.Error().Msg("Cannot connect to " + address)
				continue
			}
			skademlia.WaitUntilAuthenticated(peer)
		}
		peers := skademlia.FindNode(myNode.Node, protocol.NodeID(myNode.Node).(skademlia.ID), skademlia.BucketSize(), 8)
		log.Info().Msgf("Bootstrapped with peers: %+v", peers)
	}

	for i := 0; i < goroutinesAmount; i++{
		go myNode.SpamMsgs(spamAmount / goroutinesAmount)
	}

	myNode.Commands()
}

func (myNode *Node) SpamMsgs(amount int){
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
		}

		skademlia.BroadcastAsync(myNode.Node, msgProto)
		//time.Sleep(100 * time.Millisecond)
	}
	end := time.Now()
	result := end.Sub(start).String()
	log.Info().Msg("Tatal time for " + strconv.Itoa(amount) + " messages: " + result)
	msgSize := strconv.Itoa(len(msgProto.Date) + len(msgProto.Autor) + len(msgProto.Text))
	log.Info().Msg("Message size:  " + msgSize)
}

func (myNode Node) Commands(){
	reader := bufio.NewReader(os.Stdin)
	for {
		txt, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		if strings.Contains(txt, "peers"){
			//fmt.Println(myNode.Node.Keys.String())
			nodesInfo := []NodeInfo{}
			allNodes := skademlia.FindNode(myNode.Node, protocol.NodeID(myNode.Node).(skademlia.ID), 128, 128)
			allIPs := skademlia.Table(myNode.Node).GetPeers()
			for i := 0; i < len(allNodes); i++ {
				nodeInfo := NodeInfo{
					PubKey: fmt.Sprintf("%x", allNodes[i].PublicKey()),
					Hash:   fmt.Sprintf("%x", allNodes[i].Hash()),
					IPAddr: allIPs[i],
				}
				nodesInfo = append(nodesInfo, nodeInfo)
			}
			fmt.Println(nodesInfo)
			continue
		}

	}
}