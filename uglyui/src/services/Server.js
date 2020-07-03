import axios from "axios";
import { getAuth } from "../plugins/auth0";
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
  async categoriesCreate(category) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.post("/categories", category, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async categoriesUpdate(cat) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.put(`/categories/${cat.id}`, cat, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async categoriesDelete(id) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.delete(`/categories/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async categoriesGetAll() {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.get("/categories", {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async categoriesGetOne(id) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.get(`/categories/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async collectionsCreate(collection) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.post("/collections", collection, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async collectionsUpdate(coll) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.put(`/collections/${coll.id}`, coll, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async collectionsDelete(id) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.delete(`/collections/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async collectionsGetAll() {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.get("/collections", {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async collectionsGetOne(id) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.get(`/collections/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async contentPostRequest(contentType) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.post(
      "/content",
      { contentType },
      {
        headers: {
          Authorization: `Bearer ${token}`
        }
      }
    );
  },
  contentPut(url, contentType, data) {
    return axios.put(url, data, {
      headers: {
        "Content-Type": contentType
      }
    });
  },
  async postsGetAll() {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.get("/posts", {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async postsGetOne(id) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.get(`/posts/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async postsCreate(post) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.post("/posts", post, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async postsUpdate(post) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.put(`/posts/${post.id}`, post, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async postsDelete(id) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.delete(`/posts/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async recordsGetForPost(postId) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.get("/records", {
      params: { post: postId },
      headers: {
        Authorization: `Bearer ${token}`
      }
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
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.get(`/settings`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  },
  async settingsUpdate(post) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.put(`/settings`, post, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  }
};
