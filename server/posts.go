package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ourrootsorg/cms-server/model"
)

// GetPosts returns all posts in the database
// @summary returns all posts
// @router /posts [get]
// @tags posts
// @id getPosts
// @produce application/json
// @success 200 {array} model.Post "OK"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetPosts(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	cols, errors := app.api.GetPosts(req.Context())
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
// @failure 404 {object} api.Error "Not found"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetPost(w http.ResponseWriter, req *http.Request) {
	postID, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	post, errors := app.api.GetPost(req.Context(), postID)
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

// GetPostImage redirects the client to an image URL
// @summary Returns a redirect to an image URL
// @router /posts/{id}/images/{filePath} [get]
// @tags posts
// @id getPostImage
// @Param id path integer true "Post ID"
// @Param imageFile path string true "Image file path"
// @param noredirect query bool false "return the url as json {url: <url>} if true (optional)"
// @param height query int false "height of image thumbnail (optional)"
// @param width query int false "width of image thumbnail (optional)"
// @success 307 {header} string
// @failure 404 {object} api.Error "Not found"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) GetPostImage(w http.ResponseWriter, req *http.Request) {
	const expireSeconds = 3600
	postID, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	filePath := mux.Vars(req)["filePath"]
	if filePath == "" {
		ErrorResponse(w, http.StatusNotFound, "Not Found")
	}
	noredirect, _ := strconv.ParseBool(req.URL.Query().Get("noredirect"))
	height, err := strconv.Atoi(req.URL.Query().Get("height"))
	if err != nil {
		height = 0
	}
	width, err := strconv.Atoi(req.URL.Query().Get("width"))
	if err != nil {
		width = 0
	}
	imageMetadata, errors := app.api.GetPostImage(req.Context(), postID, filePath, expireSeconds, height, width)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	// necessary because the <img> tag in a browser can't send the auth header
	// and a javascript GET request must always follow redirects, which we don't want
	if noredirect {
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", expireSeconds))
		err := enc.Encode(imageMetadata)
		if err != nil {
			serverError(w, err)
		}
		return
	}
	http.Redirect(w, req, imageMetadata.URL, http.StatusTemporaryRedirect)
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
// @failure 415 {object} api.Error "Bad Content-Type"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
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
	post, errors := app.api.AddPost(req.Context(), in)
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
// @failure 415 {object} api.Error "Bad Content-Type"
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) PutPost(w http.ResponseWriter, req *http.Request) {
	postID, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
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
	if !model.UserAcceptedPostRecordsStatus(in.RecordsStatus) {
		msg := fmt.Sprintf("Invalid records status: %s", in.RecordsStatus)
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	if !model.UserAcceptedPostImagesStatus(in.ImagesStatus) {
		msg := fmt.Sprintf("Invalid records status: %s", in.RecordsStatus)
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}
	post, errors := app.api.UpdatePost(req.Context(), postID, in)
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
// @failure 500 {object} api.Error "Server error"
// @Security OAuth2Implicit[cms,openid,profile,email]
// @Security OAuth2AuthCode[cms,openid,profile,email]
func (app App) DeletePost(w http.ResponseWriter, req *http.Request) {
	postID, errors := getIDFromRequest(req)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	errors = app.api.DeletePost(req.Context(), postID)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
