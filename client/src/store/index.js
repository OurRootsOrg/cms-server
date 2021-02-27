import Vue from "vue";
import Vuex from "vuex";
import * as categories from "./modules/categories.js";
import * as collections from "./modules/collections.js";
import * as notifications from "./modules/notifications.js";
import * as posts from "./modules/posts.js";
import * as records from "./modules/records.js";
import * as users from "./modules/users.js";
import * as societySummaries from "./modules/societySummaries.js";
import * as societies from "./modules/societies.js";
import * as societyUsers from "./modules/societyUsers.js";
import * as invitations from "./modules/invitations.js";

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    categories,
    collections,
    notifications,
    posts,
    records,
    users,
    societySummaries,
    societies,
    societyUsers,
    invitations
  },
  state: {},
  mutations: {},
  actions: {},
  getters: {}
});
