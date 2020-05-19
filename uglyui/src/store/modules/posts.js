import Server from "@/services/Server.js";

export const state = {
  postsList: []
};

export const mutations = {
  POSTS_ADD(state, post) {
    state.postsList.push(post);
  },
  POSTS_SET(state, posts) {
    state.postsList = posts;
  }
};

export const actions = {
  postsCreate({ commit, dispatch }, post) {
    console.log("postsCreate", post);
    return Server.postsCreate(post)
      .then(post => {
        commit("POSTS_ADD", post);
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
  }
};
