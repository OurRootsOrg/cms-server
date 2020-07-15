import axios from "axios";
import { get, post, put, del } from "./ServerHelper";

export default {
  categoriesCreate(category) {
    return post("/categories", category);
  },
  categoriesUpdate(cat) {
    return put(`/categories/${cat.id}`);
  },
  categoriesDelete(id) {
    return del(`/categories/${id}`);
  },
  categoriesGetAll() {
    return get("/categories");
  },
  categoriesGetOne(id) {
    return get(`/categories/${id}`);
  },
  collectionsCreate(collection) {
    return post("/collections", collection);
  },
  collectionsUpdate(coll) {
    return put(`/collections/${coll.id}`, coll);
  },
  collectionsDelete(id) {
    return del(`/collections/${id}`);
  },
  collectionsGetAll() {
    return get("/collections");
  },
  collectionsGetOne(id) {
    return get(`/collections/${id}`);
  },
  contentPostRequest(contentType) {
    return post("/content", { contentType });
  },
  contentPut(url, contentType, data) {
    return axios.put(url, data, {
      headers: {
        "Content-Type": contentType
      }
    });
  },
  postsGetAll() {
    return get("/posts");
  },
  postsGetOne(id) {
    return get(`/posts/${id}`);
  },
  postsCreate(post) {
    return post("/posts", post);
  },
  postsUpdate(post) {
    return put(`/posts/${post.id}`, post);
  },
  postsDelete(id) {
    return del(`/posts/${id}`);
  },
  recordsGetForPost(postId) {
    return get("/records", {
      params: { post: postId }
    });
  },
  search(query) {
    return get("/search", {
      params: query
    });
  },
  searchGetResult(id) {
    return get(`/search/${id}`);
  },
  settingsGet() {
    return get(`/settings`);
  },
  settingsUpdate(post) {
    return put(`/settings`, post);
  }
};
