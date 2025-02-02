package sqltooling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOverwriteForStructureEmbedding(t *testing.T) {
	t.Parallel()
	type args struct {
		columns []string
		query   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nested SELECT * with embedded",
			args: args{
				columns: []string{"\"othertable.field\""},
				query:   "SELECT * FROM (SELECT * FROM tablename)",
			},
			want: "SELECT othertable.field \"othertable.field\" FROM (SELECT * FROM tablename)",
		},
		{
			name: "SELECT * with embedded",
			args: args{
				columns: []string{"\"othertable.field\""},
				query:   "SELECT * FROM tablename",
			},
			want: "SELECT othertable.field \"othertable.field\" FROM tablename",
		},
		{
			name: "SELECT field with embedded",
			args: args{
				columns: []string{"fieldname", "\"othertable.field\""},
				query:   "SELECT fieldname FROM tablename",
			},
			want: "SELECT fieldname FROM tablename",
		},
		{
			name: "SELECT field and not embedded",
			args: args{
				columns: []string{"fieldname"},
				query:   "SELECT fieldname FROM tablename",
			},
			want: "SELECT fieldname FROM tablename",
		},
		{
			name: "SELECT * and not embedded",
			args: args{
				columns: []string{"fieldname"},
				query:   "SELECT * FROM tablename",
			},
			want: "SELECT * FROM tablename",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := overwriteForStructureEmbedding(tt.args.columns, tt.args.query)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewQuery(t *testing.T) {
	type plainStruct struct {
		FieldName string
	}
	type args struct {
		query string
		st    any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "can start with WITH and still wrapped",
			args: args{
				query: "WITH (SELECT 1) SELECT tablename.* FROM tablename",
				st:    plainStruct{},
			},
			want: "-- query name\nSELECT \"field_name\" FROM (WITH (SELECT 1) SELECT tablename.* FROM tablename) w",
		},
		{
			name: "selecting from single table is wrapped",
			args: args{
				query: "SELECT tablename.* FROM tablename",
				st:    plainStruct{},
			},
			want: "-- query name\nSELECT \"field_name\" FROM (SELECT tablename.* FROM tablename) w",
		},
		{
			name: "trim ;",
			args: args{
				query: "SELECT * FROM tablename;",
				st:    plainStruct{},
			},
			want: "-- query name\nSELECT \"field_name\" FROM (SELECT * FROM tablename) w",
		},
		{
			name: "trim whitespaces",
			args: args{
				query: "\n\r\t SELECT * FROM tablename\n\r\t ",
				st:    plainStruct{},
			},
			want: "-- query name\nSELECT \"field_name\" FROM (SELECT * FROM tablename) w",
		},
		{
			name: "insert returning",
			args: args{
				query: "INSERT INTO testtable (id) VALUES (1) RETURNING *",
				st:    plainStruct{},
			},
			want: "-- query name\nINSERT INTO testtable (id) VALUES (1) RETURNING \"field_name\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newQuery("query name", tt.args.query, tt.args.st)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFieldNamesFromStruct(t *testing.T) {
	t.Run("SkipHyphen", func(t *testing.T) {
		type check struct {
			Foo string `db:"-"`
			Bar string `db:"bar"`
		}

		t.Run("NoTableName", func(t *testing.T) {
			col := ColumnsFromStruct(check{}, "", false)

			if len(col) != 1 {
				t.Fatalf("incorrect columns %v", col)
			}
			if col[0] != `"bar"` {
				t.Fatalf("wrong column: %q", col[0])
			}
		})
		t.Run("WithTable", func(t *testing.T) {
			col := ColumnsFromStruct(check{}, "table", false)

			if len(col) != 1 {
				t.Fatalf("incorrect columns %v", col)
			}
			if col[0] != `"table"."bar"` {
				t.Fatalf("wrong column: %q", col[0])
			}
		})
	})
}

func Test_addNameToQuery(t *testing.T) {
	type args struct {
		name  string
		query string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "happy path",
			args: args{
				name: "query name",
				query: `SELECT *
FROM operations`,
			},
			// language=PostgreSQL
			want: `-- query name
SELECT *
FROM operations`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, addNameToQuery(tt.args.name, tt.args.query))
		})
	}
}

func Test_makeColumnsForStructureEmbedding(t *testing.T) {
	type args struct {
		columns []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			// it skips the non-embedded columns
			// because we expect that if embedded structs are not present,
			// we will not call makeColumnsForStructureEmbedding at all.
			name: "mixed embedded and non embedded columns",
			args: args{columns: []string{"my_custom_column", `"projects.id"`}},
			want: []string{`projects.id "projects.id"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeColumnsForStructureEmbedding(tt.args.columns)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_overwriteForStructureEmbedding(t *testing.T) {
	type args struct {
		columns []string
		query   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: `mixed plain columns and embedded columns`,
			args: args{
				columns: []string{"my_custom_column", `"projects.id"`},
				query:   "SELECT *, my_custom_column FROM tablename",
			},
			want: `SELECT projects.id "projects.id", my_custom_column FROM tablename`,
		},
		{
			name: `plain columns should come last (after *)`,
			args: args{
				columns: []string{"my_custom_column", `"projects.id"`},
				query:   "SELECT my_custom_column, * FROM tablename",
			},
			// as you can see, emdedded columns will not be put into query,
			// any you will not get the auto-unwrapping of SELECT *
			//
			// see "mixed plain columns and embedded columns" test case to understand
			// how to build the correct query.
			want: `SELECT my_custom_column, * FROM tablename`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := overwriteForStructureEmbedding(tt.args.columns, tt.args.query)
			require.Equal(t, tt.want, got)
		})
	}
}
