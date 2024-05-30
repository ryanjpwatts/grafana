package sqlstash

import (
	"database/sql"
	"errors"
	"testing"
	"text/template"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/services/store/entity/db"
	"github.com/grafana/grafana/pkg/services/store/entity/db/dbimpl"
	sqltemplateMocks "github.com/grafana/grafana/pkg/services/store/entity/sqlstash/sqltemplate/mocks"
)

// NewMockDBNopSQL returns a db.DB and a sqlmock.Sqlmock that doesn't validates
// SQL. This is only meant to be used to test wrapping utilities exec, query and
// queryRow, where the actual SQL is not relevant to the unit tests, but rather
// how the possible derived error conditions handled.
func NewMockDBNopSQL(t *testing.T) (db.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(
		func(expectedSQL, actualSQL string) error {
			return nil
		},
	)))

	return newUnitTestDB(t, db, mock, err)
}

func newUnitTestDB(t *testing.T, db *sql.DB, mock sqlmock.Sqlmock, err error) (db.DB, sqlmock.Sqlmock) {
	t.Helper()

	require.NoError(t, err)

	return dbimpl.NewDB(db, "sqlmock"), mock
}

func TestCreateETag(t *testing.T) {
	t.Parallel()

	v := createETag(nil, nil, nil)
	require.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", v)
}

func TestGetCurrentUser(t *testing.T) {
	t.Parallel()

	ctx := NewDefaultTestContext(t)
	username, err := getCurrentUser(ctx)
	require.NotEmpty(t, username)
	require.NoError(t, err)

	ctx = ctx.WithoutUser()
	username, err = getCurrentUser(ctx)
	require.Empty(t, username)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUserNotFoundInContext)
}

func TestPtrOr(t *testing.T) {
	t.Parallel()

	p := ptrOr[*int]()
	require.NotNil(t, p)
	require.Zero(t, *p)

	p = ptrOr[*int](nil, nil, nil, nil, nil, nil)
	require.NotNil(t, p)
	require.Zero(t, *p)

	v := 42
	v2 := 5
	p = ptrOr(nil, nil, nil, &v, nil, &v2, nil, nil)
	require.NotNil(t, p)
	require.Equal(t, v, *p)

	p = ptrOr(nil, nil, nil, &v)
	require.NotNil(t, p)
	require.Equal(t, v, *p)
}

func TestSliceOr(t *testing.T) {
	t.Parallel()

	p := sliceOr[[]int]()
	require.NotNil(t, p)
	require.Len(t, p, 0)

	p = sliceOr[[]int](nil, nil, nil, nil)
	require.NotNil(t, p)
	require.Len(t, p, 0)

	p = sliceOr([]int{}, []int{}, []int{}, []int{})
	require.NotNil(t, p)
	require.Len(t, p, 0)

	v := []int{1, 2}
	p = sliceOr([]int{}, nil, []int{}, v, nil, []int{}, []int{10}, nil)
	require.NotNil(t, p)
	require.Equal(t, v, p)

	p = sliceOr([]int{}, nil, []int{}, v)
	require.NotNil(t, p)
	require.Equal(t, v, p)
}

func TestMapOr(t *testing.T) {
	t.Parallel()

	p := mapOr[map[string]int]()
	require.NotNil(t, p)
	require.Len(t, p, 0)

	p = mapOr(nil, map[string]int(nil), nil, map[string]int{}, nil)
	require.NotNil(t, p)
	require.Len(t, p, 0)

	v := map[string]int{"a": 0, "b": 1}
	v2 := map[string]int{"c": 2, "d": 3}

	p = mapOr(nil, map[string]int(nil), v, v2, nil, map[string]int{}, nil)
	require.NotNil(t, p)
	require.Equal(t, v, p)

	p = mapOr(nil, map[string]int(nil), v)
	require.NotNil(t, p)
	require.Equal(t, v, p)
}

func TestCountTrue(t *testing.T) {
	t.Parallel()

	v, count := countTrue(), uint64(0)
	require.Equal(t, count, v)

	v = countTrue(false)
	require.Equal(t, count, v)

	v = countTrue(false, false, false)
	require.Equal(t, count, v)

	v, count = countTrue(true), 1
	require.Equal(t, count, v)

	v = countTrue(false, true, false)
	require.Equal(t, count, v)

	v, count = countTrue(false, true, false, true, true), 3
	require.Equal(t, count, v)
}

