package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

type Base64 string

type Transaction struct {
	From      string `json:"form"`
	To        string `json:"to"`
	Value     int    `json:"value"`
	Time      int64  `json:"time"`
	Signature Base64 `json:"sign"`
}

func (tx *Transaction) String() string {
	return fmt.Sprintf("%s,%s,%d,%d", tx.From, tx.To, tx.Value, tx.Time)
}

func (tx *Transaction) Verify(publicKey Base64) bool {
	key := BytesToPublicKey(Base64ToBytes(publicKey))
	plain := string(DecryptWithPublicKey(Base64ToBytes(tx.Signature), key))
	return tx.String() == plain
}

func (tx *Transaction) Sign(secretKey Base64) {
	key := BytesToSecretKey(Base64ToBytes(secretKey))
	cipher := EncryptWithSecretKey([]byte(tx.String()), key)
	tx.Signature = BytesToBase64(cipher)
}

type Block struct {
	PrevHash     Base64        `json:"prev_hash"`
	Creator      string        `json:"creator"`
	Transactions []Transaction `json:"transactions"`
	ProveOfWork  uint          `json:"prove_of_work"`
	Hash         Base64        `json:"hash"`
}

func (b *Block) String() string {
	txs := make([]string, len(b.Transactions))
	for i, t := range b.Transactions {
		txs[i] = fmt.Sprintf("(%s)", t.String())
	}
	return fmt.Sprintf("%s,%s,[%s],%d",
		b.PrevHash, b.Creator, strings.Join(txs, ","), b.ProveOfWork)
}

func (b *Block) CalcHash() Base64 {
	bytes := sha256.Sum256([]byte(b.String()))
	return BytesToBase64(bytes[:])
}

func (b *Block) Verify() bool {
	bytes := Base64ToBytes(b.Hash)
	for i := 0; i < 3; i++ {
		if bytes[i] != 0 {
			return false
		}
	}
	return b.CalcHash() == b.Hash
}

type blockMeta struct {
	*Block
	values map[string]int
}

type BlockChain struct {
	Blocks map[Base64]blockMeta
}

func (bc *BlockChain) Longest() Base64 {
	maxhash, max := Base64(""), -1
	for hash, block := range bc.Blocks {
		cnt := 0
		for prev := block.Hash; prev != ""; prev = bc.Blocks[prev].PrevHash {
			cnt++
		}
		if cnt > max {
			maxhash, max = hash, cnt
		}
	}
	return maxhash
}

func (bc *BlockChain) Append(block Block) {
	values := make(map[string]int)
	for k, v := range bc.Blocks[block.PrevHash].values {
		values[k] = v
	}
	for _, tx := range block.Transactions {
		if _, exists := values[tx.From]; !exists {
			values[tx.From] = 0
		}
		values[tx.From] -= tx.Value
		if _, exists := values[tx.To]; !exists {
			values[tx.To] = 0
		}
		values[tx.To] += tx.Value
	}
	bc.Blocks[block.Hash] = blockMeta{
		Block:  &block,
		values: values,
	}
}
