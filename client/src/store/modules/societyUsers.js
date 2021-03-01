import Server from "@/services/Server.js";

export const state = {
  societyUserCurrent: {}
};

export const mutations = {
  SOCIETY_USER_CURRENT_SET(state, societyUser) {
    state.societyUserCurrent = societyUser;
  }
};

export const actions = {
  societyUsersGetCurrent({ commit, dispatch }, societyId) {
    return Server.societyUsersGetCurrent(societyId)
      .then(response => {
        commit("SOCIETY_USER_CURRENT_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem accessing this society: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};

export const getters = {
  authLevel: state => {
    return state.societyUserCurrent ? state.societyUserCurrent.level : 0;
  }
};
