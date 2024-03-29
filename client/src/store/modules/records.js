import Server from "@/services/Server.js";

export const state = {
  recordsList: [],
  record: null
};

export const mutations = {
  RECORDS_SET(state, records) {
    state.recordsList = records;
  },
  RECORD_SET(state, record) {
    state.record = record;
  }
};

export const actions = {
  recordsGetForPost({ commit, dispatch, rootGetters }, postId) {
    return Server.recordsGetForPost(rootGetters.currentSocietyId, postId)
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
  },
  recordsGetDetail({ commit, dispatch, rootGetters }, recordId) {
    return Server.recordsGetDetail(rootGetters.currentSocietyId, recordId)
      .then(response => {
        commit("RECORD_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading the record: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};
