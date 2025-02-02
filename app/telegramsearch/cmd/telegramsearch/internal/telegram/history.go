package telegram

import (
	_ "embed"
)

//go:embed chathistory/history.log
var history []byte

//go:embed chathistory/cylimassol.json
var cylimassol []byte

//go:embed chathistory/cymedicine.json
var cymedicine []byte