var (
	validTestTmpl   = template.Must(template.New("test").Parse("nothing special"))
	invalidTestTmpl = template.New("no definition should fail to exec")
	errTest         = errors.New("because of reasons")
)

// expectRows is a testing helper to keep mocks in sync when adding rows to a
// mocked SQL result. This is a helper to test `query` and `queryRow`.
type expectRows[T any] struct {
	*sqlmock.Rows
	ExpectedResults []T

	req *sqltemplateMocks.WithResults[T]
}

func newReturnsRow[T any](dbmock sqlmock.Sqlmock, req *sqltemplateMocks.WithResults[T]) *expectRows[T] {
	return &expectRows[T]{
		Rows: dbmock.NewRows(nil),
		req:  req,
	}
}

// Add adds a new value that should be returned by the `query` or `queryRow`
// operation.
func (r *expectRows[T]) Add(value T, err error) *expectRows[T] {
	r.req.EXPECT().GetScanDest().Return(nil).Once()
	r.req.EXPECT().Results().Return(value, err).Once()
	r.Rows.AddRow()
	r.ExpectedResults = append(r.ExpectedResults, value)

	return r
}

func TestQuery(t *testing.T) {
	t.Parallel()

	t.Run("happy path - without rows", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewWithResults[int64](t)
		db, dbmock := NewMockDBNopSQL(t)
		rows := newReturnsRow(dbmock, req)

		// setup expectations
		req.EXPECT().GetArgs().Return(nil).Once()
		dbmock.ExpectQuery("").WillReturnRows(rows.Rows)

		// execute and assert
		res, err := query(ctx, db, validTestTmpl, req)
		require.NoError(t, err)
		require.Nil(t, res)
	})

	t.Run("happy path - with rows", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewWithResults[int64](t)
		db, dbmock := NewMockDBNopSQL(t)
		rows := newReturnsRow(dbmock, req)

		// setup expectations
		req.EXPECT().GetArgs().Return(nil).Once()
		rows.Add(1, nil).Add(2, nil)
		dbmock.ExpectQuery("").WillReturnRows(rows.Rows)

		// execute and assert
		res, err := query(ctx, db, validTestTmpl, req)
		require.NoError(t, err)
		require.Equal(t, rows.ExpectedResults, res)
	})

	t.Run("error executing template", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewWithResults[int64](t)
		db, _ := NewMockDBNopSQL(t)

		// execute and assert
		res, err := query(ctx, db, invalidTestTmpl, req)
		require.Nil(t, res)
		require.Error(t, err)
		require.True(t, errors.As(err, new(template.ExecError)))
	})

	t.Run("error executing query", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewWithResults[int64](t)
		db, dbmock := NewMockDBNopSQL(t)

		// setup expectations
		req.EXPECT().GetArgs().Return(nil)
		req.EXPECT().GetScanDest().Return(nil).Maybe()
		dbmock.ExpectQuery("").WillReturnError(errTest)

		// execute and assert
		res, err := query(ctx, db, validTestTmpl, req)
		require.Zero(t, res)
		require.Error(t, err)
		require.ErrorAs(t, err, new(SQLError))
	})

	t.Run("error getting results", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewWithResults[int64](t)
		db, dbmock := NewMockDBNopSQL(t)
		rows := newReturnsRow(dbmock, req)

		// setup expectations
		req.EXPECT().GetArgs().Return(nil).Once()
		rows.Add(1, errTest)
		dbmock.ExpectQuery("").WillReturnRows(rows.Rows)

		// execute and assert
		res, err := query(ctx, db, validTestTmpl, req)
		require.Zero(t, res)
		require.Error(t, err)
		require.ErrorContains(t, err, "row results")
	})
}

