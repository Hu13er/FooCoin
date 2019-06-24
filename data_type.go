package main

type DataType interface {
	Type() [8]byte
	Decode([]byte) error
	Encode() []byte
}

type Name string

func (*Name) Type() [8]byte {
	return [8]byte{'N', 'A', 'M', 'E'}
}

func (n *Name) Encode() []byte {
	return []byte(*n)
}

func (n *Name) Decode(bytes []byte) error {
	for i := range bytes {
		bytes[i] = (*n)[i]
	}
	return nil
}
