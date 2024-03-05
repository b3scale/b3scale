package schema

import (
	"embed"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Embedded filesystem with all migrations files
//
//go:embed migrations/*
var migrationsFs embed.FS

//
// INSERT INTO __meta__ (version, description)
// VALUES (1, 'initial schema');
//

// A Migration consists of an SQL statement and an ID. The Sequence
// is the order of the migrations.
type Migration struct {
	Seq  int
	Name string
	SQL  string
}

// Migrations is a sorted collection of migrations
type Migrations []*Migration

// Len implements the sort interface
func (m Migrations) Len() int {
	return len(m)
}

// Less implements the sort interface and compares the sequence
func (m Migrations) Less(i, j int) bool {
	return m[i].Seq < m[j].Seq
}

// Swap implements the sort interface
func (m Migrations) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// migrationFromFile creates a new migration
func migrationFromFile(name string, sql []byte) *Migration {
	name = strings.TrimSuffix(name, ".sql")
	tokens := strings.Split(name, "_")
	seq, err := strconv.Atoi(tokens[0])
	if err != nil {
		panic(err)
	}
	desc := strings.Join(tokens[1:], " ")
	return &Migration{
		Seq:  seq,
		Name: desc,
		SQL:  string(sql),
	}
}

// GetMigrations retrievs all migrations from the embedded filesystem
func GetMigrations() Migrations {
	dirents, err := migrationsFs.ReadDir("migrations")
	if err != nil {
		panic(err) // This should not happen
	}

	// Get all entries in the migrations dir
	migrations := Migrations{}
	for _, ent := range dirents {
		name := ent.Name()
		sql, err := migrationsFs.ReadFile(filepath.Join("migrations", name))
		if err != nil {
			panic(err)
		}
		m := migrationFromFile(name, sql)
		migrations = append(migrations, m)
	}
	sort.Sort(migrations)
	return migrations
}
