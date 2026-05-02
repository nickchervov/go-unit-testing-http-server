package api

import (
	"encoding/json"
	"net/http"
	"shopping-api/models"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		JSONResponse(w, http.StatusInternalServerError, models.APIStatus{Code: "INTERNAL_SERVER_ERROR", Message: "ошибка при сериализации ответа"})
		return
	}
}

// Получить все товары
func GetItemsList(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, http.StatusOK, models.Cart.GetAllItems())
}

// Добавить новый товар
func CreateItem(w http.ResponseWriter, r *http.Request) {
	var item models.ShoppingItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		JSONResponse(w, http.StatusInternalServerError, models.APIStatus{Code: "INTERNAL_SERVER_ERROR", Message: "ошибка при десериализации тела запроса"})
		return
	}

	if item.Name == "" || len(item.Name) > 100 {
		JSONResponse(w, http.StatusBadRequest, models.APIStatus{Code: "VALIDATION_ERROR", Message: "название товара не может быть пустым или содержать более 100 символов"})
		return
	}
	if item.Price <= 0 {
		JSONResponse(w, http.StatusBadRequest, models.APIStatus{Code: "VALIDATION_ERROR", Message: "цена должна быть больше нуля"})
		return
	}
	if item.Quantity < 1 {
		JSONResponse(w, http.StatusBadRequest, models.APIStatus{Code: "VALIDATION_ERROR", Message: "количество должно быть равно или больше единицы"})
		return
	}

	models.Cart.AddItem(item)
	JSONResponse(w, http.StatusCreated, models.APIStatus{Code: "STATUS_OK", Message: "товар был успешно добавлен"})
}

// Получить товар по ID
func GetItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, models.APIStatus{Code: "INTERNAL_SERVER_ERROR", Message: "ошибка при конвертации id в число"})
		return
	}

	item, ok := models.Cart.List[id]
	if !ok {
		JSONResponse(w, http.StatusNotFound, models.APIStatus{Code: "NOT_FOUND_ERROR", Message: "товара с таким id не найдено"})
		return
	}
	JSONResponse(w, http.StatusOK, item)
}

// Обновить товар (полное обновление)
func FullUpdateItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, models.APIStatus{Code: "INTERNAL_SERVER_ERROR", Message: "ошибка при конвертации id в число"})
		return
	}
	_, ok := models.Cart.List[id]
	if !ok {
		JSONResponse(w, http.StatusNotFound, models.APIStatus{Code: "NOT_FOUND_ERROR", Message: "товара с таким id не найдено"})
		return
	}

	var item models.ShoppingItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		JSONResponse(w, http.StatusInternalServerError, models.APIStatus{Code: "INTERNAL_SERVER_ERROR", Message: "ошибка при десериализации тела запроса"})
		return
	}

	if item.Name == "" || len(item.Name) > 100 {
		JSONResponse(w, http.StatusBadRequest, models.APIStatus{Code: "VALIDATION_ERROR", Message: "название товара не может быть пустым или содержать более 100 символов"})
		return
	}
	if item.Price <= 0 {
		JSONResponse(w, http.StatusBadRequest, models.APIStatus{Code: "VALIDATION_ERROR", Message: "цена должна быть больше нуля"})
		return
	}
	if item.Quantity < 1 {
		JSONResponse(w, http.StatusBadRequest, models.APIStatus{Code: "VALIDATION_ERROR", Message: "количество должно быть равно или больше единицы"})
		return
	}
	item.ID = id

	models.Cart.List[id] = &item
	JSONResponse(w, http.StatusOK, models.APIStatus{Code: "STATUS_OK", Message: "товар успешно обновлён"})
}

// Частично обновить товар (только поля из JSON-запроса) PATCH запрос
func PartlyUpdateItems(w http.ResponseWriter, r *http.Request) {
	var updateItem models.UpdateShoppingItem

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, models.APIStatus{Code: "INTERNAL_SERVER_ERROR", Message: "ошибка при конвертации id в число"})
		return
	}
	_, ok := models.Cart.List[id]
	if !ok {
		JSONResponse(w, http.StatusNotFound, models.APIStatus{Code: "NOT_FOUND_ERROR", Message: "товара с таким id не найдено"})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&updateItem); err != nil {
		JSONResponse(w, http.StatusInternalServerError, models.APIStatus{Code: "INTERNAL_SERVER_ERROR", Message: "ошибка при десериализации тела запроса"})
		return
	}

	if updateItem.Name != nil {
		models.Cart.List[id].Name = *updateItem.Name
	}
	if updateItem.Category != nil {
		models.Cart.List[id].Category = *updateItem.Category
	}
	if updateItem.Price != nil {
		models.Cart.List[id].Price = *updateItem.Price
	}
	if updateItem.Quantity != nil {
		models.Cart.List[id].Quantity = *updateItem.Quantity
	}
	if updateItem.Purchased != nil {
		models.Cart.List[id].Purchased = *updateItem.Purchased
	}
	JSONResponse(w, http.StatusOK, models.APIStatus{Code: "STATUS_OK", Message: "товар успешно обновлён"})
}

// Удалить товар
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, models.APIStatus{Code: "INTERNAL_SERVER_ERROR", Message: "ошибка при конвертации id в число"})
		return
	}

	_, ok := models.Cart.List[id]
	if !ok {
		JSONResponse(w, http.StatusNotFound, models.APIStatus{Code: "NOT_FOUND_ERROR", Message: "товара с таким id не найдено"})
		return
	}

	models.Cart.RemoveItemById(id)
	JSONResponse(w, http.StatusOK, models.APIStatus{Code: "STATUS_OK", Message: "товар успешно удалён"})
}

// Получить товары по категории
func GetCategoryItems(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	var FilteredItemsByCategory []*models.ShoppingItem

	for _, v := range models.Cart.List {
		if v.Category == category {
			FilteredItemsByCategory = append(FilteredItemsByCategory, v)
		}
	}
	if len(FilteredItemsByCategory) == 0 {
		JSONResponse(w, http.StatusNotFound, models.APIStatus{Code: "NOT_FOUND_ERROR", Message: "товаров с такой категорией не найдено"})
		return
	}
	JSONResponse(w, http.StatusOK, FilteredItemsByCategory)
}
