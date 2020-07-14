import Server from "@/services/Server.js";

export const state = {
  searchList: [],
  searchTotal: 0,
  searchResult: {}
};

export const mutations = {
  SEARCH_SET(state, search) {
    state.searchList = search["hits"];
    state.searchTotal = search["total"];
  },
  SEARCH_RESULT_SET(state, result) {
    state.searchResult = result;
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
          error,
          type: "error",
          message: "There was a problem during search: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  searchGetResult({ commit, dispatch }, id) {
    return Server.searchGetResult(id)
      .then(response => {
        commit("SEARCH_RESULT_SET", response.data);
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem getting the search result: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};
