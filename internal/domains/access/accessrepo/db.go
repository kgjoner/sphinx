package accessrepo

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/kgjoner/sphinx/internal/domains/access"
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

func (f *Factory) NewDAO(ctx context.Context, dbtx shared.DBTX) access.Repo {
	return &DAO{
		ctx,
		dbtx,
	}
}

/* =============================================================================
	Raw Queries
============================================================================= */

//go:embed queries/*.sql
var sqlFiles embed.FS

var rawQueries = map[string]string{}
var ErrNoQuery = fmt.Errorf("accessrepo: raw query not found")

func init() {
	readAndParse("application.sql")
	readAndParse("link.sql")
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
