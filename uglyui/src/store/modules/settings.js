import Server from "@/services/Server.js";

export const state = {
  settings: {}
};

export const mutations = {
  SETTINGS_SET(state, settings) {
    state.settings = settings;
  }
};

export const actions = {
  settingsUpdate({ commit, dispatch }, settings) {
    return Server.settingsUpdate(settings)
      .then(response => {
        commit("SETTINGS_SET", response.data);
        const notification = {
          type: "success",
          message: "Your settings have been updated!"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem updating your settings: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  settingsGet({ commit, dispatch }) {
    return Server.settingsGet()
      .then(response => {
        commit("SETTINGS_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading settings: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};
