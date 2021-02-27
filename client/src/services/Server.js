import axios from "axios";
import {get, post, put, del, getWithoutAccessToken} from "./ServerHelper";

export default {
  categoriesCreate(societyId, category) {
    return post(`/societies/${societyId}/categories`, category);
  },
  categoriesUpdate(societyId, cat) {
    return put(`/societies/${societyId}/categories/${cat.id}`);
  },
  categoriesDelete(societyId, id) {
    return del(`/societies/${societyId}/categories/${id}`);
  },
  categoriesGetAll(societyId) {
    return get(`/societies/${societyId}/categories`);
  },
  categoriesGetOne(societyId, id) {
    return get(`/societies/${societyId}/categories/${id}`);
  },
  collectionsCreate(societyId, collection) {
    return post(`/societies/${societyId}/collections`, collection);
  },
  collectionsUpdate(societyId, coll) {
    return put(`/societies/${societyId}/collections/${coll.id}`, coll);
  },
  collectionsDelete(societyId, id) {
    return del(`/societies/${societyId}/collections/${id}`);
  },
  collectionsGetAll(societyId) {
    return get(`/societies/${societyId}/collections`);
  },
  collectionsGetOne(societyId, id) {
    return get(`/societies/${societyId}/collections/${id}`);
  },
  contentPostRequest(societyId, contentType) {
    return post(`/societies/${societyId}/content`, { contentType });
  },
  contentPut(url, contentType, data) {
    return axios.put(url, data, {
      headers: {
        "Content-Type": contentType
      }
    });
  },
  currentUser() {
    return get(`/current_user`);
  },
  invitationsCreate(societyId, invitation) {
    return post(`/societies/${societyId}/invitations`, invitation);
  },
  invitationsDelete(societyId, id) {
    return del(`/societies/${societyId}/invitations/${id}`);
  },
  invitationsGetAll(societyId) {
    return get(`/societies/${societyId}/invitations`);
  },
  invitationGetForCode(code) {
    return getWithoutAccessToken(`/invitations/${code}`);
  },
  invitationAccept(code) {
    return post(`/invitations/${code}`, {});
  },
  placeSearch(prefix) {
    return get(`places`, {
      params: { prefix: prefix, count: 8 }
    });
  },
  postsGetAll(societyId) {
    return get(`/societies/${societyId}/posts`);
  },
  postsGetOne(societyId, id) {
    return get(`/societies/${societyId}/posts/${id}`);
  },
  postsGetImage(societyId, postId, imagePath, thumbnail) {
    let url = `/societies/${societyId}/posts/${postId}/images/${imagePath}?noredirect=true`;
    if (thumbnail) {
      url += `&thumbnail=${thumbnail}`;
    }
    return get(url);
  },
  postsCreate(societyId, pst) {
    return post(`/societies/${societyId}/posts`, pst);
  },
  postsUpdate(societyId, pst) {
    return put(`/societies/${societyId}/posts/${pst.id}`, pst);
  },
  postsDelete(societyId, id) {
    return del(`/societies/${societyId}/posts/${id}`);
  },
  recordsGetDetail(societyId, id) {
    return get(`/societies/${societyId}/records/${id}?details=true`);
  },
  recordsGetForPost(societyId, postId) {
    return get(`/societies/${societyId}/records`, {
      params: { post: postId }
    });
  },
  societySummariesGetAll() {
    return get(`/society_summaries`);
  },
  societySummariesGetOne(societyId) {
    return get(`/society_summaries/${societyId}`);
  },
  societiesCreate(society) {
    return post(`/societies`, society);
  },
  societyUsersGetCurrent(societyId) {
    return get(`/societies/${societyId}/current_user`);
  },
  societiesGetOne(societyId) {
    return get(`/societies/${societyId}`);
  },
  societiesUpdate(society) {
    return put(`/societies/${society.id}`, society);
  },
  usersGetAll(societyId) {
    return get(`/societies/${societyId}/users`);
  },
  usersUpdate(societyId, user) {
    return put(`/societies/${societyId}/users/${user.id}`, user);
  },
  usersDelete(societyId, id) {
    return del(`/societies/${societyId}/users/${id}`);
  }
};
