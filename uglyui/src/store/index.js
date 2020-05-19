import Vue from "vue";
import Vuex from "vuex";
import * as auth from "./modules/auth.js";
import * as categories from "./modules/categories.js";
import * as collections from "./modules/collections.js";
import * as notifications from "./modules/notifications.js";
import * as posts from "./modules/posts.js";

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    auth,
    categories,
    collections,
    notifications,
    posts
  },
  state: {},
  mutations: {},
  actions: {}
});
