// Package storage implements data storage types for the persistence layer
package storage

import (
	"crypto/rand"
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrMissingID is an error when ID is missing
	ErrMissingID = errors.New("missing ID")
	// ErrNotFound is an error when record not found
	ErrNotFound = errors.New("record not found")
)

// InMemory is a type that mimics basic database operations.
type InMemory struct {
	records map[ID]interface{}
}

// NewInMemory return a new instance on InMemory storage.
func NewInMemory() *InMemory {
	return &InMemory{
		records: make(map[ID]interface{}),
	}
}

// Matcher is a func that filters dataset result.
type Matcher func(value interface{}) bool

// Find return a result of filtered records.
func (s *InMemory) Find(m Matcher) ([]interface{}, error) {
	var res []interface{}

	for _, v := range s.records {
		if m(v) {
			res = append(res, v)
		}
	}

	return res, nil
}

// Insert adds a new record to the dataset.
func (s *InMemory) Insert(r interface{}) error {
	id := getID(r)
	if id == "" {
		return ErrMissingID
	}

	s.records[id] = r

	return nil
}

// Update updates an existing record.
func (s *InMemory) Update(r interface{}) error {
	id := getID(r)
	if id == "" {
		return ErrMissingID
	}

	if _, ok := s.records[id]; !ok {
		return ErrNotFound
	}

	s.records[id] = r

	return nil
}

// Remove deletes record from dataset.
func (s *InMemory) Remove(id ID) error {
	if id == "" {
		return ErrMissingID
	}

	if _, ok := s.records[id]; !ok {
		return ErrNotFound
	}

	delete(s.records, id)

	return nil
}

func getID(v interface{}) ID {
	val := reflect.ValueOf(v)
	valueField := val.FieldByName("ID")

	return valueField.Interface().(ID)
}

// ID is a type that represents record's unique identifier.
type ID string

// NewID generates a pseudo-random ID sequence.
func NewID() (ID, error) {
	b := make([]byte, 10)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return ID(fmt.Sprintf("%x", b)), nil
}
