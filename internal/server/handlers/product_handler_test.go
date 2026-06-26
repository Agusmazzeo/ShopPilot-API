package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/services"
)

// Mock ProductService
type mockProductService struct {
	createProductFunc      func(ctx context.Context, req *services.CreateProductRequest) (*models.Product, error)
	getProductFunc         func(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.Product, error)
	updateProductFunc      func(ctx context.Context, clientID uuid.UUID, productID uuid.UUID, req *services.UpdateProductRequest) error
	deleteProductFunc      func(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) error
	listProductsFunc       func(ctx context.Context, clientID uuid.UUID, shopID *uuid.UUID, page, pageSize int) ([]*models.Product, int, error)
	searchProductsFunc     func(ctx context.Context, clientID uuid.UUID, query string, page, pageSize int) ([]*models.Product, int, error)
	createVariantFunc      func(ctx context.Context, clientID uuid.UUID, productID uuid.UUID, req *services.CreateVariantRequest) (*models.ProductVariant, error)
	getVariantFunc         func(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (*models.ProductVariant, error)
	updateVariantFunc      func(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, req *services.UpdateVariantRequest) error
	deleteVariantFunc      func(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) error
	listVariantsFunc       func(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) ([]*models.ProductVariant, error)
	adjustInventoryFunc    func(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, delta int) error
	setInventoryFunc       func(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, quantity int) error
	checkStockFunc         func(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (int, error)
}

func (m *mockProductService) CreateProduct(ctx context.Context, req *services.CreateProductRequest) (*models.Product, error) {
	return m.createProductFunc(ctx, req)
}

func (m *mockProductService) GetProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) (*models.Product, error) {
	return m.getProductFunc(ctx, clientID, productID)
}

func (m *mockProductService) UpdateProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID, req *services.UpdateProductRequest) error {
	return m.updateProductFunc(ctx, clientID, productID, req)
}

func (m *mockProductService) DeleteProduct(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) error {
	return m.deleteProductFunc(ctx, clientID, productID)
}

func (m *mockProductService) ListProducts(ctx context.Context, clientID uuid.UUID, shopID *uuid.UUID, page, pageSize int) ([]*models.Product, int, error) {
	return m.listProductsFunc(ctx, clientID, shopID, page, pageSize)
}

func (m *mockProductService) SearchProducts(ctx context.Context, clientID uuid.UUID, query string, page, pageSize int) ([]*models.Product, int, error) {
	return m.searchProductsFunc(ctx, clientID, query, page, pageSize)
}

func (m *mockProductService) CreateVariant(ctx context.Context, clientID uuid.UUID, productID uuid.UUID, req *services.CreateVariantRequest) (*models.ProductVariant, error) {
	return m.createVariantFunc(ctx, clientID, productID, req)
}

func (m *mockProductService) GetVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (*models.ProductVariant, error) {
	return m.getVariantFunc(ctx, clientID, variantID)
}

func (m *mockProductService) UpdateVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, req *services.UpdateVariantRequest) error {
	return m.updateVariantFunc(ctx, clientID, variantID, req)
}

func (m *mockProductService) DeleteVariant(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) error {
	return m.deleteVariantFunc(ctx, clientID, variantID)
}

func (m *mockProductService) ListVariants(ctx context.Context, clientID uuid.UUID, productID uuid.UUID) ([]*models.ProductVariant, error) {
	return m.listVariantsFunc(ctx, clientID, productID)
}

func (m *mockProductService) AdjustInventory(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, delta int) error {
	return m.adjustInventoryFunc(ctx, clientID, variantID, delta)
}

func (m *mockProductService) SetInventory(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID, quantity int) error {
	return m.setInventoryFunc(ctx, clientID, variantID, quantity)
}

func (m *mockProductService) CheckStock(ctx context.Context, clientID uuid.UUID, variantID uuid.UUID) (int, error) {
	return m.checkStockFunc(ctx, clientID, variantID)
}

// Helper functions

func setupProductHandler(mockService services.ProductService) *ProductHandler {
	return NewProductHandler(mockService)
}

