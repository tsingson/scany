module github.com/georgysavva/scany

go 1.14

replace github.com/jackc/pgx/v5 v5.0.0 => ../pgx

require (
	github.com/cockroachdb/cockroach-go/v2 v2.2.0
	github.com/jackc/pgtype v1.6.2
	github.com/jackc/pgx/v4 v4.10.1
	github.com/jackc/pgx/v5 v5.0.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)
