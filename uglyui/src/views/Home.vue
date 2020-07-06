<template>
  <v-container fluid class="home">
    <!--    <img alt="Vue logo" src="../assets/logo.png" />-->
    <h1>Home</h1>
    <Test></Test>
    <!-- Check that the SDK client is not currently loading before accessing is methods -->
    <div v-if="!$auth.loading">
      <div v-if="$auth.user">
        <v-avatar><img :src="$auth.user.picture"/></v-avatar>
        <h2>{{ $auth.user.name }}</h2>
        <p>{{ $auth.user.email }}</p>
      </div>
      <!--      <div>-->
      <!--        <pre>{{ JSON.stringify($auth.user, null, 2) }}</pre>-->
      <!--      </div>-->
      <!-- show login when not authenticated -->
      <v-btn color="primary" v-if="!$auth.isAuthenticated" @click="login">Log in</v-btn>
      <!-- show logout when authenticated -->
      <v-btn v-if="$auth.isAuthenticated" @click="logout">Log out</v-btn>
    </div>
    <br />
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
      <li>Pretty search UI (replace the Search screen on this UI)</li>
    </ul>
    <br />
    <h2>Roadmap for end of the year</h2>
    <ul>
      <li>Records can have images uploaded and associated with them</li>
      <li>Limit volunteer actions (only authorized volunteers can publish)</li>
      <li>Custom transformations (lookup tables) applied to spreadsheet data during upload</li>
      <li>Optional integration with WordPress membership system for paid search</li>
      <li>Optional integration with WordPress to display custom search forms</li>
    </ul>
    <br />
    <h2>Future ideas</h2>
    <ul>
      <li>Searchable books</li>
      <li>CSI integration for indexing support</li>
      <li>Searchable GEDCOMs</li>
      <li>Searchable record streams (users can submit individual records to index)</li>
    </ul>
    <br />
    <h2>Feedback?</h2>
    <ul>
      <li>heatherdevaun@gmail.com</li>
      <li>dallan@gmail.com</li>
    </ul>
  </v-container>
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
