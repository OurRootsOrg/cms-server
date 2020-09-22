export const state = {
  notificationsList: []
};

let nextId = 1;

export const mutations = {
  NOTIFICATIONS_PUSH(state, notification) {
    state.notificationsList.push({
      ...notification,
      id: nextId++
    });
  },
  NOTIFICATIONS_DELETE(state, id) {
    state.notificationsList = state.notificationsList.filter(notification => notification.id !== id);
  }
};
export const actions = {
  notificationsAdd({ commit }, notification) {
    commit("NOTIFICATIONS_PUSH", notification);
  },
  notificationsRemove({ commit }, id) {
    commit("NOTIFICATIONS_DELETE", id);
  }
};
