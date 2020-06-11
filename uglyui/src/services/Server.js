import axios from "axios";
import { getAuth } from "../auth";

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
  async categoriesCreate(category) {
    let auth = await getAuth();
    let token = await auth.getTokenSilently();
    return apiClient.post("/categories", category, {
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
    return apiClient.put(coll.id, coll, {
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
    return apiClient.get(id, {
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
      });
  },
  contentPut(url, contentType, data) {
    return axios.put(url, data, {
      headers: {
        "Content-Type": contentType
      },
      timeout: 10000
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
    return apiClient.get(id, {
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
    return apiClient.put(post.id, post, {
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
  }
};
