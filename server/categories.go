package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
)

// GetAllCategories returns all categories in the database
// @summary returns all categories
// @router /categories [get]
// @tags categories
// @id getCategories
// @produce application/json
// @success 200 {array} model.Category "OK"
// @failure 500 {object} model.Errors "Server error"
func (app App) GetAllCategories(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	cats, errors := app.api.GetCategories(app.Context())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(cats)
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
// @Param id path integer true "Category ID"
// @produce application/json
// @success 200 {object} model.Category "OK"
// @failure 404 {object} model.Errors "Not found"
// @failure 500 {object} model.Errors "Server error"
func (app App) GetCategory(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	category, errors := app.api.GetCategory(app.Context(), req.URL.String())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(category)
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
// @Param category body model.CategoryIn true "Add Category"
// @accept application/json
// @produce application/json
// @success 201 {object} model.Category "OK"
// @failure 415 {object} model.Errors "Bad Content-Type"
// @failure 500 {object} model.Errors "Server error"
func (app App) PostCategory(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	in := model.CategoryIn{}
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	category, errors := app.api.AddCategory(app.Context(), in)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(category)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PutCategory updates a Category in the database
// @summary updates a Category
// @router /categories/{id} [put]
// @tags categories
// @id updateCategory
// @Param id path integer true "Category ID"
// @Param category body model.Category true "Update Category"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Category "OK"
// @failure 415 {object} model.Errors "Bad Content-Type"
// @failure 500 {object} model.Errors "Server error"
func (app App) PutCategory(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	var in model.Category
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	category, errors := app.api.UpdateCategory(app.Context(), req.URL.String(), in)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
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
// @Param id path integer true "Category ID"
// @success 204 "OK"
// @failure 500 {object} model.Errors "Server error"
func (app App) DeleteCategory(w http.ResponseWriter, req *http.Request) {
	errors := app.api.DeleteCategory(app.Context(), req.URL.String())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
