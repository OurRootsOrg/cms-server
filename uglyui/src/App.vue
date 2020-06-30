<template>
  <div id="app">
    <div id="nav">
      <div>
        <h3>Home</h3>
        <ul>
          <li><router-link to="/">Go</router-link></li>
        </ul>
      </div>
      <div>
        <h3>Dashboard</h3>
        <ul>
          <li><router-link to="/dashboard">Go</router-link></li>
        </ul>
      </div>
      <div>
        <h3>Categories</h3>
        <ul>
          <li><router-link to="/categories">List</router-link></li>
          <li><router-link to="/categories/create">Create</router-link></li>
        </ul>
      </div>
      <div>
        <h3>Collections</h3>
        <ul>
          <li><router-link to="/collections">List</router-link></li>
          <li><router-link to="/collections/create">Create</router-link></li>
        </ul>
      </div>
      <div>
        <h3>Posts</h3>
        <ul>
          <li><router-link to="/posts">List</router-link></li>
          <li><router-link to="/posts/create">Create</router-link></li>
        </ul>
      </div>
      <div>
        <h3>Users</h3>
        <ul>
          <li><router-link to="/users">List</router-link></li>
        </ul>
      </div>
      <div>
        <h3>Search</h3>
        <ul>
          <li><router-link to="/search">Go</router-link></li>
        </ul>
      </div>
      <div>
        <h3>Settings</h3>
        <ul>
          <li><router-link to="/settings">Go</router-link></li>
        </ul>
      </div>
    </div>
    <Notifications />
    <router-view id="view" :key="$route.fullPath"></router-view>
  </div>
  <!--  <v-app>-->
  <!--    <v-app-bar app color="primary" dark>-->
  <!--      <div class="d-flex align-center">-->
  <!--        <v-img-->
  <!--          alt="Vuetify Logo"-->
  <!--          class="shrink mr-2"-->
  <!--          contain-->
  <!--          src="https://cdn.vuetifyjs.com/images/logos/vuetify-logo-dark.png"-->
  <!--          transition="scale-transition"-->
  <!--          width="40"-->
  <!--        />-->

  <!--        <v-img-->
  <!--          alt="Vuetify Name"-->
  <!--          class="shrink mt-1 hidden-sm-and-down"-->
  <!--          contain-->
  <!--          min-width="100"-->
  <!--          src="https://cdn.vuetifyjs.com/images/logos/vuetify-name-dark.png"-->
  <!--          width="100"-->
  <!--        />-->
  <!--      </div>-->

  <!--      <v-spacer></v-spacer>-->

  <!--      <v-btn href="https://github.com/vuetifyjs/vuetify/releases/latest" target="_blank" text>-->
  <!--        <span class="mr-2">Latest Release</span>-->
  <!--        <v-icon>mdi-open-in-new</v-icon>-->
  <!--      </v-btn>-->
  <!--    </v-app-bar>-->

  <!--    <v-content>-->
  <!--      <HelloWorld />-->
  <!--    </v-content>-->
  <!--  </v-app>-->
</template>

<script>
import Notifications from "@/components/Notifications.vue";
//import HelloWorld from "./components/HelloWorld";

export default {
  name: "App",
  components: {
    // HelloWorld,
    Notifications
  }
};
</script>

<style lang="scss">
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

#nav {
  vertical-align: top;
  padding: 30px;
  display: inline-block;
  width: 200px;

  a {
    font-weight: bold;
    color: #2c3e50;

    &.router-link-exact-active {
      color: #42b983;
    }
  }
}

#view {
  display: inline-block;
}
</style>
