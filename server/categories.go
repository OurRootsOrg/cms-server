package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"

	"github.com/jancona/ourroots/model"
)

// GetAllCategories returns all categories in the database
// @summary returns all categories
// @router /categories [get]
// @tags categories
// @id getCategories
// @produce application/json
// @success 200 {array} model.Category "OK"
// @failure 500 {object} model.Error "Server error"
func (app App) GetAllCategories(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	v := make([]model.Category, 0, len(app.Categories))

	for _, value := range app.Categories {
		v = append(v, value)
	}
	err := enc.Encode(v)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PostCategory adds a new Category to the database
// @summary adds a new Category
// @router /categories [post]
// @tags categories
// @id addCategory
// @Param category body model.CategoryInput true "Add Category"
// @accept application/json
// @produce application/json
// @success 201 {object} model.Category "OK"
// @failure 415 {object} model.Error "Bad Content-Type"
// @failure 500 {object} model.Error "Server error"
func (app App) PostCategory(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		log.Print(msg)
		errorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("Bad MIME type '%s'", mt))
		return
	}
	if mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		log.Print(msg)
		errorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("Bad MIME type '%s'", mt))
		return
	}
	ci, _ := model.NewCategoryInput("")
	err = json.NewDecoder(req.Body).Decode(&ci)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		log.Print(msg)
		errorResponse(w, http.StatusBadRequest, msg)
		return
	}
	category := model.NewCategory(ci)
	// Add to "database"
	app.Categories[category.ID] = category
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// GetCategory gets a Category from the database
// @summary gets a Category
// @router /categories/{id} [get]
// @tags categories
// @id getCategory
// @Param id path string true "Category ID" format(uuid)
// @produce application/json
// @success 200 {object} model.Category "OK"
// @failure 404 {object} model.Error "Not found"
// @failure 500 {object} model.Error "Server error"
func (app App) GetCategory(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	category, found := app.Categories[req.URL.String()]
	if !found {
		notFound(w, req)
		return
	}
	err := enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PatchCategory updates a Category in the database
// @summary updates a Category
// @router /categories/{id} [patch]
// @tags categories
// @id updateCategory
// @Param id path string true "Category ID" format(uuid)
// @Param category body model.CategoryInput true "Update Category"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Category "OK"
// @failure 415 {object} model.Error "Bad Content-Type"
// @failure 500 {object} model.Error "Server error"
func (app App) PatchCategory(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		log.Print(msg)
		errorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("Bad MIME type '%s'", mt))
		return
	}
	if mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		log.Print(msg)
		errorResponse(w, http.StatusUnsupportedMediaType, fmt.Sprintf("Bad MIME type '%s'", mt))
		return
	}
	_, found := app.Categories[req.URL.String()]
	if !found {
		// Not allowed to add a Category with PATCH
		notFound(w, req)
		return
	}
	var tdi model.CategoryInput
	err = json.NewDecoder(req.Body).Decode(&tdi)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		log.Print(msg)
		errorResponse(w, http.StatusBadRequest, msg)
		return
	}
	category := app.Categories[req.URL.String()]
	category.Name = tdi.Name
	// Add to "database"
	app.Categories[req.URL.String()] = category
	w.Header().Set("Content-Type", contentType)
	enc := json.NewEncoder(w)
	err = enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// DeleteCategory deletes a Category from the database
// @summary deletes a Category
// @router /categories/{id} [delete]
// @tags categories
// @id deleteCategory
// @Param id path string true "Category ID" format(uuid)
// @success 204 {object} model.Category "OK"
// @failure 500 {object} model.Error "Server error"
func (app App) DeleteCategory(w http.ResponseWriter, req *http.Request) {
	delete(app.Categories, req.URL.String())
	w.WriteHeader(http.StatusNoContent)
}
