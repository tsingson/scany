// Package pgxscan allows scanning data from pgx.Rows into complex Go types.
/*
pgxscan is a wrapper around github.com/georgysavva/dbscan package.
It contains adapters and proxy functions that are meant to connect github.com/jackc/pgx/v4
with github.com/georgysavva/dbscan functionality. pgxscan mirrors all capabilities provided by dbscan.
See dbscan docs to get familiar with all details and features.

How to use

The most common way to use pgxscan is by calling QueryAll or QueryOne function,
it's as simple as this:

	type User struct {
		ID    string `db:"user_id"`
		Name  string
		Email string
		Age   int
	}

	db, _ := pgxpool.Connect(ctx, "example-connection-url")

	// Use QueryAll to query multiple records.
	var users []*User
	if err := pgxscan.QueryAll(
		ctx, &users, db, `SELECT user_id, name, email, age from users`,
	); err != nil {
		// Handle query or rows processing error.
	}
	// users variable now contains data from all rows.

	// Use QueryOne to query exactly one record.
	var user User
	if err := pgxscan.QueryOne(
		ctx, &user, db, `SELECT user_id, name, email, age from users where id='bob'`,
	); err != nil {
		// Handle query or rows processing error.
	}
	// users variable now contains data from all rows.

pgx custom types

pgx has concept of custom types: https://pkg.go.dev/github.com/jackc/pgx/v4@v4.6.0?tab=doc#hdr-Custom_Type_Support,
You can use them with pgxscan too, here is an example of a struct with pgtype.Text field:

	type User struct {
		UserID string
		Name   string
		Bio    pgtype.Text
	}

Note that you must use pgtype.Text by value, not by pointer. This will not work:

	type User struct {
		UserID string
		Name   string
		Bio    *pgtype.Text // pgxscan won't be able to scan data into field defined that way.
	}

This happens because struct fields are always passed to pgx.Rows.Scan() as pointers,
and if the field type is *pgtype.Text, pgx.Rows.Scan() will receive **pgtype.Text and
pgx won't be able to handle that type, since only *pgtype.Text implements pgx custom type interfaces.
*/
package pgxscan