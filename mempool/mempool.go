package mempool

import "github.com/perlin-network/noise"
import "container/list"

//Слайс строк - мемпулл
type MemoryPool []noise.Message

type LinkedMPool list.List

func (mempool *MemoryPool) AddTx(tx noise.Message){
	*mempool = append(*mempool, tx)
}

func (mempool *MemoryPool) GetTxAmount() int{
	return len(*mempool)
}