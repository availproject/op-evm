package avail

// Sender provides interface for sending blocks to Avail. It returns a Future
// to query result of block finalisation.
type Sender interface {
	SubmitData(b *Block) /* Future[Result], */ error
}
