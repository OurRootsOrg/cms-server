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
  currentUser() {
    return get(`/currentuser`);
  },
  placeSearch(prefix) {
    return get(`places`, {
      params: { prefix: prefix, count: 8 }
    });
  },
  postsGetAll() {
    return get("/posts");
  },
  postsGetOne(id) {
    return get(`/posts/${id}`);
  },
  postsGetImage(postId, imagePath, thumbnail) {
    let url = `/posts/${postId}/images/${imagePath}?noredirect=true`;
    if (thumbnail) {
      url += `&thumbnail=${thumbnail}`;
    }
    return get(url);
  },
  postsCreate(pst) {
    return post("/posts", pst);
  },
  postsUpdate(pst) {
    return put(`/posts/${pst.id}`, pst);
  },
  postsDelete(id) {
    return del(`/posts/${id}`);
  },
  recordsGetDetail(id) {
    return get(`/records/${id}?details=true`);
  },
  recordsGetForPost(postId) {
    return get("/records", {
      params: { post: postId }
    });
  },
  settingsGet() {
    return get(`/settings`);
  },
  settingsUpdate(post) {
    return put(`/settings`, post);
  }
};
