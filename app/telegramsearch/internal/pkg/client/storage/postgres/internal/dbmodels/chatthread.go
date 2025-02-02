package dbmodels

type SimilaritySearch struct {
	Message string
}

type ChatThread struct {
	ThreadID int64 `db:"thread_id"`
	Body     []byte
	ChatID   string
}
