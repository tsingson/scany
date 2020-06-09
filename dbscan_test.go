package dbscan_test

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"testing"

	"github.com/georgysavva/dbscan/internal/testutil"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/georgysavva/dbscan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB *sql.DB
	ctx    = context.Background()
)

type testDst struct {
	Foo string
	Bar string
}

func TestQueryAll(t *testing.T) {
	t.Parallel()
	sqlText := `
		SELECT *
		FROM (
			VALUES ('foo val', 'bar val'), ('foo val 2', 'bar val 2'), ('foo val 3', 'bar val 3')
		) AS t (foo, bar)
	`
	expected := []*testDst{
		{Foo: "foo val", Bar: "bar val"},
		{Foo: "foo val 2", Bar: "bar val 2"},
		{Foo: "foo val 3", Bar: "bar val 3"},
	}

	var got []*testDst
	err := dbscan.QueryAll(ctx, testDB, &got, sqlText)
	require.NoError(t, err)

	assert.Equal(t, expected, got)
}

func TestQueryOne(t *testing.T) {
	t.Parallel()
	sqlText := `
		SELECT 'foo val' AS foo, 'bar val' AS bar
	`
	expected := testDst{Foo: "foo val", Bar: "bar val"}

	var got testDst
	err := dbscan.QueryOne(ctx, testDB, &got, sqlText)
	require.NoError(t, err)

	assert.Equal(t, expected, got)
}

func TestScanAll(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		rows     testRows
		expected interface{}
	}{
		{
			name: "slice of structs",
			rows: testRows{
				columns: []string{"foo", "bar"},
				data: [][]interface{}{
					{"foo val", "bar val"},
					{"foo val 2", "bar val 2"},
					{"foo val 3", "bar val 3"},
				},
			},
			expected: []struct {
				Foo string
				Bar string
			}{
				{Foo: "foo val", Bar: "bar val"},
				{Foo: "foo val 2", Bar: "bar val 2"},
				{Foo: "foo val 3", Bar: "bar val 3"},
			},
		},
		{
			name: "slice of structs by ptr",
			rows: testRows{
				columns: []string{"foo", "bar"},
				data: [][]interface{}{
					{"foo val", "bar val"},
					{"foo val 2", "bar val 2"},
					{"foo val 3", "bar val 3"},
				},
			},
			expected: []*struct {
				Foo string
				Bar string
			}{
				{Foo: "foo val", Bar: "bar val"},
				{Foo: "foo val 2", Bar: "bar val 2"},
				{Foo: "foo val 3", Bar: "bar val 3"},
			},
		},
		{
			name: "slice of maps",
			rows: testRows{
				columns: []string{"foo", "bar"},
				data: [][]interface{}{
					{"foo val", "bar val"},
					{"foo val 2", "bar val 2"},
					{"foo val 3", "bar val 3"},
				},
			},
			expected: []map[string]interface{}{
				{"foo": "foo val", "bar": "bar val"},
				{"foo": "foo val 2", "bar": "bar val 2"},
				{"foo": "foo val 3", "bar": "bar val 3"},
			},
		},
		{
			name: "slice of strings",
			rows: testRows{
				columns: []string{"foo"},
				data: [][]interface{}{
					{"foo val"},
					{"foo val 2"},
					{"foo val 3"},
				},
			},
			expected: []string{"foo val", "foo val 2", "foo val 3"},
		},
		{
			name: "slice of strings by ptr",
			rows: testRows{
				columns: []string{"foo"},
				data: [][]interface{}{
					{makeStrPtr("foo val")},
					{nil},
					{makeStrPtr("foo val 3")},
				},
			},
			expected: []*string{makeStrPtr("foo val"), nil, makeStrPtr("foo val 3")},
		},
		{
			name: "slice of maps by ptr treated as primitive type case",
			rows: testRows{
				columns: []string{"json"},
				data: [][]interface{}{
					{&map[string]interface{}{"key": "key val"}},
					{nil},
					{&map[string]interface{}{"key": "key val 3"}},
				},
			},
			expected: []*map[string]interface{}{
				{"key": "key val"},
				nil,
				{"key": "key val 3"},
			},
		},
		{
			name: "slice of slices",
			rows: testRows{
				columns: []string{"foo"},
				data: [][]interface{}{
					{[]string{"foo val", "foo val 2"}},
					{[]string{"foo val 3", "foo val 4"}},
					{[]string{"foo val 5", "foo val 6"}},
				},
			},
			expected: [][]string{
				{"foo val", "foo val 2"},
				{"foo val 3", "foo val 4"},
				{"foo val 5", "foo val 6"},
			},
		},
		{
			name: "slice of slices by ptr",
			rows: testRows{
				columns: []string{"foo"},
				data: [][]interface{}{
					{&[]string{"foo val", "foo val 2"}},
					{nil},
					{&[]string{"foo val 5", "foo val 6"}},
				},
			},
			expected: []*[]string{
				{"foo val", "foo val 2"},
				nil,
				{"foo val 5", "foo val 6"},
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dstVal := newDstValue(tc.expected)
			err := dbscan.ScanAll(dstVal.Addr().Interface(), &tc.rows)
			require.NoError(t, err)
			assertDstValueEqual(t, tc.expected, dstVal)
		})
	}
}

