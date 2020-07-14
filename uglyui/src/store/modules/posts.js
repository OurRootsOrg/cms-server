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
      for (let i = 0; i < state.postsList.length; i++) {
        if (state.postsList[i].id === post.id) {
          state.postsList[i] = post;
        }
      }
    }
  }
};

export const actions = {
  postsCreate({ commit, dispatch }, post) {
    return Server.postsCreate(post)
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
  postsUpdate({ commit, dispatch }, post) {
    return Server.postsUpdate(post)
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
  postsDelete({ commit, dispatch }, id) {
    return Server.postsDelete(id)
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
  postsGetAll({ commit, dispatch }) {
    return Server.postsGetAll()
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
  postsGetOne({ commit, dispatch }, id) {
    return Server.postsGetOne(id)
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
