import { get } from "./ServerHelper";

export default {
  placeSearch(prefix) {
    return get(`places`, {
      params: { prefix: prefix, count: 8 }
    });
  },
  postsGetImage(societyId, postId, imagePath, thumbnail) {
    let url = `/search-image/${societyId}/${postId}/${imagePath}`;
    if (thumbnail) {
      url += `?thumbnail=${thumbnail}`;
    }
    return get(url);
  },
  search(query) {
    return get("/search", {
      params: query
    });
  },
  searchGetResult(id) {
    return get(`/search/${id}`);
  }
};
