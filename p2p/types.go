package p2p

import (
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/payload"
	"github.com/pkg/errors"
	"noise/proto"
)

type Msg proto.Msg

func (m Msg) Read(reader payload.Reader) (noise.Message, error) {
	text, err := reader.ReadString()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read chat msg")
	}

	return Msg{Text: text}, nil
}

func (m Msg) Write() []byte {

	return payload.NewWriter(nil).WriteString(fmt.Sprintf("Autor: %v, Text: %v, Date: %v, NodesVisited: %v", m.Autor, m.Text, m.Date, m.VN)).Bytes()
}

type Node struct{
	Node *noise.Node
	IPAddr string
	Port uint
	NodeName string
}

func NewNode(ip string, port uint, name string) Node{
	return Node{
		Node:   nil,
		IPAddr: ip,
		Port:   port,
		NodeName: name,
	}
}
