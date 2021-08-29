<template>
  <v-app id="app">
    <v-app-bar :clipped-left="true" app dark color="primary">
      <!--      <v-app-bar-nav-icon @click.stop="drawer = !drawer"></v-app-bar-nav-icon>-->
      <router-link to="/">
        <v-toolbar-title style="color: white; width: 300px" class="ml-0 pl-4">
          <img src="./assets/roots-white.svg" height="25" class="mt-1 mb-n2" />
          <span class="hidden-sm-and-down pl-2">Databases for Genealogy</span>
        </v-toolbar-title>
      </router-link>
      <v-spacer></v-spacer>
      <!--      <v-btn icon>-->
      <!--        <v-icon>mdi-bell</v-icon>-->
      <!--      </v-btn>-->
      <v-menu offset-y :close-on-click="true" v-if="users.user">
        <template v-slot:activator="{ on, attrs }">
          <v-btn icon v-bind="attrs" v-on="on">
            <v-avatar size="32">
              <img v-if="users.user.picture" :src="users.user.picture" />
              <v-icon v-else>mdi-account</v-icon>
            </v-avatar>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="logout()">
            <v-list-item-title>Logout</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>

    <v-main>
      <div id="container-wrapper">
        <Notifications />
        <v-container>
          <v-row class="pa-4 pt-0">
            <router-view id="view" :key="$route.fullPath"></router-view>
          </v-row>
        </v-container>
      </div>
    </v-main>
  </v-app>
</template>

<script>
import Notifications from "@/components/Notifications.vue";
import NProgress from "nprogress";
import Auth from "@/services/Auth";
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
  computed: mapState(["users"]),
  data: () => ({
    dialog: false
  }),
  methods: {
    isLoggedIn() {
      return store.getters.userIsLoggedIn;
    },
    logout() {
      Auth.logout();
    }
  }
};
</script>

<style>
.rowHover {
  cursor: pointer;
}
.v-data-table--dense .v-data-table-header {
  background: #f1f1f1;
  padding-top: 5px;
}
.columns-resize-bar {
  border-left: solid 1px #ccc;
  height: 100px;
  max-height: 31px;
}
.errorMessage {
  color: red;
}
.no-underline a {
  text-decoration: none;
}
.text-first-caps {
  text-transform: capitalize;
}
.smallCheckbox .v-checkbox {
  color: #0097a7 !important;
}

.smallCheckbox i {
  font-size: 17px !important;
  color: #0097a7 !important;
  margin-top: -3px;
}
</style>
