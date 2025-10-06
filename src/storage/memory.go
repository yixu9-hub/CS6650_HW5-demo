package storage

import (
	"errors"
	"sync"

	"hw5/models"
)

// MemoryStore keeps products in-memory using a concurrency-safe map.
type MemoryStore struct {
	mu       sync.RWMutex
	products map[int]models.Product
}

// NewMemoryStore constructs an empty MemoryStore instance.
// key type: int (ProductID), value type: models.Product
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{products: make(map[int]models.Product)}
}

// ErrNotFound indicates that a product with the provided ID was not found.
var ErrNotFound = errors.New("product not found")

// UpsertProduct creates or updates a product by its identifier.
func (s *MemoryStore) UpsertProduct(product models.Product) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.products[product.ProductID] = product
}

// GetProduct fetches a product by ID.
func (s *MemoryStore) GetProduct(id int) (models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, ok := s.products[id]
	if !ok {
		return models.Product{}, ErrNotFound
	}

	return product, nil
}
