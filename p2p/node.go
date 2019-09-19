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
	"noise/proto"
	"os"
	"strconv"
	"strings"
	"time"
)

/** DEFINE MESSAGES111 **/
var (
	opcodeChat noise.Opcode
	//_          noise.Message = (*Msg)(nil)
)

/** ENTRY POINT **/
func (myNode *Node) setup() {
	opcodeChat = noise.RegisterMessage(noise.NextAvailableOpcode(), (*MyMsg)(nil))
	myNode.Node.OnPeerConnected(func(node *noise.Node, peer *noise.Peer) error{
		fmt.Println("Peeeeeeeeeeeeeeer connected")
		return nil
	})
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
				log.Info().Msgf("[%s]: %s", protocol.PeerID(peer), msg.(MyMsg).Text)
			}
		}()
		return nil
	})
}

//Запуск ноды
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
		go myNode.TxsToMempool(spamAmount / goroutinesAmount)
	}

	myNode.Commands()
}

//Отправляет огромное кол-во сообщений на все ноды, делает это в нескольких горутинах
func (myNode *Node) TxsToMempool(amount int){
	var msg MyMsg
	var text string
	for i := 0; i < 700; i++{
		text += "q"
	}
	for i := 0; i < amount; i++{
		msg = MyMsg{
			&proto.Msg{Autor: "Vladislav",
				Text: text,
				Date: time.Now().String(),
			},
		}

		myNode.Mempool.AddTx(msg)
		//skademlia.BroadcastAsync(myNode.Node, msgProto)
		//time.Sleep(100 * time.Millisecond)
	}

}

func (myNode *Node) SendAllTxs(){
	start := time.Now()
	totalTxs := myNode.Mempool.GetTxAmount()
	oneOfMsg := myNode.Mempool[1].(MyMsg)
	for i := 0 ; i < totalTxs; i++{
		skademlia.BroadcastAsync(myNode.Node, myNode.Mempool[i])
		myNode.Mempool[i] = nil
	}
	end := time.Now()
	result := end.Sub(start).String()
	log.Info().Msg("Tatal time for " + strconv.Itoa(totalTxs) + " messages: " + result)
	log.Info().Msg("Message size: " + strconv.Itoa(len(oneOfMsg.Autor) + len(oneOfMsg.Text) + len(oneOfMsg.Date)))
}

//Ждет ввода команды, для взаимодействия с с приложением
func (myNode *Node) Commands(){
	reader := bufio.NewReader(os.Stdin)
	for {
		txt, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		//Возвращает списо5к всех нод
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

		//Выводит на экран все элемента мемпула
		if strings.Contains(txt, "mempool"){
			fmt.Println(myNode.Mempool)
		}

		//Возвращает размер мемпула
		if strings.Contains(txt, "memsize"){
			fmt.Println(myNode.Mempool.GetTxAmount())
		}

		//Отправляет все транзакции из мемпула
		if strings.Contains(txt, "send"){
			myNode.SendAllTxs()
		}

		if strings.Contains(txt, "set_key"){
			myNode.Node.Set("isnode", "yes")
		}

		if strings.Contains(txt, "get_key"){
			fmt.Println(myNode.Node.Get("isnode"))
		}

		if strings.Contains(txt, "newpeer"){

		}
	}
}