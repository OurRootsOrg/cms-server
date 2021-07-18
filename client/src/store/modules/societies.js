import Server from "@/services/Server.js";

export const state = {
  society: {}
};

export const mutations = {
  SOCIETY_SET(state, society) {
    state.society = society;
  }
};

export const actions = {
  societiesCreate({ commit, dispatch }, society) {
    return Server.societiesCreate(society)
      .then(response => {
        commit("SOCIETY_SET", response.data);
        const notification = {
          type: "success",
          message: "Your society has been created"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem creating your society: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  societiesUpdate({ commit, dispatch }, society) {
    return Server.societiesUpdate(society)
      .then(response => {
        commit("SOCIETY_SET", response.data);
        const notification = {
          type: "success",
          message: "Your society has been updated"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem updating your society: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  societiesGetCurrent({ commit, dispatch, rootGetters }) {
    return Server.societiesGetOne(rootGetters.currentSocietyId)
      .then(response => {
        if (!response.data.postMetadata) {
          response.data.postMetadata = [];
        }
        commit("SOCIETY_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading the society: " + rootGetters.currentSocietyId + " " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};
