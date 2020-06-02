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
          message: "Your post has been created!"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
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
          message: "Your post has been updated!"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem updating your post: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  },
  postsGetAll({ commit, dispatch }) {
    return Server.postsGetAll()
      .then(response => {
        commit("POSTS_SET", response.data.posts);
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem fetching posts: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  },
  postsGetOne({ commit, getters, state }, id) {
    if (id === state.post.id) {
      return state.post;
    }

    let post = getters.getPostById(id);

    if (post) {
      commit("POST_SET", post);
      return post;
    } else {
      return Server.postsGetOne(id).then(response => {
        commit("POST_SET", response.data);
        return response.data;
      });
    }
  }
};
export const getters = {
  getPostById: state => id => {
    return state.postsList ? state.postsList.find(post => post.id === id) : null;
  }
};
