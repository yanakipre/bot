package sqltooling

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/yanakipre/bot/internal/encodingtooling"
)

const EmbeddedStructurePrefix = "pg"

// structHasFields handles special case, when we want to execute a query,
// but NOT wanting the query result to be mapped onto a struct.
//
// Examples:
// 1. call Exec, not Query or Select or Get,
// 2. call Select but map the result by ourselves for some reason.
//
// We pass a nil to NewStmt to highlight this.
func structHasFields(s any) bool {
	if s == nil {
		return false
	}
	st := reflect.TypeOf(s)
	if st.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Got unexpected type. Should be a struct. Got %#v", s))
	}
	if st.NumField() == 0 {
		panic("Got struct without fields?")
	}
	return true
}

// ColumnsFromStruct returns list of columns to scan in returned record set.
// If tableName in non-empty it will be prepended to each column name as
// "tableName"."columnName".
func ColumnsFromStruct(s any, tableName string, recusionStarted bool) []string {
	ifv := reflect.ValueOf(s)
	st := reflect.TypeOf(s)
	var columns []string
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		fieldV := ifv.Field(i)
		if fieldV.Kind() == reflect.Struct &&
			strings.HasPrefix(strings.ToLower(field.Name), EmbeddedStructurePrefix) {
			dbTag := field.Tag.Get("db")
			if dbTag == "" {
				panic(fmt.Sprintf(
					"you must provide embedded structure %q.%q with 'db' tag that is equal to database table name",
					st.Name(),
					field.Name,
				))
			}
			if dbTag == "-" {
				continue
			}

			for _, c := range ColumnsFromStruct(fieldV.Interface(), "", true) {
				cName := fmt.Sprintf("\"%s.%s\"", dbTag, c)
				columns = append(columns, cName)
			}
			continue
		}
		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		}
		if len(dbTag) == 0 {
			dbTag = encodingtooling.CamelToSnake(field.Name)
		}
		if dbTag == "" {
			panic(dbTag)
		}
		if recusionStarted {
			columns = append(columns, dbTag)
		} else {
			col := fmt.Sprintf("\"%s\"", dbTag)
			if tableName != "" {
				col = fmt.Sprintf("\"%s\".%s", tableName, col)
			}

			columns = append(columns, col)
		}
	}
	return columns
}

// newQuery wrap query columns.
// It allows you to write SQL requests with using * as list of columns. It produces
// SELECT {columns from GetColumnsFromStruct} FROM ( {your query} } w.
// Therefor if any db table has more columns, than our structure, then the result will not fail.
// You also can write complex joins, by using structure embedding.
//
//	type PGUser struct {
//	  UserName string
//	}
//
//	type PGApiKey struct {
//	  UserID string
//	}
//
//	type ComplexJoinResult struct {
//		 PGApiKey `db:"api_keys"`
//	  PGUser   `db:"users"`
//	}
//
// var query = `SELECT * FROM api_keys JOIN users ON id = user_id`
//
// var r ComplexJoinResult;
// r.toApiKeyModel() // if PGApiKey.toApiKeyModel exists
// r.toUserModel() // if PGUser.toUserModel exists
//
// For more info, read
//   - https://jmoiron.github.io/sqlx/#advancedScanning
//   - https://github.com/jmoiron/sqlx/blob/master/sqlx_test.go#L470
//
// By convention for this to work the models MUST be prefixed
// with EmbeddedStructurePrefix in any case.
func newQuery(name string, query string, st any) string {
	query = strings.TrimSpace(query)
	query = strings.TrimRight(query, ";") // strip the ';' from the right
	result := query
	if !structHasFields(st) {
		// short path, user does not want the results
		return addNameToQuery(name, result)
	} else if strings.Contains(query, returningStart) {
		columns := ColumnsFromStruct(st, "", false)
		result = strings.Replace(query, returningStart, "RETURNING "+strings.Join(columns, ", "), 1)
	} else {
		columns := ColumnsFromStruct(st, "", false)
		query = overwriteForStructureEmbedding(columns, query)
		result = fmt.Sprintf("SELECT %s FROM (%s) w", strings.Join(columns, ", "), query)
	}
	return addNameToQuery(name, result)
}

func NewStmt(name string, query string, st any) *Stmt {
	return &Stmt{
		Name:  name,
		Query: newQuery(name, query, st),
	}
}

type Stmt struct {
	Query string
	Name  string
}

func addNameToQuery(name, query string) string {
	if name == "" {
		panic("name must be specified")
	}
	if strings.Contains(name, "\n") {
		panic("name cannot contain new lines")
	}
	return fmt.Sprintf("-- %s\n%s", name, query)
}

// ColumnsForEmbeddedStructs returns list of columns for squirrel
// sqlx rows.Scan needs, for example, column name "projects.name" to work with structures.
// But squirrel that we use for some of the queries generates '"projects"."name"'.
// So, we need to make query like this:
// SELECT "projects.name" FROM (SELECT projects.name "projects.name" FROM projects)
//
// see newQuery for more details
func ColumnsForEmbeddedStructs(st any) ([]string, []string) {
	outer := ColumnsFromStruct(st, "", false)
	inner := makeColumnsForStructureEmbedding(outer)
	return outer, inner
}

func isColumnForEmbeddedStruct(c string) bool {
	return strings.Contains(c, ".") && strings.HasPrefix(c, "\"")
}

const (
	selectStart    = "SELECT *"
	returningStart = "RETURNING *"
)

// overwriteForStructureEmbedding transforms
//
// SELECT *
//
// into
//
// SELECT tablename.field "tablename.id", ...
//
// to allow structure embedding with SELECT *
func overwriteForStructureEmbedding(columns []string, query string) string {
	embeddedColumnsExist := false
	for _, c := range columns {
		if isColumnForEmbeddedStruct(c) {
			embeddedColumnsExist = true
			break
		}
	}
	if !embeddedColumnsExist {
		return query
	}
	transColumns := makeColumnsForStructureEmbedding(columns)
	return strings.Replace(query, selectStart, "SELECT "+strings.Join(transColumns, ", "), 1)
}

// iterate through all the columns
// and produce SQL compatible slice of columns in a special form,
// turning
//
// > "projects.id"
//
// into
//
// > projects.id "projects.id"
func makeColumnsForStructureEmbedding(columns []string) []string {
	transColumns := make([]string, 0, len(columns))
	for _, c := range columns {
		// keep in mind it skips the non-embedded columns
		// because we expect that if embedded structs are not present,
		// we will not call makeColumnsForStructureEmbedding at all.
		if isColumnForEmbeddedStruct(c) {
			transColumns = append(transColumns, fmt.Sprintf("%s %s", c[1:len(c)-1], c))
		}
	}

	return transColumns
}
