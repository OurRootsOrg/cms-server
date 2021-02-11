<template>
  <v-container fluid class="home">
    <h1>Home</h1>
    <p>
      Select an option to the left. If you have not registered or logged in, do that first.
    </p>
    <p>
      <i>NOTE:</i> This demo server hibernates after 15 minutes of inactivity to save money. It takes 30 seconds for the
      server to wake up. You won't experience a delay in production.
    </p>
    <!-- Check that the SDK client is not currently loading before accessing is methods -->
    <div v-if="user.user">
      <v-avatar v-if="user.user.picture"><img :src="user.user.picture"/></v-avatar>
      <h2>{{ user.user.name }}</h2>
      <p>{{ user.user.email }}</p>
    </div>
    <!-- show login when not authenticated -->
    <v-btn color="primary" v-if="!user.user" @click="login">Log in</v-btn>
    <!-- show logout when authenticated -->
    <v-btn v-if="!!user.user" @click="logout">Log out</v-btn>
    <br />
    <h2>Future ideas</h2>
    <ul>
      <li>Searchable books</li>
      <li>Searchable GEDCOMs</li>
    </ul>
    <br />
    <h2>Feedback?</h2>
    <ul>
      <li>heather@ourroots.org</li>
      <li>dallan@ourroots.org</li>
    </ul>
  </v-container>
</template>

<script>
import Auth from "@/services/Auth";
import { mapState } from "vuex";

export default {
  name: "Home",
  computed: mapState(["user"]),
  methods: {
    // Log the user in
    login() {
      Auth.login();
    },
    // Log the user out
    logout() {
      Auth.logout();
    }
  }
};
</script>
