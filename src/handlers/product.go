package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"hw5/models"
	"hw5/storage"
)

// ErrorResponse models the error schema defined in the OpenAPI specification.
type ErrorResponse struct {
	Error   string  `json:"error"`
	Message string  `json:"message"`
	Details *string `json:"details,omitempty"`
}

// Handler exposes HTTP handlers for product operations.
type Handler struct {
	store *storage.MemoryStore
}

// NewHandler creates a Handler backed by the provided store.
func NewHandler(store *storage.MemoryStore) *Handler {
	return &Handler{store: store}
}

// RegisterRoutes wires product routes onto the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/products", func(router chi.Router) {
		router.Get("/{productId}", h.handleGetProduct)
		router.Post("/{productId}/details", h.handleUpsertProduct)
	})
}

func (h *Handler) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := parseProductID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	product, err := h.store.GetProduct(productID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "The requested product does not exist")
			return
		}
		log.Printf("ERROR: failed to retrieve product %d: %v", productID, err)
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve product")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) handleUpsertProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := parseProductID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	var payload models.Product
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON payload")
		return
	}

	if validationErr := validateProductPayload(productID, payload); validationErr != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_INPUT", validationErr.Error())
		return
	}

	h.store.UpsertProduct(payload)

	w.WriteHeader(http.StatusNoContent)
}

func parseProductID(r *http.Request) (int, error) {
	value := chi.URLParam(r, "productId")
	if value == "" {
		return 0, errors.New("productId path parameter is required")
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("productId must be an integer")
	}

	if id < 1 {
		return 0, errors.New("productId must be a positive integer")
	}

	return id, nil
}

func validateProductPayload(expectedID int, product models.Product) error {
	if product.ProductID != expectedID {
		return errors.New("product_id in the body must match the productId path parameter")
	}

	if product.SKU == "" {
		return errors.New("sku is required")
	}
	if len(product.SKU) > 100 {
		return errors.New("sku must be 100 characters or fewer")
	}
	if product.Manufacturer == "" {
		return errors.New("manufacturer is required")
	}
	if len(product.Manufacturer) > 200 {
		return errors.New("manufacturer must be 200 characters or fewer")
	}
	if product.CategoryID < 1 {
		return errors.New("category_id must be a positive integer")
	}
	if product.Weight < 0 {
		return errors.New("weight must be zero or a positive integer")
	}
	if product.SomeOtherID < 1 {
		return errors.New("some_other_id must be a positive integer")
	}

	return nil
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Error: code, Message: message})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
