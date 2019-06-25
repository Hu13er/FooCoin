package main

import "encoding/json"

type BlockOrTransaction struct {
	Type        string       `json:"type"`
	Block       *Block       `json:"block,omitempty"`
	Transaction *Transaction `json:"transaction,omitempty"`
}

type Consumer struct {
	*Node
	blockChain BlockChain
}

func NewConsumer(cnf Config) *Consumer {
	return &Consumer{
		Node: NewNode(cnf),
	}
}

func (c *Consumer) Start() error {
	err := c.Node.Start()
	if err != nil {
		return err
	}
	c.Node.ReadAny(c.dataArrived)
	return nil
}

func (c *Consumer) dataArrived(from string, data []byte) {
	var blkTx BlockOrTransaction
	if err := json.Unmarshal(data, &blkTx); err != nil {
		return
	}
	if blkTx.Type == "transaction" {
		// Ignore transactions
		return
	}
	if blkTx.Block == nil {
		return
	}
	block := blkTx.Block
	if !block.Verify() {
		return
	}

	c.blockChain.Append(*block)
}

func (c *Consumer) Values() map[string]int {
	longest := c.blockChain.Longest()
	return c.blockChain.Blocks[longest].values
}
