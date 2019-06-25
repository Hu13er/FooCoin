package main

import (
	"fmt"
	"log"
)

func main() {
	pk1, sk1 := GenerateKeyPair(2048)
	c1 := NewConsumer(Config{
		Addr:      ":5000",
		Name:      "C1",
		PublicKey: BytesToBase64(PublicKeyToBytes(pk1)),
		SecretKey: BytesToBase64(SecretKeyToBytes(sk1)),
	})
	log.Println("STARTED")
	log.Println(c1.Start())

	pk2, sk2 := GenerateKeyPair(2048)
	c2 := NewConsumer(Config{
		Addr:      ":5001",
		Name:      "C2",
		PublicKey: BytesToBase64(PublicKeyToBytes(pk2)),
		SecretKey: BytesToBase64(SecretKeyToBytes(sk2)),
	})
	log.Println(c2.Start())
	log.Println("STARTED")
	c2.Connect("localhost:5000")

	pk3, sk3 := GenerateKeyPair(2048)
	m1 := NewMiner(Config{
		Addr:      ":5002",
		Name:      "M1",
		PublicKey: BytesToBase64(PublicKeyToBytes(pk3)),
		SecretKey: BytesToBase64(SecretKeyToBytes(sk3)),
	})
	log.Println(m1.Start())
	log.Println("STARTED")
	m1.Connect("localhost:5000")
	m1.Connect("localhost:5001")

	c1.NewTransaction("C2", 5)
	c1.NewTransaction("C2", 3)

	fmt.Scanln()

	fmt.Println(c1.Values())
	fmt.Println(c2.Values())
}
