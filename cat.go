package main

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
)

type Cat struct {
	ID          int64
	Name        string
	Gender      string
	BirthDate   dbtype.Date
	Status      string // baby - available - pregnant - died
	LastMatedAt dbtype.Date
	LastMatedId int64
}
