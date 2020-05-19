import axios from "axios";

const apiClient = axios.create({
  baseURL: `http://localhost:8000`,
  withCredentials: false, // This is the default
  headers: {
    Accept: "application/json",
    "Content-Type": "application/json"
  },
  timeout: 10000
});

export default {
  authRegister(user) {
    return apiClient.post("/auth/register", user);
  },
  categoriesCreate(category) {
    return apiClient.post("/categories", category);
  },
  categoriesGetAll() {
    return apiClient.get("/categories");
  },
  collectionsCreate(collection) {
    return apiClient.post("/collections", collection);
  },
  collectionsGetAll() {
    return apiClient.get("/collections");
  },
  contentPostRequest(contentType) {
    return apiClient.post("/content", { contentType });
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
    return apiClient.get("/posts");
  },
  postsCreate(post) {
    return apiClient.post("/posts", post);
  }
};
