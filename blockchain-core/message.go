package core

type StatusMessage struct { // Message used to send the status of a server / node
	ID               string // ID of the server / node
	BlockchainHeight uint32 // Height of the blockchain
}

type GetStatusMessage struct { // Empty message used to request the status of a server / node
}

type GetBlocksMessage struct {
	FromIndex uint32
	ToIndex   uint32 // if 0, then get all blocks from FromIndex to the latest
}

type BlocksMessage struct { // message that contains a list of encoded blocks requested
	Blocks [][]byte
}
