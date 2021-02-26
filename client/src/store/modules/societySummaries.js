import Server from "@/services/Server.js";

export const state = {
  societySummariesList: [],
  societySummary: {}
};

export const mutations = {
  SOCIETY_SUMMARIES_SET(state, societySummaries) {
    state.societySummariesList = societySummaries;
  },
  SOCIETY_SUMMARY_SET(state, societySummary) {
    state.societySummary = societySummary;
  }
};

export const actions = {
  societySummariesGetAll({ commit, dispatch }) {
    return Server.societySummariesGetAll()
      .then(response => {
        console.log("societySummaries.GetAll", response);
        commit("SOCIETY_SUMMARIES_SET", response.data);
        return response.data.societySummaries;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading societies for user: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  societySummariesGetOne({ commit, dispatch }, societyId) {
    return Server.societySummariesGetOne(societyId)
      .then(response => {
        console.log("societySummaries.GetOne", response);
        commit("SOCIETY_SUMMARY_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading society: " + societyId + " " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};

export const getters = {
  currentSocietyId: state => {
    return state.societySummary.id;
  }
};
