<template>
  <v-container class="society">
    <v-navigation-drawer
      v-model="drawer"
      :clipped="true"
      :mini-variant="$vuetify.breakpoint.mdAndDown"
      :permanent="drawer"
      app
    >
      <v-list dense>
        <template v-for="item in items">
          <v-row v-if="!itemAuthorized(item)" :key="item.heading + item.text"></v-row>
          <v-row v-else-if="item.heading" :key="item.heading" align="center">
            <v-col cols="6">
              <v-subheader v-if="item.heading">{{ item.heading }}</v-subheader>
            </v-col>
            <v-col cols="6" class="text-center">
              <a href="#!" class="body-2 black--text">EDIT</a>
            </v-col>
          </v-row>
          <v-list-group
            v-else-if="item.children"
            :key="item.text"
            v-model="item.model"
            :append-icon="item.model ? item.post_icon : item['post_icon-alt']"
          >
            <template v-slot:activator>
              <v-list-item-action>
                <v-icon :title="item.text">{{ item.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title>{{ item.text }}</v-list-item-title>
              </v-list-item-content>
            </template>
            <v-list-item v-for="(child, i) in item.children" :key="i" :to="{ name: child.link }" link>
              <v-list-item-action v-if="child.icon">
                <v-icon :title="child.text">{{ child.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title>{{ child.text }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list-group>
          <v-list-item v-else-if="item.external" :key="item.text" :href="item.link" target="_blank">
            <v-list-item-action>
              <v-icon :title="item.text">{{ item.icon }}</v-icon>
            </v-list-item-action>
            <v-list-item-content>
              <v-list-item-title>{{ item.text }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item v-else :key="item.text" :to="{ name: item.link }" link>
            <v-list-item-action>
              <v-icon :title="item.text">{{ item.icon }}</v-icon>
            </v-list-item-action>
            <v-list-item-content>
              <v-list-item-title>{{ item.text }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>
      </v-list>
    </v-navigation-drawer>

    <router-view id="society-view" :key="$route.fullPath"></router-view>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";

const AUTH_LEVEL_READER = 1;
const AUTH_LEVEL_ADMIN = 4;

function getContent(societyId, next) {
  Promise.all([
    store.dispatch("societySummariesGetOne", societyId),
    store.dispatch("societyUsersGetCurrent", societyId)
  ])
    .then(() => {
      next();
    })
    .catch(() => {
      next("/");
    });
}

export default {
  name: "Society",
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    console.log("society.beforeRouteEnter");
    getContent(routeTo.params.society, next);
  },
  beforeRouteUpdate: function(routeTo, routeFrom, next) {
    console.log("society.beforeRouteUpdate");
    getContent(routeTo.params.society, next);
  },
  created() {
    this.items[0].text = this.societySummaries.societySummary.name;
  },
  data: () => ({
    drawer: true,
    items: [
      { icon: "mdi-home", text: "", link: "society-home", authLevel: AUTH_LEVEL_READER },
      { icon: "mdi-shape", text: "Categories", link: "categories-list", authLevel: AUTH_LEVEL_READER },
      { icon: "mdi-book-open-variant", text: "Collections", link: "collections-list", authLevel: AUTH_LEVEL_READER },
      { icon: "mdi-cloud-upload", text: "Record sets", link: "posts-list", authLevel: AUTH_LEVEL_READER },
      { icon: "mdi-account-circle", text: "Users", link: "users-list", authLevel: AUTH_LEVEL_ADMIN },
      // { icon: "mdi-open-in-new", text: "Search", link: process.env.VUE_APP_SEARCH_URL, external: true },
      { icon: "mdi-cog", text: "Settings", link: "settings", authLevel: AUTH_LEVEL_ADMIN }
    ]
  }),
  computed: mapState(["societySummaries"]),
  methods: {
    itemAuthorized(item) {
      return item.authLevel <= store.getters.authLevel;
    }
  }
};
</script>