func TestQueryRow(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewWithResults[int64](t)
		db, dbmock := NewMockDBNopSQL(t)
		rows := newReturnsRow(dbmock, req)

		// setup expectations
		req.EXPECT().GetArgs().Return(nil).Once()
		rows.Add(1, nil)
		dbmock.ExpectQuery("").WillReturnRows(rows.Rows)

		// execute and assert
		res, err := queryRow(ctx, db, validTestTmpl, req)
		require.NoError(t, err)
		require.Equal(t, rows.ExpectedResults[0], res)
	})

	t.Run("error executing template", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewWithResults[int64](t)
		db, _ := NewMockDBNopSQL(t)

		// execute and assert
		res, err := queryRow(ctx, db, invalidTestTmpl, req)
		require.Zero(t, res)
		require.Error(t, err)
		require.ErrorContains(t, err, "execute template")
	})

	t.Run("error executing query", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewWithResults[int64](t)
		db, dbmock := NewMockDBNopSQL(t)

		// setup expectations
		req.EXPECT().GetArgs().Return(nil)
		req.EXPECT().GetScanDest().Return(nil).Maybe()
		dbmock.ExpectQuery("").WillReturnError(errTest)

		// execute and assert
		res, err := queryRow(ctx, db, validTestTmpl, req)
		require.Zero(t, res)
		require.Error(t, err)
		require.ErrorAs(t, err, new(SQLError))
	})
}

// scannerFunc is an adapter for the `scanner` interface.
type scannerFunc func(dest ...any) error

func (f scannerFunc) Scan(dest ...any) error {
	return f(dest...)
}

func TestScanRow(t *testing.T) {
	t.Parallel()

	const value int64 = 1

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		// test declarations
		req := sqltemplateMocks.NewWithResults[int64](t)
		sc := scannerFunc(func(dest ...any) error {
			return nil
		})

		// setup expectations
		req.EXPECT().GetScanDest().Return(nil).Once()
		req.EXPECT().Results().Return(value, nil).Once()

		// execute and assert
		res, err := scanRow(sc, req)
		require.NoError(t, err)
		require.Equal(t, value, res)
	})

	t.Run("scan error", func(t *testing.T) {
		t.Parallel()

		// test declarations
		req := sqltemplateMocks.NewWithResults[int64](t)
		sc := scannerFunc(func(dest ...any) error {
			return errTest
		})

		// setup expectations
		req.EXPECT().GetScanDest().Return(nil).Once()

		// execute and assert
		res, err := scanRow(sc, req)
		require.Zero(t, res)
		require.Error(t, err)
		require.ErrorIs(t, err, errTest)
	})

	t.Run("results error", func(t *testing.T) {
		t.Parallel()

		// test declarations
		req := sqltemplateMocks.NewWithResults[int64](t)
		sc := scannerFunc(func(dest ...any) error {
			return nil
		})

		// setup expectations
		req.EXPECT().GetScanDest().Return(nil).Once()
		req.EXPECT().Results().Return(0, errTest).Once()

		// execute and assert
		res, err := scanRow(sc, req)
		require.Zero(t, res)
		require.Error(t, err)
		require.ErrorIs(t, err, errTest)
	})
}

func TestExec(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewSQLTemplateIface(t)
		db, dbmock := NewMockDBNopSQL(t)

		// setup expectations
		req.EXPECT().GetArgs().Return(nil).Once()
		dbmock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 0))

		// execute and assert
		res, err := exec(ctx, db, validTestTmpl, req)
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("error executing template", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewSQLTemplateIface(t)
		db, _ := NewMockDBNopSQL(t)

		// execute and assert
		res, err := exec(ctx, db, invalidTestTmpl, req)
		require.Nil(t, res)
		require.Error(t, err)
		require.ErrorContains(t, err, "execute template")
	})

	t.Run("error executing SQL", func(t *testing.T) {
		t.Parallel()

		// test declarations
		ctx := NewDefaultTestContext(t)
		req := sqltemplateMocks.NewSQLTemplateIface(t)
		db, dbmock := NewMockDBNopSQL(t)

		// setup expectations
		req.EXPECT().GetArgs().Return(nil)
		dbmock.ExpectExec("").WillReturnError(errTest)

		// execute and assert
		res, err := exec(ctx, db, validTestTmpl, req)
		require.Nil(t, res)
		require.Error(t, err)
		require.ErrorAs(t, err, new(SQLError))
	})
}
