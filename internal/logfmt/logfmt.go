package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/ast/astutil"
)

// command line flags
var (
	paths   string
	verbose bool
)

func init() {
	logfmtCmd.PersistentFlags().
		StringVar(&paths, "paths", ".", "space-separated list of paths to recursively search for *.go files in")
	logfmtCmd.PersistentFlags().
		BoolVarP(&verbose, "verbose", "v", false, "if set to true, prints out all unique zap log field keys found")
}

type WeightedLogKey struct {
	Key    string
	Weight int
}

var logfmtCmd = &cobra.Command{
	Use: "go run logfmt.go",
	Short: `logfmt is a zap.Type(k, v) log field key formatter
		that enforces snake_case key names
		`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := logfmt(); err != nil {
			panic(err)
		}
	},
}

func main() {
	if err := logfmtCmd.Execute(); err != nil {
		panic(err)
	}
}

func logfmt() error {
	var srcfiles []string

	inputPaths := strings.Split(paths, " ")
	searchPaths := make([]string, 0, len(paths))
	for _, p := range inputPaths {
		if path.IsAbs(p) {
			searchPaths = append(searchPaths, p)
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			searchPaths = append(searchPaths, path.Clean(path.Join(wd, p)))
		}
	}

	for _, rootPath := range searchPaths {
		err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(d.Name()) != ".go" {
				return nil
			}
			srcfiles = append(srcfiles, path)
			return nil
		})
		if err != nil {
			return err
		}
	}

	logKeysMap := make(map[string]struct{})

	for _, srcfile := range srcfiles {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, srcfile, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		var shouldRewriteFile bool
		rewrittenFile := astutil.Apply(file, func(c *astutil.Cursor) bool {
			zapFieldKey := getZapFieldKey(c.Node())
			if zapFieldKey == "" {
				return true
			}

			formattedZapFieldKey := toSnakeCase(strings.Trim(zapFieldKey, `"`))

			logKeysMap[formattedZapFieldKey] = struct{}{}

			if formattedZapFieldKey == zapFieldKey {
				return true
			}

			setZapFieldKey(c.Node(), strings.Join([]string{`"`, formattedZapFieldKey, `"`}, ""))
			c.Replace(c.Node())
			shouldRewriteFile = true
			return true
		}, nil)

		if shouldRewriteFile {
			var buf bytes.Buffer
			if err := format.Node(&buf, fset, rewrittenFile); err != nil {
				return err
			}

			if err := os.WriteFile(srcfile, buf.Bytes(), 0o600); err != nil {
				return err
			}
		}
	}

	if verbose {
		logKeys := lo.MapToSlice(logKeysMap, func(k string, _ struct{}) string {
			return strings.Trim(k, `"`)
		})
		sort.Strings(logKeys)
		for _, k := range logKeys {
			fmt.Println(k)
		}
		fmt.Println()
		fmt.Println(len(logKeys), "total unique log keys found in paths:", searchPaths)
	}

	return nil
}

func getZapFieldKey(n ast.Node) string {
	f, ok := n.(*ast.CallExpr)
	if !ok {
		return ""
	}
	mthd, ok := f.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	ident, ok := mthd.X.(*ast.Ident)
	if !ok {
		return ""
	}
	if ident.Name != "zap" {
		return ""
	}
	lit, ok := f.Args[0].(*ast.BasicLit)
	if !ok {
		return ""
	}
	if lit.Kind != token.STRING {
		return ""
	}
	return lit.Value
}

func setZapFieldKey(n ast.Node, key string) {
	f := n.(*ast.CallExpr)
	arg := f.Args[0].(*ast.BasicLit)

	// rewriting ast node anyway, so in-place mutation is ok
	f.Args[0] = &ast.BasicLit{
		ValuePos: arg.ValuePos,
		Kind:     arg.Kind,
		Value:    key,
	}
}

const snakeCaseSeparator = '_'

func toSnakeCase(in string) string {
	var buf bytes.Buffer
	var capitalSequence bool
	for i, r := range in {
		// convert capital letters to lower register
		if isUpper(r) {
			if buf.Len() > 0 && i > 0 &&
				// do not separate sequences of upper case letters
				// like URL, ID, etc.
				isUpper(rune(in[i-1])) {
				capitalSequence = true
			}
			if !capitalSequence {
				buf.WriteRune(snakeCaseSeparator)
			}
			buf.WriteRune(r - 'A' + 'a')
		} else if isLowerAlphaNum(r) {
			if capitalSequence {
				buf.WriteRune(snakeCaseSeparator)
			}
			capitalSequence = false
			buf.WriteRune(r)
		} else {
			if capitalSequence {
				buf.WriteRune(snakeCaseSeparator)
			}
			capitalSequence = false
			// rewrite all non-alphanumeric chars
			buf.WriteRune(snakeCaseSeparator)
		}
	}

	in = buf.String()
	buf.Reset()

	// trim leading and trailing snakeCaseSeparator
	in = strings.Trim(in, "_")

	// normalize and reduced inner separators
	var repeatingSeparator bool
	for _, r := range in {
		if r == snakeCaseSeparator && repeatingSeparator {
			// skip repeating snakeCaseSeparator
			continue
		}
		if r == snakeCaseSeparator {
			// write out individual separator as is
			buf.WriteRune(r)
			repeatingSeparator = true
			continue
		}
		if !isLowerAlphaNum(r) {
			// replace all non-alpha-num chars with separators
			buf.WriteRune(snakeCaseSeparator)
			repeatingSeparator = true
			continue
		}

		// write alpha-numeric-chars as is
		buf.WriteRune(r)

		// unset snakeCaseSeparator separator sequence marker
		repeatingSeparator = false
	}

	return buf.String()
}

func isUpper(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

func isLowerAlphaNum(r rune) bool {
	return ('a' <= r && r <= 'z') || ('0' <= r && r <= '9')
}
