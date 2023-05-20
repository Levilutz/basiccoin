package db

type TxIn struct {
	OriginTxId     HashT
	OriginTxOutInd int
	PublicKey      []byte
	Signature      []byte
}

func (txi TxIn) Hash() (HashT, error) {
	return HashGenericItems(
		txi.OriginTxId[:],
		txi.OriginTxOutInd,
		txi.PublicKey,
		txi.Signature,
	)
}

func TxInPackHasher(txins []TxIn) []Hasher {
	out := make([]Hasher, len(txins))
	for i := 0; i < len(txins); i++ {
		out[i] = txins[i]
	}
	return out
}

type TxOut struct {
	Value         int
	PublicKeyHash HashT
}

func (txo TxOut) Hash() (HashT, error) {
	return HashGenericItems(txo.Value, txo.PublicKeyHash)
}

type Tx struct {
	MinBlock int
	Inputs   []TxIn
	Outputs  []TxOut
}

func (tx Tx) Hash() (HashT, error) {
	inputsHash, err := NewDHashList(tx.Inputs)
	if err != nil {
		return HashT{}, err
	}
	outputsHash, err := NewDHashList(tx.Outputs)
	if err != nil {
		return HashT{}, err
	}
	return HashGenericItems(tx.MinBlock, inputsHash, outputsHash)
}

func HashPreSig(minBlock int, outputs []TxOut) (HashT, error) {
	outputsHash, err := NewDHashList(outputs)
	if err != nil {
		return HashT{}, err
	}
	return HashGenericItems(minBlock, outputsHash)
}

type Block struct {
	PrevBlockId HashT
	MerkleRoot  HashT
	Difficulty  HashT
	Nonce       HashT
}

func (b Block) Hash() (HashT, error) {
	return HashGenericItems(b.PrevBlockId, b.MerkleRoot, b.Difficulty, b.Nonce)
}
