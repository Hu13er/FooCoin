package main

import (
	"encoding/json"
	"math/rand"
	"sync"
)

type Miner struct {
	*Consumer

	stack  []Transaction
	mutex  sync.Mutex
	cancel chan struct{}
}

func (m *Miner) Start() error {
	if err := m.Consumer.Start(); err != nil {
		return err
	}
	go func() {
		for {
			m.mine()
		}
	}()
	return nil
}

func (m *Miner) dataArrived(from string, data []byte) {
	var blkTx BlockOrTransaction
	if err := json.Unmarshal(data, &blkTx); err != nil {
		return
	}
	switch blkTx.Type {
	case "block":
		if blkTx.Block == nil {
			return
		}
		m.blockArrived(*blkTx.Block)
	case "transaction":
		if blkTx.Transaction == nil {
			return
		}
		m.txArrived(*blkTx.Transaction)
	}
}

func (m *Miner) txArrived(tx Transaction) {
	herpk := m.Parties[tx.From].PublicKey
	if !tx.Verify(herpk) {
		return
	}
	m.mutex.Lock()
	m.stack = append(m.stack, tx)
	m.mutex.Unlock()
}

func (m *Miner) blockArrived(blk Block) {
	select {
	case m.cancel <- struct{}{}:
	default:
	}
}

func (m *Miner) mine() {
	prev := m.blockChain.Longest()
	blk := Block{
		Creator:      m.Name,
		PrevHash:     prev,
		Transactions: m.stack,
		ProveOfWork:  uint(rand.Uint32()),
	}
	for blk.Verify() {
		select {
		case <-m.cancel:
			return
		default:
		}
		blk.ProveOfWork++
	}
	m.SendAll(BlockOrTransaction{
		Type:  "block",
		Block: &blk,
	})
	m.mutex.Lock()
	m.stack = make([]Transaction, 0)
	m.mutex.Unlock()
}
