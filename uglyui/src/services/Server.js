import axios from "axios";
import axiosRetry from "axios-retry";

const apiClient = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL,
  withCredentials: false, // This is the default
  headers: {
    Accept: "application/json",
    "Content-Type": "application/json"
  }
});
axiosRetry(apiClient, { retries: 3 }); // retry non-POST requests on network or 5XX errors

export default {
  login(token) {
    apiClient.defaults.headers.common["Authorization"] = `Bearer ${token}`;
  },
  isLoggedIn() {
    return !!apiClient.defaults.headers.common["Authorization"];
  },
  categoriesCreate(category) {
    return apiClient.post("/categories", category);
  },
  async categoriesUpdate(cat) {
    return apiClient.put(`/categories/${cat.id}`);
  },
  categoriesDelete(id) {
    return apiClient.delete(`/categories/${id}`);
  },
  categoriesGetAll() {
    return apiClient.get("/categories");
  },
  categoriesGetOne(id) {
    return apiClient.get(`/categories/${id}`);
  },
  async collectionsCreate(collection) {
    return apiClient.post("/collections", collection);
  },
  async collectionsUpdate(coll) {
    return apiClient.put(`/collections/${coll.id}`, coll);
  },
  async collectionsDelete(id) {
    return apiClient.delete(`/collections/${id}`);
  },
  collectionsGetAll() {
    return apiClient.get("/collections");
  },
  async collectionsGetOne(id) {
    return apiClient.get(`/collections/${id}`);
  },
  async contentPostRequest(contentType) {
    return apiClient.post("/content", { contentType });
  },
  contentPut(url, contentType, data) {
    return axios.put(url, data, {
      headers: {
        "Content-Type": contentType
      }
    });
  },
  async postsGetAll() {
    return apiClient.get("/posts");
  },
  async postsGetOne(id) {
    return apiClient.get(`/posts/${id}`);
  },
  async postsCreate(post) {
    return apiClient.post("/posts", post);
  },
  async postsUpdate(post) {
    return apiClient.put(`/posts/${post.id}`, post);
  },
  async postsDelete(id) {
    return apiClient.delete(`/posts/${id}`);
  },
  async recordsGetForPost(postId) {
    return apiClient.get("/records", {
      params: { post: postId }
    });
  },
  search(query) {
    return apiClient.get("/search", {
      params: query
    });
  },
  searchGetResult(id) {
    return apiClient.get(`/search/${id}`);
  },
  async settingsGet() {
    return apiClient.get(`/settings`);
  },
  async settingsUpdate(post) {
    return apiClient.put(`/settings`, post);
  }
};
