package sqltemplate

import (
	"reflect"
	"testing"
	"text/template"
)

func TestSQLTemplate(t *testing.T) {
	t.Parallel()

	tmpl := New(MySQL)
	tmpl.Arg(1)
	tmpl.Into(reflect.ValueOf(new(int)), "colname")
	tmpl.Reset()
	a := tmpl.GetArgs()
	d := tmpl.GetScanDest()
	if len(a) != 0 || len(d) != 0 {
		t.Fatalf("unexpected values after Reset(). Args: %v, ScanDest: %v", a, d)
	}
}

func TestExecute(t *testing.T) {
	t.Parallel()

	tmpl := template.Must(template.New("test").Parse(`{{ .ID }}`))

	data := struct {
		ID int
	}{
		ID: 1,
	}

	txt, err := Execute(tmpl, data)
	if txt != "1" || err != nil {
		t.Fatalf("unexpected error, txt: %q, err: %v", txt, err)
	}

	txt, err = Execute(tmpl, 1)
	if txt != "" || err == nil {
		t.Fatalf("unexpected result, txt: %q, err: %v", txt, err)
	}
}

func TestFormatSQL(t *testing.T) {
	t.Parallel()

	// TODO: improve testing

	const (
		input = `
			SELECT *
				FROM "mytab" AS t
				WHERE "id">= 3 AND   "str" = ?  ;
		`
		expected = `SELECT *
    FROM "mytab" AS t
    WHERE "id" >= 3 AND "str" = ?;`
	)

	got := FormatSQL(input)
	if expected != got {
		t.Fatalf("Unexpected output.\n\tExpected: %s\n\tActual: %s", expected,
			got)
	}
}
