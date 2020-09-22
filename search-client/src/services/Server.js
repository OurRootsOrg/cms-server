import { get } from "./ServerHelper";

export default {
  placeSearch(prefix) {
    console.log("server.placeSearch", prefix);
    return get(`places`, {
      params: { prefix: prefix, count: 8 }
    });
  },
  postsGetImage(postId, imagePath, thumbnail) {
    let url = `/posts/${postId}/images/${imagePath}?noredirect=true`;
    if (thumbnail) {
      url += `&thumbnail=${thumbnail}`;
    }
    return get(url);
  },
  search(query) {
    console.log("server.search", query);
    return get("/search", {
      params: query
    });
  },
  searchGetResult(id) {
    console.log("server.searchGetResult", id);
    return get(`/search/${id}`);
  }
};