func setupRouter(handler *ProductHandler) *mux.Router {
	router := mux.NewRouter()

	// Product routes - more specific routes first
	router.HandleFunc("/api/v1/clients/{clientId}/products/search", handler.Search).Methods("GET")
	router.HandleFunc("/api/v1/clients/{clientId}/products", handler.Create).Methods("POST")
	router.HandleFunc("/api/v1/clients/{clientId}/products", handler.List).Methods("GET")
	router.HandleFunc("/api/v1/clients/{clientId}/products/{id}", handler.Get).Methods("GET")
	router.HandleFunc("/api/v1/clients/{clientId}/products/{id}", handler.Update).Methods("PUT")
	router.HandleFunc("/api/v1/clients/{clientId}/products/{id}", handler.Delete).Methods("DELETE")

	// Variant routes
	router.HandleFunc("/api/v1/clients/{clientId}/products/{productId}/variants", handler.CreateVariant).Methods("POST")
	router.HandleFunc("/api/v1/clients/{clientId}/variants/{id}", handler.GetVariant).Methods("GET")
	router.HandleFunc("/api/v1/clients/{clientId}/variants/{id}", handler.UpdateVariant).Methods("PUT")
	router.HandleFunc("/api/v1/clients/{clientId}/variants/{id}", handler.DeleteVariant).Methods("DELETE")
	router.HandleFunc("/api/v1/clients/{clientId}/products/{productId}/variants", handler.ListVariants).Methods("GET")

	// Inventory routes
	router.HandleFunc("/api/v1/clients/{clientId}/variants/{id}/inventory/adjust", handler.AdjustInventory).Methods("POST")
	router.HandleFunc("/api/v1/clients/{clientId}/variants/{id}/inventory", handler.SetInventory).Methods("PUT")
	router.HandleFunc("/api/v1/clients/{clientId}/variants/{id}/inventory", handler.CheckStock).Methods("GET")

	return router
}

// Product tests

