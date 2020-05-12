<template>
  <div id="app">
    <div id="nav">
      <div>
        <h3>Authentication</h3>
        <ul>
          <li><router-link to="/auth/register">Register</router-link></li>
          <li><router-link to="/auth/login">Login</router-link></li>
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
      <div>
        <h3>Flatfile demo</h3>
        <ul>
          <li><router-link to="/flatfile">Go</router-link></li>
        </ul>
      </div>
    </div>
    <Notifications />
    <router-view id="view" />
  </div>
</template>

<script>
import Notifications from "@/components/Notifications.vue";

export default {
  components: {
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