func TestScanAll_NonEmptySlice_ResetsDstSlice(t *testing.T) {
	t.Parallel()
	fr := &testRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
			{"foo val 2"},
			{"foo val 3"},
		},
	}
	expected := []string{"foo val", "foo val 2", "foo val 3"}

	got := []string{"junk data", "junk data 2"}
	err := dbscan.ScanAll(&got, fr)
	require.NoError(t, err)

	assert.Equal(t, expected, got)
}

func TestScanAll_NonSliceDestination_ReturnsErr(t *testing.T) {
	t.Parallel()
	rows := &testRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
			{"foo val 2"},
			{"foo val 3"},
		},
	}
	var dst struct {
		Foo string
	}
	expectedErr := "sqlscan: destination must be a slice, got: struct { Foo string }"

	err := dbscan.ScanAll(&dst, rows)

	assert.EqualError(t, err, expectedErr)
}

func TestScanAll_SliceByPointerToPointerDestination_ReturnsErr(t *testing.T) {
	t.Parallel()
	rows := &testRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
			{"foo val 2"},
			{"foo val 3"},
		},
	}
	var dst *[]string
	expectedErr := "sqlscan: destination must be a slice, got: *[]string"

	err := dbscan.ScanAll(&dst, rows)

	assert.EqualError(t, err, expectedErr)
}

func TestScanOne(t *testing.T) {
	t.Parallel()
	rows := &testRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
		},
	}
	type dst struct {
		Foo string
	}
	expected := dst{Foo: "foo val"}

	got := dst{}
	err := dbscan.ScanOne(&got, rows)
	require.NoError(t, err)

	assert.Equal(t, expected, got)
}

func TestScanRow(t *testing.T) {
	t.Parallel()
	rows := &testRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
		},
	}
	type dst struct {
		Foo string
	}
	rows.Next()
	expected := dst{Foo: "foo val"}

	var got dst
	err := dbscan.ScanRow(&got, rows)
	require.NoError(t, err)

	assert.Equal(t, expected, got)
}

func TestScanOne_ZeroRows_ReturnsNotFoundErr(t *testing.T) {
	t.Parallel()
	rows := &testRows{
		columns: []string{"foo"},
		data:    [][]interface{}{},
	}

	var dst string
	err := dbscan.ScanOne(&dst, rows)
	got := dbscan.NotFound(err)

	assert.True(t, got)
}

func TestScanOne_MultipleRows_ReturnsErr(t *testing.T) {
	t.Parallel()
	rows := &testRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
			{"foo val 2"},
			{"foo val 3"},
		},
	}
	expectedErr := "sqlscan: expected 1 row, got: 3"

	var dst string
	err := dbscan.ScanOne(&dst, rows)

	assert.EqualError(t, err, expectedErr)
}

func TestMain(m *testing.M) {
	exitCode := func() int {
		flag.Parse()
		ts := testutil.StartCrdbServer()
		defer ts.Stop()
		var err error
		testDB, err = sql.Open("pgx", ts.PGURL().String())
		if err != nil {
			panic(err)
		}
		defer func() {
			_ = testDB.Close()
		}()
		return m.Run()
	}()
	os.Exit(exitCode)
}