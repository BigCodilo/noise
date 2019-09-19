package p2p

import (
	"fmt"
	//"github.com/BigCodilo/noise/proto"
	pb "github.com/golang/protobuf/proto"

	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/payload"
	"github.com/pkg/errors"
	"noise/mempool"
	. "noise/proto"
)

//Тип сообщения с протобафа
//type Msg proto.Msg
type myString string
type MyMsg struct{
	*Msg
}
//Первая часть реализации ипнтерфейса для отправки сообщения внутри нойцза, вызывается когда приходит какое-то сообщение
//Первая часть реализации ипнтерфейса для отправки сообщения внутри нойцза, вызывается когда приходит какое-то сообщение
func (m MyMsg) Read(reader payload.Reader) (noise.Message, error) {
	text, err := reader.ReadBytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read chat msg")
	}
	myMsg := MyMsg{Msg:new(Msg)}
	err = pb.Unmarshal(text, myMsg.Msg)
	if err != nil {
		fmt.Println(err)
	}
	return myMsg, nil
}

//Вторая часмть реализации интерфейса для отправки сообщений внутри нойза
func (m MyMsg) Write() []byte {
	msgProto, _ := pb.Marshal(m)
	fmt.Println(len(msgProto))
	return payload.NewWriter(nil).WriteString(string(msgProto)).Bytes()
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

