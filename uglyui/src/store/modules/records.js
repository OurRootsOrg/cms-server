import Server from "@/services/Server.js";

export const state = {
  recordsList: []
};

export const mutations = {
  RECORDS_SET(state, records) {
    state.recordsList = records;
  }
};

export const actions = {
  recordsGetForPost({ commit, dispatch }, postId) {
    return Server.recordsGetForPost(postId)
      .then(response => {
        commit("RECORDS_SET", response.data.records);
        return response.data.records;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading records: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};
