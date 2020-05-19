package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
)

// GetPosts returns all posts in the database
// @summary returns all posts
// @router /posts [get]
// @tags posts
// @id getPosts
// @produce application/json
// @success 200 {array} model.Post "OK"
// @failure 500 {object} model.Errors "Server error"
func (app App) GetPosts(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	cols, errors := app.api.GetPosts(app.Context())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(cols)
	if err != nil {
		serverError(w, err)
		return
	}
}

// GetPost gets a Post from the database
// @summary gets a Post
// @router /posts/{id} [get]
// @tags posts
// @id getPost
// @Param id path integer true "Post ID"
// @produce application/json
// @success 200 {object} model.Post "OK"
// @failure 404 {object} model.Errors "Not found"
// @failure 500 {object} model.Errors "Server error"
func (app App) GetPost(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	post, errors := app.api.GetPost(app.Context(), req.URL.String())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	err := enc.Encode(post)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PostPost adds a new Post to the database
// @summary adds a new Post
// @router /posts [post]
// @tags posts
// @id addPost
// @Param post body model.PostIn true "Add Post"
// @accept application/json
// @produce application/json
// @success 201 {object} model.Post "OK"
// @failure 415 {object} model.Errors "Bad Content-Type"
// @failure 500 {object} model.Errors "Server error"
func (app App) PostPost(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	in := model.PostIn{}
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err)
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	post, errors := app.api.AddPost(app.Context(), in)
	if errors != nil {
		log.Printf("[DEBUG] PostPost AddPost %v\n", errors)
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(post)
	if err != nil {
		serverError(w, err)
		return
	}
}

// PutPost updates a Post in the database
// @summary updates a Post
// @router /posts/{id} [put]
// @tags posts
// @id updatePost
// @Param id path integer true "Post ID"
// @Param post body model.Post true "Update Post"
// @accept application/json
// @produce application/json
// @success 200 {object} model.Post "OK"
// @failure 415 {object} model.Errors "Bad Content-Type"
// @failure 500 {object} model.Errors "Server error"
func (app App) PutPost(w http.ResponseWriter, req *http.Request) {
	mt, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mt != contentType {
		msg := fmt.Sprintf("Bad Content-Type '%s'", mt)
		ErrorResponse(w, http.StatusUnsupportedMediaType, msg)
		return
	}
	var in model.Post
	err = json.NewDecoder(req.Body).Decode(&in)
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err)
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	post, errors := app.api.UpdatePost(app.Context(), req.URL.String(), in)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.Header().Set("Content-Type", contentType)
	enc := json.NewEncoder(w)
	err = enc.Encode(post)
	if err != nil {
		serverError(w, err)
		return
	}
}

// DeletePost deletes a Post from the database
// @summary deletes a Post
// @router /posts/{id} [delete]
// @tags posts
// @id deletePost
// @Param id path integer true "Post ID"
// @success 204 {object} model.Post "OK"
// @failure 500 {object} model.Errors "Server error"
func (app App) DeletePost(w http.ResponseWriter, req *http.Request) {
	errors := app.api.DeletePost(app.Context(), req.URL.String())
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
