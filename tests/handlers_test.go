package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"shopping-api/api"
	"shopping-api/models"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetItemsList(t *testing.T) {
	models.Cart.List = map[int]*models.ShoppingItem{
		1: {ID: 1, Name: "Молоко", Category: "Продукты", Price: 89.50, Quantity: 2, Purchased: false},
		2: {ID: 2, Name: "Хлеб", Category: "Продукты", Price: 45.00, Quantity: 1, Purchased: false},
		3: {ID: 3, Name: "Яблоки", Category: "Фрукты", Price: 120.00, Quantity: 3, Purchased: true},
	}
	r := chi.NewRouter()
	r.Get("/items", api.GetItemsList)
	testCases := []struct {
		desc         string
		expectedCode int
		expectedLen  int
	}{
		{"right list", http.StatusOK, len(models.Cart.List)},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/items", nil)

			r.ServeHTTP(resp, req)

			require.Equal(t, tc.expectedCode, resp.Code)
			require.True(t, json.Valid(resp.Body.Bytes()))

			var decodedResp map[int]*models.ShoppingItem
			if err := json.NewDecoder(resp.Body).Decode(&decodedResp); err != nil {
				t.Errorf("ошибка при десериализации ответа сервера: %v", err)
			}

			assert.Len(t, decodedResp, tc.expectedLen)
		})
	}
}
func TestCreateItem(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/items", api.CreateItem)

	goodRequestBody := models.ShoppingItem{
		Name:     "Кофе",
		Category: "Продукты",
		Price:    350.00,
		Quantity: 1,
	}
	longName := strings.Repeat("a", 200)
	longNameRequestBody := models.ShoppingItem{
		Name:     longName,
		Category: "Продукты",
		Price:    350.00,
		Quantity: 1,
	}
	zeroPriceRequestBody := models.ShoppingItem{
		Name:     "Кофе",
		Category: "Продукты",
		Price:    0,
		Quantity: 1,
	}
	negativePriceRequestBody := models.ShoppingItem{
		Name:     "Кофе",
		Category: "Продукты",
		Price:    -100,
		Quantity: 1,
	}
	zeroQuantityRequestBody := models.ShoppingItem{
		Name:     "Кофе",
		Category: "Продукты",
		Price:    350.00,
		Quantity: 0,
	}
	negativeQuantityRequestBody := models.ShoppingItem{
		Name:     "Кофе",
		Category: "Продукты",
		Price:    350.00,
		Quantity: -1,
	}

	expectedOK := models.APIStatus{
		Code:    "STATUS_OK",
		Message: "товар был успешно добавлен",
	}
	expectedErrLongName := models.APIStatus{
		Code:    "VALIDATION_ERROR",
		Message: "название товара не может быть пустым или содержать более 100 символов",
	}
	expectedErrZeroPrice := models.APIStatus{
		Code:    "VALIDATION_ERROR",
		Message: "цена должна быть больше нуля",
	}
	expectedErrNegativePrice := models.APIStatus{
		Code:    "VALIDATION_ERROR",
		Message: "цена должна быть больше нуля",
	}
	expectedErrZeroQuantity := models.APIStatus{
		Code:    "VALIDATION_ERROR",
		Message: "количество должно быть равно или больше единицы",
	}
	expectedErrNegativeQuantity := models.APIStatus{
		Code:    "VALIDATION_ERROR",
		Message: "количество должно быть равно или больше единицы",
	}

	testCases := []struct {
		desc         string
		item         models.ShoppingItem
		expectedCode int
		expectedResp models.APIStatus
	}{
		{"add good item", goodRequestBody, http.StatusCreated, expectedOK},
		{"long name", longNameRequestBody, http.StatusBadRequest, expectedErrLongName},
		{"zero price", zeroPriceRequestBody, http.StatusBadRequest, expectedErrZeroPrice},
		{"negative price", negativePriceRequestBody, http.StatusBadRequest, expectedErrNegativePrice},
		{"zero quantity", zeroQuantityRequestBody, http.StatusBadRequest, expectedErrZeroQuantity},
		{"negative quantity", negativeQuantityRequestBody, http.StatusBadRequest, expectedErrNegativeQuantity},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			jsonReqBody, err := json.Marshal(tc.item)
			if err != nil {
				t.Errorf("ошибка при сериализации мокового объекта: %v", err)
			}
			req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(jsonReqBody))
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			require.Equal(t, tc.expectedCode, resp.Code)
			require.True(t, json.Valid(resp.Body.Bytes()))

			var decodedResponse models.APIStatus
			if err := json.NewDecoder(resp.Body).Decode(&decodedResponse); err != nil {
				t.Errorf("ошибка при десериализации тела ответа: %v", err)
			}

			assert.Equal(t, tc.expectedResp, decodedResponse)
		})
	}
}

