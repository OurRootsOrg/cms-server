import Server from "@/services/Server";

export const state = {
  usersList: [],
  user: null
};

export const mutations = {
  USERS_SET(state, users) {
    state.usersList = users;
  },
  USERS_REMOVE(state, id) {
    state.usersList = state.usersList.filter(u => u.id !== id);
  },
  USER_SET(state, user) {
    state.user = user;
  }
};

export const actions = {
  usersGetAll({ commit, dispatch, rootGetters }) {
    return Server.usersGetAll(rootGetters.currentSocietyId)
      .then(response => {
        commit("USERS_SET", response.data);
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
  usersUpdate({ commit, dispatch, rootGetters }, user) {
    return Server.usersUpdate(rootGetters.currentSocietyId, user)
      .then(response => {
        commit("USERS_UPDATE", response.data);
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
  usersDelete({ commit, dispatch, rootGetters }, id) {
    return Server.usersDelete(rootGetters.currentSocietyId, id)
      .then(() => {
        commit("USERS_REMOVE", id);
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
  },
  userSet({ commit }, user) {
    commit("USER_SET", user);
  }
};

export const getters = {
  userIsLoggedIn: state => {
    return !!state.user;
  },
  userId: state => {
    return state.user ? state.user.id : 0;
  }
};
