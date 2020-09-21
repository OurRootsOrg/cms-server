<template>
  <v-container fluid class="home">
    <h1>Home</h1>
    <p>
      Select an option to the left. If you have not registered or logged in, do that first.
    </p>
    <p>
      <i>NOTE:</i> The demo server hibernates after 15 minutes of inactivity to save money. It takes 30 seconds for the
      server to wake up. You won't experience the 30-second delay in production.
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