func TestGetItem(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/items/{id}", api.GetItem)
	// Предварительная подготовка данных
	models.Cart.List = map[int]*models.ShoppingItem{
		5: {ID: 5, Name: "Бананы", Category: "Фрукты", Price: 180.00, Quantity: 2, Purchased: false},
	}
	expectedGoodResponse := models.ShoppingItem{
		ID:        5,
		Name:      "Бананы",
		Category:  "Фрукты",
		Price:     180.00,
		Quantity:  2,
		Purchased: false,
	}
	expectedNotFoundResponse := models.APIStatus{
		Code:    "NOT_FOUND_ERROR",
		Message: "товара с таким id не найдено",
	}
	expectedInternalErrorResponse := models.APIStatus{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "ошибка при конвертации id в число",
	}
	testCases := []struct {
		desc           string
		id             any
		expectedStatus int
		expectedResp   any
	}{
		{"get good item", 5, http.StatusOK, expectedGoodResponse},
		{"not found item", 999, http.StatusNotFound, expectedNotFoundResponse},
		{"invalid type id item", "adada", http.StatusInternalServerError, expectedInternalErrorResponse},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			target := fmt.Sprintf("/items/%v", tc.id)
			req := httptest.NewRequest(http.MethodGet, target, nil)
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			require.Equal(t, tc.expectedStatus, resp.Code)
			require.True(t, json.Valid(resp.Body.Bytes()))

			if tc.expectedStatus == http.StatusOK {
				var decodedResponse models.ShoppingItem
				if err := json.NewDecoder(resp.Body).Decode(&decodedResponse); err != nil {
					t.Errorf("ошибка при десериализации тела ответа: %v", err)
				}
				assert.Equal(t, tc.expectedResp, decodedResponse)
				return
			}
			var decodedResponse models.APIStatus
			if err := json.NewDecoder(resp.Body).Decode(&decodedResponse); err != nil {
				t.Errorf("ошибка при десериализации тела ответа: %v", err)
			}
			assert.Equal(t, tc.expectedResp, decodedResponse)
		})
	}
}

func TestFullUpdateItem(t *testing.T) {
	r := chi.NewRouter()
	r.Put("/items/{id}", api.FullUpdateItem)
	// Предварительная подготовка
	models.Cart.List = map[int]*models.ShoppingItem{
		7: {ID: 7, Name: "Оригинальное название", Category: "Старая категория", Price: 100.00, Quantity: 5},
	}
	goodRequestBody := models.ShoppingItem{
		Name:     "новое название",
		Category: "New cat",
		Price:    200,
		Quantity: 300,
	}
	emptyNameRequestBody := models.ShoppingItem{
		Name:     "",
		Category: "New cat",
		Price:    200,
		Quantity: 300,
	}
	negativePriceRequestBody := models.ShoppingItem{
		Name:     "новое название",
		Category: "New cat",
		Price:    -200,
		Quantity: 300,
	}
	expectedGoodStatus := models.APIStatus{
		Code:    "STATUS_OK",
		Message: "товар успешно обновлён",
	}
	expectedNotFoundStatus := models.APIStatus{Code: "NOT_FOUND_ERROR", Message: "товара с таким id не найдено"}
	expectedEmptyName := models.APIStatus{Code: "VALIDATION_ERROR", Message: "название товара не может быть пустым или содержать более 100 символов"}
	expectedNegativePrice := models.APIStatus{Code: "VALIDATION_ERROR", Message: "цена должна быть больше нуля"}
	testCases := []struct {
		desc           string
		id             int
		requestBody    models.ShoppingItem
		expectedCode   int
		expectedStatus models.APIStatus
	}{
		{"good item", 7, goodRequestBody, http.StatusOK, expectedGoodStatus},
		{"not found item", 999, goodRequestBody, http.StatusNotFound, expectedNotFoundStatus},
		{"empty name item", 7, emptyNameRequestBody, http.StatusBadRequest, expectedEmptyName},
		{"negative price item", 7, negativePriceRequestBody, http.StatusBadRequest, expectedNegativePrice},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			jsonReqBody, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Errorf("ошибка при сериализации тела запроса: %v", err)
			}
			target := fmt.Sprintf("/items/%d", tc.id)
			req := httptest.NewRequest(http.MethodPut, target, bytes.NewReader(jsonReqBody))
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			require.Equal(t, tc.expectedCode, resp.Code)
			require.True(t, json.Valid(resp.Body.Bytes()))

			var decodedResponse models.APIStatus
			if err := json.NewDecoder(resp.Body).Decode(&decodedResponse); err != nil {
				t.Errorf("ошибка при десериализации тела ответа: %v", err)
			}
			assert.Equal(t, tc.expectedStatus, decodedResponse)

			if tc.expectedCode == http.StatusOK {
				assert.Equal(t, "новое название", models.Cart.List[7].Name)
			}
		})
	}
}

