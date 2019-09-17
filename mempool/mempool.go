package mempool

import "github.com/perlin-network/noise"

//Слайс строк - мемпулл
type MemoryPool []noise.Message

func (mempool *MemoryPool) AddTx(tx noise.Message){
	*mempool = append(*mempool, tx)
}

func (mempool *MemoryPool) GetTxAmount() int{
	return len(*mempool)
}