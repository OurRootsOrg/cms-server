import Vue from "vue";
import Vuex from "vuex";
import * as notifications from "./modules/notifications.js";
import * as search from "./modules/search.js";

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    notifications,
    search
  },
  state: {},
  mutations: {},
  actions: {},
  getters: {}
});