func TestPartlyUpdateItems(t *testing.T) {
	r := chi.NewRouter()
	r.Patch("/items/{id}", api.PartlyUpdateItems)

	models.Cart.List = map[int]*models.ShoppingItem{
		10: {ID: 10, Name: "Товар", Category: "Продукты", Price: 50.00, Quantity: 2},
	}
	onlyPriceGoodRequestBody := models.UpdateShoppingItem{
		Price: func() *float64 { p := 150.00; return &p }(),
	}
	fullUpdateRequestBody := models.UpdateShoppingItem{
		Name:     func() *string { n := "новое название"; return &n }(),
		Category: func() *string { c := "новая категория"; return &c }(),
		Price:    func() *float64 { p := 150.00; return &p }(),
		Quantity: func() *int { q := 1; return &q }(),
	}
	expectedOkStatus := models.APIStatus{Code: "STATUS_OK", Message: "товар успешно обновлён"}
	expectedNotFoundStatus := models.APIStatus{Code: "NOT_FOUND_ERROR", Message: "товара с таким id не найдено"}

	testCases := []struct {
		desc           string
		id             int
		requestBody    models.UpdateShoppingItem
		expectedCode   int
		expectedStatus models.APIStatus
	}{
		{"good item", 10, onlyPriceGoodRequestBody, http.StatusOK, expectedOkStatus},
		{"full update item", 10, fullUpdateRequestBody, http.StatusOK, expectedOkStatus},
		{"not found item", 999, fullUpdateRequestBody, http.StatusNotFound, expectedNotFoundStatus},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			jsonReqBody, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Errorf("ошибка при сериализации тела запроса: %v", err)
			}
			target := fmt.Sprintf("/items/%d", tc.id)
			req := httptest.NewRequest(http.MethodPatch, target, bytes.NewReader(jsonReqBody))
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			require.Equal(t, tc.expectedCode, resp.Code)
			require.True(t, json.Valid(resp.Body.Bytes()))

			var decodedResponse models.APIStatus
			if err := json.NewDecoder(resp.Body).Decode(&decodedResponse); err != nil {
				t.Errorf("ошибка при десериализации тела ответа: %v", err)
			}
			assert.Equal(t, tc.expectedStatus, decodedResponse)

			if tc.requestBody == onlyPriceGoodRequestBody {
				assert.Equal(t, 150.00, models.Cart.List[10].Price)
			}
			if tc.requestBody == fullUpdateRequestBody && tc.expectedCode == http.StatusOK {
				assert.Equal(t, "новое название", models.Cart.List[10].Name)
				assert.Equal(t, "новая категория", models.Cart.List[10].Category)
				assert.Equal(t, 150.00, models.Cart.List[10].Price)
				assert.Equal(t, 1, models.Cart.List[10].Quantity)
			}
		})
	}
}

