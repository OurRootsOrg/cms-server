<template>
  <v-app id="app">
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

export default {
  components: {
    Notifications,
  },
  mounted() {
    console.log("Search mounted");
    NProgress.configure({ parent: "#container-wrapper" });
  },
  props: {
    source: String,
  },
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
  color: #607d8b !important;
}

.smallCheckbox i {
  font-size: 17px !important;
  color: #607d8b !important;
  margin-top: -3px;
}
</style>
