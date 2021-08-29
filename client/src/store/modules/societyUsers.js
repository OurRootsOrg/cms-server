import Server from "@/services/Server.js";

export const state = {
  societyUserCurrent: {},
  societyUsersList: []
};

export const mutations = {
  SOCIETY_USER_CURRENT_SET(state, societyUser) {
    state.societyUserCurrent = societyUser;
  },
  SOCIETY_USERS_SET(state, societyUsers) {
    state.societyUsersList = societyUsers;
  },
  SOCIETY_USERS_UPDATE(state, societyUser) {
    if (state.societyUserCurrent && state.societyUserCurrent.id === societyUser.id) {
      state.societyUserCurrent = societyUser;
    }
    if (state.societyUsersList) {
      state.societyUsersList = state.societyUsersList.map(su => (su.id === societyUser.id ? societyUser : su));
    }
  },
  SOCIETY_USERS_REMOVE(state, id) {
    state.societyUsersList = state.societyUsersList.filter(su => su.id !== id);
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
  },
  societyUsersGetAll({ commit, dispatch, rootGetters }) {
    return Server.societyUsersGetAll(rootGetters.currentSocietyId)
      .then(response => {
        commit("SOCIETY_USERS_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading users: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  societyUsersUpdate({ commit, dispatch, rootGetters }, societyUser) {
    return Server.societyUsersUpdate(rootGetters.currentSocietyId, societyUser)
      .then(response => {
        commit("SOCIETY_USERS_UPDATE", response.data);
        const notification = {
          type: "success",
          message: "The user has been updated"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem updating the user: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  societyUsersDelete({ commit, dispatch, rootGetters }, id) {
    return Server.societyUsersDelete(rootGetters.currentSocietyId, id)
      .then(() => {
        commit("SOCIETY_USERS_REMOVE", id);
        const notification = {
          type: "success",
          message: "The user has been removed from this society"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem removing the user: " + error.message
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
