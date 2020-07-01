<template>
  <div class="home">
    <!--    <img alt="Vue logo" src="../assets/logo.png" />-->
    <h1>Home</h1>
    <Test></Test>
    <!-- Check that the SDK client is not currently loading before accessing is methods -->
    <div v-if="!$auth.loading">
      <div v-if="$auth.user">
        <img :src="$auth.user.picture" />
        <h2>{{ $auth.user.name }}</h2>
        <p>{{ $auth.user.email }}</p>
      </div>
      <!--      <div>-->
      <!--        <pre>{{ JSON.stringify($auth.user, null, 2) }}</pre>-->
      <!--      </div>-->
      <!-- show login when not authenticated -->
      <button v-if="!$auth.isAuthenticated" @click="login">Log in</button>
      <!-- show logout when authenticated -->
      <button v-if="$auth.isAuthenticated" @click="logout">Log out</button>
    </div>
    <h2>Roadmap for end of September</h2>
    <ul>
      <li>Dashboard</li>
      <li>List users and authorize who can sign in</li>
      <li>Search dates</li>
      <li>Search places</li>
      <li>Include common nicknames in name search</li>
      <li>Choose name coders in global settings</li>
      <li>Glue households</li>
      <li>Collection field templates</li>
      <li>Pretty search UI</li>
    </ul>
    <h2>Feedback?</h2>
    <ul>
      <li>heatherdevaun@gmail.com</li>
      <li>dallan@gmail.com</li>
    </ul>
  </div>
</template>

<script>
// @ is an alias to /src
import Test from "@/components/Test.vue";

export default {
  name: "Home",
  components: {
    Test
  },
  methods: {
    // Log the user in
    login() {
      this.$auth.loginWithRedirect();
    },
    // Log the user out
    logout() {
      this.$auth.logout({
        returnTo: window.location.origin
      });
    }
  }
};
</script>
