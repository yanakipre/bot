package sqldata

type ActionEnum string

type Key int

var DataKey Key

type Option func(*Data)

type Data struct {
	Operation string
	Action    ActionEnum
	Stmt      string
	Args      []any
}
