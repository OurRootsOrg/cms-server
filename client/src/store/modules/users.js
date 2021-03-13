export const state = {
  user: null
};

export const mutations = {
  USER_SET(state, user) {
    state.user = user;
  }
};

export const actions = {
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
