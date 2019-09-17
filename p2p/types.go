package p2p

import (
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/payload"
	"github.com/pkg/errors"
	"noise/mempool"
	"noise/proto"
)

//Тип сообщения с протобафа
type Msg proto.Msg

//Первая часть реализации ипнтерфейса для отправки сообщения внутри нойцза, вызывается когда приходит какое-то сообщение
func (Msg) Read(reader payload.Reader) (noise.Message, error) {
	text, err := reader.ReadString()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read chat msg")
	}
	return Msg{Text: text}, nil
}

//Вторая часмть реализации интерфейса для отправки сообщений внутри нойза
func (m Msg) Write() []byte {
	fmt.Println(3)
	return payload.NewWriter(nil).WriteString(fmt.Sprintf("Autor: %v, Text: %v, Date: %v", m.Autor, m.Text, m.Date)).Bytes()
}

//Расширенная структура ноды
type Node struct{
	Node *noise.Node
	IPAddr string
	Port uint
	NodeName string
	Mempool mempool.MemoryPool
}

//Возвращает структуру новой ноды
func NewNode(ip string, port uint, name string) Node{
	return Node{
		Node:   nil,
		IPAddr: ip,
		Port:   port,
		NodeName: name,
	}
}

//Структура хранящая информацию о ноде
type NodeInfo struct{
	IPAddr string
	PubKey string
	Hash string
}

