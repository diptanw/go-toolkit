// Package storage implements data storage types for the persistence layer
package storage

import (
	"errors"
	"sync"
)

var (
	// ErrMissingID is an error when the entity ID is missing.
	ErrMissingID = errors.New("missing record ID")
	// ErrNotFound is an error when record is not found.
	ErrNotFound = errors.New("record is not found")
)

// ID is a type that represents record's unique identifier.
type ID string

// Record is an interface that defines record's requirements.
type Record interface {
	ID() ID
}

// InMemory is a simple in-memory storage.
type InMemory[T Record] struct {
	records sync.Map
}

// NewInMemory return a new instance on InMemory storage for a given type.
func NewInMemory[T Record]() *InMemory[T] {
	return &InMemory[T]{
		records: sync.Map{},
	}
}

// Get returns a single record matching the provided ID.
func (s *InMemory[T]) Get(id ID) (t T, err error) {
	if id == "" {
		return t, ErrMissingID
	}

	r, ok := s.records.Load(id)
	if !ok {
		return t, ErrNotFound
	}

	return r.(T), nil
}

// Update updates an existing record to inserts it if not exist..
func (s *InMemory[T]) Update(r T) error {
	if r.ID() == "" {
		return ErrMissingID
	}

	s.records.Store(r.ID(), r)

	return nil
}

// Remove deletes record from the dataset.
func (s *InMemory[T]) Remove(id ID) error {
	if id == "" {
		return ErrMissingID
	}

	if _, ok := s.records.LoadAndDelete(id); !ok {
		return ErrNotFound
	}

	return nil
}
