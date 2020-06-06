import Server from "@/services/Server.js";

export const state = {
  searchList: [],
  searchTotal: 0
};

export const mutations = {
  SEARCH_SET(state, search) {
    state.searchList = search["hits"];
    state.searchTotal = search["total"];
  }
};

export const actions = {
  search({ commit, dispatch }, query) {
    return Server.search(query)
      .then(response => {
        commit("SEARCH_SET", response.data);
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem during search: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  }
};
