package identrepo

import (
	"bufio"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/kgjoner/cornucopia/v3/prim"
	"github.com/kgjoner/sphinx/internal/domains/identity"
	"github.com/kgjoner/sphinx/internal/shared"
	_ "github.com/lib/pq"
)

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

type DAO struct {
	ctx  context.Context
	dbtx shared.DBTX
}

func (f *Factory) NewDAO(ctx context.Context, dbtx shared.DBTX) identity.Repo {
	return &DAO{
		ctx,
		dbtx,
	}
}

/* =============================================================================
	Listing helpers
============================================================================= */

func handleListQuery[T any](rows *sql.Rows, pag *prim.Pagination, dest func(item *T) []any) (*prim.PaginatedData[T], error) {
	items := []T{}
	for rows.Next() {
		var item T
		err := rows.Scan(dest(&item)...)
		if err == sql.ErrNoRows {
			return nil, nil
		} else if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(items) > pag.Limit {
		pag.HasNext = true
		items = items[:pag.Limit]
	}

	return prim.NewPaginatedData(*pag, items), nil
}

/* =============================================================================
	Raw Queries
============================================================================= */

//go:embed queries/*.sql
var sqlFiles embed.FS

var rawQueries = map[string]string{}
var ErrNoQuery = fmt.Errorf("identrepo: raw query not found")

func init() {
	readAndParse("user.sql")
	readAndParse("external_credential.sql")
}

func readAndParse(filename string) {
	file, err := sqlFiles.Open("queries/" + filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentName string
	var currentQuery string
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "-- name:") {
			content := strings.Split(line, " ")
			currentName = content[2]
		} else if strings.Trim(line, " ") == "" {
			continue
		} else {
			currentQuery += line + "\n"
		}

		if strings.HasSuffix(line, ";") {
			rawQueries[currentName] = currentQuery
			currentName = ""
			currentQuery = ""
		}
	}
}
