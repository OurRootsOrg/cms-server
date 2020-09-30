import Vue from "vue";
import Vuex from "vuex";
import * as categories from "./modules/categories.js";
import * as collections from "./modules/collections.js";
import * as notifications from "./modules/notifications.js";
import * as posts from "./modules/posts.js";
import * as records from "./modules/records.js";
import * as settings from "./modules/settings.js";
import * as user from "./modules/user.js";

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    categories,
    collections,
    notifications,
    posts,
    records,
    settings,
    user
  },
  state: {},
  mutations: {},
  actions: {},
  getters: {}
});
