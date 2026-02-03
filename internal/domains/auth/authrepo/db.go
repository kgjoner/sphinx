package authrepo

import (
	"bufio"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/kgjoner/sphinx/internal/domains/auth"
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

func (f *Factory) NewDAO(ctx context.Context, dbtx shared.DBTX) auth.Repo {
	return &DAO{
		ctx,
		dbtx,
	}
}

/* =============================================================================
	Advisory Lock
============================================================================= */

const (
	SIGNING_KEY_LOCK_ID   int64 = 2026020100
)

// acquireDistributedLock acquires a PostgreSQL advisory lock to prevent concurrent key operations.
// Returns a release function that must be called to release the lock.
// This prevents race conditions when multiple API instances try to generate/rotate keys simultaneously.
func (d DAO) AcquireSigningKeyLock() (release func(), err error) {
	// Try to acquire advisory lock (non-blocking with timeout)
	// pg_try_advisory_lock returns true if lock acquired, false if already held
	var locked bool
	_, isTx := d.dbtx.(*sql.Tx)
	if isTx {
		err = d.dbtx.QueryRowContext(
			d.ctx,
			"SELECT pg_try_advisory_xact_lock($1)",
			SIGNING_KEY_LOCK_ID,
		).Scan(&locked)
	} else {
		err = d.dbtx.QueryRowContext(
			d.ctx,
			"SELECT pg_try_advisory_lock($1)",
			SIGNING_KEY_LOCK_ID,
		).Scan(&locked)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to acquire distributed lock: %w", err)
	}

	if !locked {
		return nil, fmt.Errorf("another instance is currently performing key operations")
	}

	if isTx {
		// In transaction, lock will be released automatically at transaction end
		return func() {}, nil
	}

	// Return release function
	return func() {
		// Release advisory lock
		_, _ = d.dbtx.ExecContext(
			d.ctx,
			"SELECT pg_advisory_unlock($1)",
			SIGNING_KEY_LOCK_ID,
		)
	}, nil
}

/* =============================================================================
	Raw Queries
============================================================================= */

//go:embed queries/*.sql
var sqlFiles embed.FS

var rawQueries = map[string]string{}
var ErrNoQuery = fmt.Errorf("authrepo: raw query not found")

func init() {
	readAndParse("client.sql")
	readAndParse("principal.sql")
	readAndParse("session.sql")
	readAndParse("signing_key.sql")
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
