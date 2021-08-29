import Server from "@/services/Server.js";

export const state = {
  postsList: [],
  post: {}
};

export const mutations = {
  POSTS_SET(state, posts) {
    state.postsList = posts;
  },
  POSTS_ADD(state, post) {
    state.postsList.push(post);
  },
  POSTS_REMOVE(state, id) {
    state.postsList = state.postsList.filter(post => post.id !== id);
  },
  POST_SET(state, post) {
    state.post = post;
  },
  POST_UPDATE(state, post) {
    if (state.post && state.post.id === post.id) {
      state.post = post;
    }
    if (state.postsList) {
      state.postsList = state.postsList.map(p => (p.id === post.id ? post : p));
    }
  }
};

export const actions = {
  postsCreate({ commit, dispatch, rootGetters }, post) {
    return Server.postsCreate(rootGetters.currentSocietyId, post)
      .then(response => {
        commit("POSTS_ADD", response.data);
        const notification = {
          type: "success",
          message: "Your post has been created"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem creating your post: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  postsUpdate({ commit, dispatch, rootGetters }, post) {
    return Server.postsUpdate(rootGetters.currentSocietyId, post)
      .then(response => {
        commit("POST_UPDATE", response.data);
        const notification = {
          type: "success",
          message: "Your post has been updated"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem updating your post: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  postsDelete({ commit, dispatch, rootGetters }, id) {
    return Server.postsDelete(rootGetters.currentSocietyId, id)
      .then(() => {
        commit("POSTS_REMOVE", id);
        const notification = {
          type: "success",
          message: "Your post has been deleted"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem deleting the post: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  postsGetAll({ commit, dispatch, rootGetters }) {
    return Server.postsGetAll(rootGetters.currentSocietyId)
      .then(response => {
        commit("POSTS_SET", response.data.posts);
        return response.data.posts;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading posts: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  postsGetOne({ commit, dispatch, rootGetters }, id) {
    return Server.postsGetOne(rootGetters.currentSocietyId, id)
      .then(response => {
        commit("POST_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading post: " + id + " " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};