func TestDeleteItem(t *testing.T) {
	r := chi.NewRouter()
	r.Delete("/items/{id}", api.DeleteItem)

	// Предварительная подготовка
	models.Cart.List = map[int]*models.ShoppingItem{
		12: {ID: 12, Name: "Удаляемый товар", Category: "Продукты", Price: 75.00, Quantity: 3},
		13: {ID: 13, Name: "Удаляемый товар2", Category: "Продукты", Price: 100.00, Quantity: 1},
	}
	expectedOkStatus := models.APIStatus{Code: "STATUS_OK", Message: "товар успешно удалён"}
	expectedNotFound := models.APIStatus{Code: "NOT_FOUND_ERROR", Message: "товара с таким id не найдено"}

	testCases := []struct {
		desc           string
		id             int
		expectedCode   int
		expectedStatus models.APIStatus
	}{
		{"good delete item", 12, http.StatusOK, expectedOkStatus},
		{"not found delete item", 999, http.StatusNotFound, expectedNotFound},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			target := fmt.Sprintf("/items/%d", tc.id)
			req := httptest.NewRequest(http.MethodDelete, target, nil)
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			require.Equal(t, tc.expectedCode, resp.Code)
			require.True(t, json.Valid(resp.Body.Bytes()))

			var decodedResponse models.APIStatus
			if err := json.NewDecoder(resp.Body).Decode(&decodedResponse); err != nil {
				t.Errorf("ошибка при десериализации тела ответа: %v", err)
			}
			assert.Equal(t, tc.expectedStatus, decodedResponse)
		})
	}
}

func TestGetCategoryItems(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/items/category/{category}", api.GetCategoryItems)
	models.Cart.List = map[int]*models.ShoppingItem{
		13: {ID: 13, Name: "Молоко", Category: "Продукты", Price: 89.50, Quantity: 2},
		14: {ID: 14, Name: "Хлеб", Category: "Продукты", Price: 45.00, Quantity: 1},
		15: {ID: 15, Name: "Яблоки", Category: "Фрукты", Price: 120.00, Quantity: 3},
	}
	expectedGoodByProductCategory := []*models.ShoppingItem{
		{ID: 13, Name: "Молоко", Category: "Продукты", Price: 89.50, Quantity: 2},
		{ID: 14, Name: "Хлеб", Category: "Продукты", Price: 45.00, Quantity: 1},
	}
	expectedNotFoundCategoryStatus := models.APIStatus{Code: "NOT_FOUND_ERROR", Message: "товаров с такой категорией не найдено"}
	testCases := []struct {
		desc             string
		category         string
		expectedCode     int
		expectedResponse any
	}{
		{"good products category", "Продукты", http.StatusOK, expectedGoodByProductCategory},
		{"not found category", "Вещи", http.StatusNotFound, expectedNotFoundCategoryStatus},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			target := fmt.Sprintf("/items/category/%s", tc.category)
			req := httptest.NewRequest(http.MethodGet, target, nil)
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			require.Equal(t, tc.expectedCode, resp.Code)
			require.True(t, json.Valid(resp.Body.Bytes()))

			if tc.expectedCode == http.StatusOK {
				var decodedResponse []*models.ShoppingItem
				if err := json.NewDecoder(resp.Body).Decode(&decodedResponse); err != nil {
					t.Errorf("ошибка при десериализации тела ответа: %v", err)
				}
				assert.Equal(t, tc.expectedResponse, decodedResponse)
				return
			}
			var decodedResponse models.APIStatus
			if err := json.NewDecoder(resp.Body).Decode(&decodedResponse); err != nil {
				t.Errorf("ошибка при десериализации тела ответа: %v", err)
			}
			assert.Equal(t, tc.expectedResponse, decodedResponse)
		})
	}
}

func TestJsonResponse(t *testing.T) {
	statusOkResponseData := models.APIStatus{Code: "STATUS_OK", Message: "Успешно выполнено"}
	serilErrResponseData := models.APIStatus{Code: "STATUS_INTERNAL_SERVER_ERROR", Message: "ошибка при сериализации ответа"}

	expectedOkJson := `{"code":"STATUS_OK","message":"Успешно выполнено"}`
	expectedInternalErrorJson := `{"code":"STATUS_INTERNAL_SERVER_ERROR","message":"ошибка при сериализации ответа"}`

	testCases := []struct {
		desc         string
		status       int
		responseData models.APIStatus
		expectedCode int
		expectedJson string
	}{
		{"status ok", http.StatusOK, statusOkResponseData, http.StatusOK, expectedOkJson},
		{"status internal server error", http.StatusInternalServerError, serilErrResponseData, http.StatusInternalServerError, expectedInternalErrorJson},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			resp := httptest.NewRecorder()
			api.JSONResponse(resp, tc.status, tc.responseData)

			require.Equal(t, tc.expectedCode, resp.Code)
			require.True(t, json.Valid(resp.Body.Bytes()))
			assert.Equal(t, tc.expectedJson, strings.TrimSpace(resp.Body.String()))
		})
	}
}
