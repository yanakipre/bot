package sqltooling

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/secret"
)

func TestConnStringKV_kvConnectionString(t *testing.T) {
	type args struct {
		kv ConnStringKV
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "issue #712",
			args: args{
				kv: ConnStringKV{
					Host:     `thost`,
					Database: `injective db\\; drop database yanakipredb; :`,
					User:     `tuser`,
					Password: secret.NewString("tpass"),
					Port:     0,
				},
			},
			expected: `host='thost' database='injective db\\\\; drop database yanakipredb; :' user='tuser' password='tpass' port='0' sslmode='disable' client_encoding='UTF8'`,
		},
		{
			name: "empty values work",
			args: args{
				kv: ConnStringKV{
					Host:     ``,
					Database: "",
					User:     ``,
					Password: secret.NewString(""),
					Port:     0,
				},
			},
			expected: `host='' database='' user='' password='' port='0' sslmode='disable' client_encoding='UTF8'`,
		},
		{
			name: "double quotes",
			args: args{
				kv: ConnStringKV{
					Host:     `"`,
					Database: `"`,
					User:     `"`,
					Password: secret.NewString(`"`),
					Port:     0,
				},
			},
			expected: `host='"' database='"' user='"' password='"' port='0' sslmode='disable' client_encoding='UTF8'`,
		},
		{
			name: "fields get escaped",
			args: args{
				kv: ConnStringKV{
					Host:     `t1'\`,
					Database: "'db'",
					User:     `t2'\`,
					Password: secret.NewString("'pass'"),
					Port:     0,
				},
			},
			expected: `host='t1\'\\' database='\'db\'' user='t2\'\\' password='\'pass\'' port='0' sslmode='disable' client_encoding='UTF8'`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.kv.KVConnectionString()
			require.Equal(t, tt.expected, got.Unmask())
		})
	}
}

func TestConnectionURI_ConnectionURI(t *testing.T) {
	type fields struct {
		Host      string
		Database  string
		User      string
		Password  secret.String
		Port      int
		Paramspec string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "no database",
			fields: fields{
				Host:     "test.host",
				Database: "",
				User:     "user",
				Password: secret.NewString(""),
				Port:     5433,
			},
			want: "postgresql://user@test.host:5433?sslmode=require",
		},
		{
			name: "complex name with escaping and no password",
			fields: fields{
				Host:     "test.host",
				Database: "test",
				User:     "injective role\\; drop database yanakipredb;",
				Password: secret.NewString(""),
				Port:     5432,
			},
			want: "postgresql://injective%20role%5C;%20drop%20database%20yanakipredb;@test.host/test?sslmode=require",
		},
		{
			name: "simple case",
			fields: fields{
				Host:     "fancy-pine-693566.cloud.yanakipre.tech",
				Database: "testdatabase",
				User:     "alice",
				Password: secret.NewString("password"),
				Port:     5432,
			},
			want: "postgresql://alice:password@fancy-pine-693566.cloud.yanakipre.tech/testdatabase?sslmode=require",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConnectionURI{
				Host:     tt.fields.Host,
				Database: tt.fields.Database,
				User:     tt.fields.User,
				Password: tt.fields.Password,
				Port:     tt.fields.Port,
			}
			got := c.ConnectionURI().Unmask()
			require.Equal(t, tt.want, got)
		})
	}
}
