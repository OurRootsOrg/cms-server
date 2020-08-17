<template>
  <v-app id="app">
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
            <v-list-item v-for="(child, i) in item.children" :key="i" :to="child.link" link>
              <v-list-item-action v-if="child.icon">
                <v-icon :title="child.text">{{ child.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title>{{ child.text }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list-group>
          <v-list-item v-else :key="item.text" :to="item.link" link>
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

    <v-app-bar :clipped-left="true" app dark color="primary">
      <v-app-bar-nav-icon @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title style="width: 300px" class="ml-0 pl-4">
        <img src="./assets/roots-white.svg" height="25" class="mt-1 mb-n2" />
        <span class="hidden-sm-and-down pl-2">OurRoots CMS Sandbox</span>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn icon>
        <v-icon>mdi-bell</v-icon>
      </v-btn>
      <v-btn icon>
        <v-avatar v-if="user.user && user.user.picture" size="28"><img :src="user.user.picture"/></v-avatar>
      </v-btn>
    </v-app-bar>

    <v-main>
      <div id="container-wrapper">
        <Notifications />
        <v-container>
          <v-row class="pa-4">
            <router-view id="view" :key="$route.fullPath"></router-view>
          </v-row>
        </v-container>
      </div>
    </v-main>

    <!--FAB commented out for now
    <v-btn bottom color="pink" dark fab fixed right @click="dialog = !dialog">
      <v-icon>mdi-plus</v-icon>
    </v-btn>-->
  </v-app>
</template>

<script>
import Notifications from "@/components/Notifications.vue";
import NProgress from "nprogress";
import store from "@/store";
import { mapState } from "vuex";

export default {
  components: {
    Notifications
  },
  mounted() {
    NProgress.configure({ parent: "#container-wrapper" });
  },
  props: {
    source: String
  },
  computed: mapState(["user"]),
  data: () => ({
    dialog: false,
    drawer: true,
    items: [
      { icon: "mdi-home", text: "Home", link: "/" },
      { icon: "mdi-chart-areaspline", text: "Dashboard", link: "/dashboard", authRequired: true },
      { icon: "mdi-shape", text: "Categories", link: "/categories", authRequired: true },
      { icon: "mdi-book-open-variant", text: "Collections", link: "/collections", authRequired: true },
      { icon: "mdi-cloud-upload", text: "Posts", link: "/posts", authRequired: true },
      { icon: "mdi-account-circle", text: "Users", link: "/users", authRequired: true },
      { icon: "mdi-open-in-new", text: "Search", link: "/search" },
      { icon: "mdi-cog", text: "Settings", link: "/settings", authRequired: true }
    ]
  }),
  methods: {
    itemAuthorized(item) {
      return !item.authRequired || store.getters.userIsLoggedIn;
    }
  }
};
</script>

<style>
.rowHover {
  cursor: pointer;
}

</style>