func TestProductHandler_Create(t *testing.T) {
	clientID := uuid.New()
	shopID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockService    func() *mockProductService
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful creation",
			requestBody: CreateProductRequestDTO{
				ShopID:      shopID.String(),
				Code:        "PROD001",
				Name:        "Test Product",
				Description: "Test description",
				IsActive:    true,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					createProductFunc: func(ctx context.Context, req *services.CreateProductRequest) (*models.Product, error) {
						return &models.Product{
							ID:          productID,
							ClientID:    clientID,
							ShopID:      shopID,
							Code:        "PROD001",
							Name:        "Test Product",
							Description: "Test description",
							IsActive:    true,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: "invalid json",
			mockService: func() *mockProductService {
				return &mockProductService{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "missing required fields",
			requestBody: CreateProductRequestDTO{
				ShopID: shopID.String(),
			},
			mockService: func() *mockProductService {
				return &mockProductService{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_FIELDS",
		},
		{
			name: "service error",
			requestBody: CreateProductRequestDTO{
				ShopID:   shopID.String(),
				Code:     "PROD001",
				Name:     "Test Product",
				IsActive: true,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					createProductFunc: func(ctx context.Context, req *services.CreateProductRequest) (*models.Product, error) {
						return nil, errors.New("service error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "CREATE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest("POST", "/api/v1/clients/"+clientID.String()+"/products", &body)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response APIResponse
				json.NewDecoder(w.Body).Decode(&response)
				if response.Error == nil || response.Error.Code != tt.expectedError {
					t.Errorf("expected error code %s, got %v", tt.expectedError, response.Error)
				}
			}
		})
	}
}

func TestProductHandler_Get(t *testing.T) {
	clientID := uuid.New()
	productID := uuid.New()
	shopID := uuid.New()

	tests := []struct {
		name           string
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful retrieval",
			mockService: func() *mockProductService {
				return &mockProductService{
					getProductFunc: func(ctx context.Context, cID uuid.UUID, pID uuid.UUID) (*models.Product, error) {
						return &models.Product{
							ID:       productID,
							ClientID: clientID,
							ShopID:   shopID,
							Code:     "PROD001",
							Name:     "Test Product",
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "product not found",
			mockService: func() *mockProductService {
				return &mockProductService{
					getProductFunc: func(ctx context.Context, cID uuid.UUID, pID uuid.UUID) (*models.Product, error) {
						return nil, errors.New("not found")
					},
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			req := httptest.NewRequest("GET", "/api/v1/clients/"+clientID.String()+"/products/"+productID.String(), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_Update(t *testing.T) {
	clientID := uuid.New()
	productID := uuid.New()
	newName := "Updated Product"

	tests := []struct {
		name           string
		requestBody    interface{}
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful update",
			requestBody: UpdateProductRequestDTO{
				Name: &newName,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					updateProductFunc: func(ctx context.Context, cID uuid.UUID, pID uuid.UUID, req *services.UpdateProductRequest) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			requestBody: UpdateProductRequestDTO{
				Name: &newName,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					updateProductFunc: func(ctx context.Context, cID uuid.UUID, pID uuid.UUID, req *services.UpdateProductRequest) error {
						return errors.New("update failed")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest("PUT", "/api/v1/clients/"+clientID.String()+"/products/"+productID.String(), &body)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_Delete(t *testing.T) {
	clientID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name           string
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful deletion",
			mockService: func() *mockProductService {
				return &mockProductService{
					deleteProductFunc: func(ctx context.Context, cID uuid.UUID, pID uuid.UUID) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			mockService: func() *mockProductService {
				return &mockProductService{
					deleteProductFunc: func(ctx context.Context, cID uuid.UUID, pID uuid.UUID) error {
						return errors.New("delete failed")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			req := httptest.NewRequest("DELETE", "/api/v1/clients/"+clientID.String()+"/products/"+productID.String(), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_List(t *testing.T) {
	clientID := uuid.New()
	shopID := uuid.New()

	tests := []struct {
		name           string
		queryParams    string
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name:        "successful list all products",
			queryParams: "",
			mockService: func() *mockProductService {
				return &mockProductService{
					listProductsFunc: func(ctx context.Context, cID uuid.UUID, sID *uuid.UUID, page, pageSize int) ([]*models.Product, int, error) {
						return []*models.Product{
							{ID: uuid.New(), ClientID: clientID, Code: "PROD001", Name: "Product 1"},
							{ID: uuid.New(), ClientID: clientID, Code: "PROD002", Name: "Product 2"},
						}, 2, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "list with shop filter",
			queryParams: "?shopId=" + shopID.String(),
			mockService: func() *mockProductService {
				return &mockProductService{
					listProductsFunc: func(ctx context.Context, cID uuid.UUID, sID *uuid.UUID, page, pageSize int) ([]*models.Product, int, error) {
						if sID == nil || *sID != shopID {
							t.Error("expected shopID to be passed")
						}
						return []*models.Product{
							{ID: uuid.New(), ClientID: clientID, ShopID: shopID, Code: "PROD001", Name: "Product 1"},
						}, 1, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "list with pagination",
			queryParams: "?page=2&pageSize=10",
			mockService: func() *mockProductService {
				return &mockProductService{
					listProductsFunc: func(ctx context.Context, cID uuid.UUID, sID *uuid.UUID, page, pageSize int) ([]*models.Product, int, error) {
						if page != 2 || pageSize != 10 {
							t.Errorf("expected page=2, pageSize=10, got page=%d, pageSize=%d", page, pageSize)
						}
						return []*models.Product{}, 0, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service error",
			queryParams: "",
			mockService: func() *mockProductService {
				return &mockProductService{
					listProductsFunc: func(ctx context.Context, cID uuid.UUID, sID *uuid.UUID, page, pageSize int) ([]*models.Product, int, error) {
						return nil, 0, errors.New("list failed")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			req := httptest.NewRequest("GET", "/api/v1/clients/"+clientID.String()+"/products"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_Search(t *testing.T) {
	clientID := uuid.New()

	tests := []struct {
		name           string
		queryParams    string
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name:        "successful search",
			queryParams: "?q=test",
			mockService: func() *mockProductService {
				return &mockProductService{
					searchProductsFunc: func(ctx context.Context, cID uuid.UUID, query string, page, pageSize int) ([]*models.Product, int, error) {
						return []*models.Product{
							{ID: uuid.New(), ClientID: clientID, Code: "PROD001", Name: "Test Product"},
						}, 1, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "missing query parameter",
			queryParams: "",
			mockService: func() *mockProductService {
				return &mockProductService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			req := httptest.NewRequest("GET", "/api/v1/clients/"+clientID.String()+"/products/search"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// Variant tests

func TestProductHandler_CreateVariant(t *testing.T) {
	clientID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful creation",
			requestBody: CreateVariantRequestDTO{
				SKU:      "SKU001",
				Name:     "Variant 1",
				Price:    99.99,
				Quantity: 10,
				IsActive: true,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					createVariantFunc: func(ctx context.Context, cID uuid.UUID, pID uuid.UUID, req *services.CreateVariantRequest) (*models.ProductVariant, error) {
						return &models.ProductVariant{
							ID:        variantID,
							ClientID:  clientID,
							ProductID: productID,
							SKU:       "SKU001",
							Name:      "Variant 1",
							Price:     99.99,
							Quantity:  10,
							IsActive:  true,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			requestBody: CreateVariantRequestDTO{
				Price: 99.99,
			},
			mockService: func() *mockProductService {
				return &mockProductService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest("POST", "/api/v1/clients/"+clientID.String()+"/products/"+productID.String()+"/variants", &body)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_GetVariant(t *testing.T) {
	clientID := uuid.New()
	variantID := uuid.New()

	tests := []struct {
		name           string
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful retrieval",
			mockService: func() *mockProductService {
				return &mockProductService{
					getVariantFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID) (*models.ProductVariant, error) {
						return &models.ProductVariant{
							ID:       variantID,
							ClientID: clientID,
							SKU:      "SKU001",
							Name:     "Variant 1",
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "variant not found",
			mockService: func() *mockProductService {
				return &mockProductService{
					getVariantFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID) (*models.ProductVariant, error) {
						return nil, errors.New("not found")
					},
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			req := httptest.NewRequest("GET", "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String(), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_UpdateVariant(t *testing.T) {
	clientID := uuid.New()
	variantID := uuid.New()
	newPrice := 199.99

	tests := []struct {
		name           string
		requestBody    interface{}
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful update",
			requestBody: UpdateVariantRequestDTO{
				Price: &newPrice,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					updateVariantFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID, req *services.UpdateVariantRequest) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest("PUT", "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String(), &body)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_DeleteVariant(t *testing.T) {
	clientID := uuid.New()
	variantID := uuid.New()

	tests := []struct {
		name           string
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful deletion",
			mockService: func() *mockProductService {
				return &mockProductService{
					deleteVariantFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "cannot delete last variant",
			mockService: func() *mockProductService {
				return &mockProductService{
					deleteVariantFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID) error {
						return errors.New("cannot delete the last variant")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			req := httptest.NewRequest("DELETE", "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String(), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_ListVariants(t *testing.T) {
	clientID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name           string
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful list",
			mockService: func() *mockProductService {
				return &mockProductService{
					listVariantsFunc: func(ctx context.Context, cID uuid.UUID, pID uuid.UUID) ([]*models.ProductVariant, error) {
						return []*models.ProductVariant{
							{ID: uuid.New(), ClientID: clientID, ProductID: productID, SKU: "SKU001", Name: "Variant 1"},
							{ID: uuid.New(), ClientID: clientID, ProductID: productID, SKU: "SKU002", Name: "Variant 2"},
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			req := httptest.NewRequest("GET", "/api/v1/clients/"+clientID.String()+"/products/"+productID.String()+"/variants", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// Inventory tests

func TestProductHandler_AdjustInventory(t *testing.T) {
	clientID := uuid.New()
	variantID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful adjustment - positive delta",
			requestBody: AdjustInventoryRequestDTO{
				Delta: 10,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					adjustInventoryFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID, delta int) error {
						if delta != 10 {
							t.Errorf("expected delta 10, got %d", delta)
						}
						return nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful adjustment - negative delta",
			requestBody: AdjustInventoryRequestDTO{
				Delta: -5,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					adjustInventoryFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID, delta int) error {
						if delta != -5 {
							t.Errorf("expected delta -5, got %d", delta)
						}
						return nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "insufficient inventory",
			requestBody: AdjustInventoryRequestDTO{
				Delta: -100,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					adjustInventoryFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID, delta int) error {
						return errors.New("insufficient inventory")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest("POST", "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/inventory/adjust", &body)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_SetInventory(t *testing.T) {
	clientID := uuid.New()
	variantID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockService    func() *mockProductService
		expectedStatus int
	}{
		{
			name: "successful set",
			requestBody: SetInventoryRequestDTO{
				Quantity: 50,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					setInventoryFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID, quantity int) error {
						if quantity != 50 {
							t.Errorf("expected quantity 50, got %d", quantity)
						}
						return nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			requestBody: SetInventoryRequestDTO{
				Quantity: -10,
			},
			mockService: func() *mockProductService {
				return &mockProductService{
					setInventoryFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID, quantity int) error {
						return errors.New("quantity cannot be negative")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest("PUT", "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/inventory", &body)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProductHandler_CheckStock(t *testing.T) {
	clientID := uuid.New()
	variantID := uuid.New()

	tests := []struct {
		name           string
		mockService    func() *mockProductService
		expectedStatus int
		expectedStock  int
	}{
		{
			name: "successful check",
			mockService: func() *mockProductService {
				return &mockProductService{
					checkStockFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID) (int, error) {
						return 42, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedStock:  42,
		},
		{
			name: "variant not found",
			mockService: func() *mockProductService {
				return &mockProductService{
					checkStockFunc: func(ctx context.Context, cID uuid.UUID, vID uuid.UUID) (int, error) {
						return 0, errors.New("variant not found")
					},
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupProductHandler(tt.mockService())
			router := setupRouter(handler)

			req := httptest.NewRequest("GET", "/api/v1/clients/"+clientID.String()+"/variants/"+variantID.String()+"/inventory", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response APIResponse
				json.NewDecoder(w.Body).Decode(&response)
				data := response.Data.(map[string]interface{})
				if int(data["quantity"].(float64)) != tt.expectedStock {
					t.Errorf("expected stock %d, got %v", tt.expectedStock, data["quantity"])
				}
			}
		})
	}
}
