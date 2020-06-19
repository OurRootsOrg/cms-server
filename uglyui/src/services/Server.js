import axios from "axios";
import { getInstance } from "../auth";

const apiClient = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL,
  withCredentials: false, // This is the default
  headers: {
    Accept: "application/json",
    "Content-Type": "application/json"
  },
  timeout: 10000
});

export default {
  categoriesCreate(category) {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.post("/categories", category, {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
      });
  },
  categoriesGetAll() {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.get("/categories", {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
      });
  },
  collectionsCreate(collection) {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.post("/collections", collection, {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
      });
  },
  collectionsGetAll() {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.get("/collections", {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
      });
  },
  contentPostRequest(contentType) {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.post(
          "/content",
          { contentType },
          {
            headers: {
              Authorization: `Bearer ${token}`
            }
          }
        );
      });
  },
  contentPut(url, contentType, data) {
    console.log("contentPut", url, contentType, data);
    return axios.put(url, data, {
      headers: {
        "Content-Type": contentType
      },
      timeout: 10000
    });
  },
  postsGetAll() {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.get("/posts", {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
      });
  },
  postsGetOne(id) {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.get(`/posts/${id}`, {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
      });
  },
  postsCreate(post) {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.post("/posts", post, {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
      });
  },
  postsUpdate(post) {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.put(`/posts/${post.id}`, post, {
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
      });
  },
  recordsGetForPost(postId) {
    return getInstance()
      .getTokenSilently()
      .then(token => {
        return apiClient.get("/records", {
          params: { post: postId },
          headers: {
            Authorization: `Bearer ${token}`
          }
        });
      });
  },
  search(query) {
    return apiClient.get("/search", {
      params: query
    });
  }
};
