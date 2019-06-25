package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
)

type Miner struct {
	*Consumer

	stack  []Transaction
	mutex  sync.Mutex
	cancel chan struct{}
}

func NewMiner(cnf Config) *Miner {
	return &Miner{
		Consumer: NewConsumer(cnf),
		stack:    make([]Transaction, 0),
		cancel:   make(chan struct{}, 1),
	}
}

func (m *Miner) Start() error {
	if err := m.Consumer.Start(); err != nil {
		return err
	}
	m.ReadAny(m.dataArrived)
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
	log.Println(m.Name, ": TX ARRIVED", tx)
	m.mutex.Lock()
	m.stack = append(m.stack, tx)
	log.Println(m.stack)
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
	m.mutex.Lock()
	blk := Block{
		Creator:      m.Name,
		PrevHash:     prev,
		Transactions: m.stack,
		ProveOfWork:  uint(rand.Uint32()),
	}
	m.stack = make([]Transaction, 0)
	m.mutex.Unlock()
	log.Println(m.Name, ": CALCULATING BLOCK", blk)
	for {
		select {
		case <-m.cancel:
			return
		default:
		}
		blk.Hash = blk.CalcHash()
		if blk.Verify() {
			break
		}
		blk.ProveOfWork++
	}
	log.Println(m.Name, ": FOUND BLOCK!!", blk)
	m.blockChain.Append(blk)
	m.SendAll(BlockOrTransaction{
		Type:  "block",
		Block: &blk,
	})
}
