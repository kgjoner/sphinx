package baserepo

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
)

//go:embed queries/*.sql
var sqlFiles embed.FS

var rawQueries = map[string]string{}
var ErrNoQuery = fmt.Errorf("baserepo: raw query not found")

func init() {
	readAndParse("account.sql")
	readAndParse("application.sql")
	readAndParse("link.sql")
	readAndParse("session.sql")
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
