<template>
  <v-container fluid class="home">
    <h1>Home</h1>
    <div v-if="users.user">
      <p v-if="users.user.name">Welcome {{ users.user.name }}</p>
      <p>Select a society</p>
      <ul>
        <li v-for="society in societySummaries.societySummariesList" :key="society.id">
          <v-btn text @click="openSociety(society.id)">{{ society.name }}</v-btn>
        </li>
      </ul>
      <p>Or create a new society</p>
      <v-dialog v-model="dialog" persistent max-width="600px">
        <template v-slot:activator="{ on, attrs }">
          <v-btn color="primary" v-bind="attrs" v-on="on">Create a Society</v-btn>
        </template>
        <v-card>
          <v-card-title>
            <span class="headline">Create a Society</span>
          </v-card-title>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col cols="12">
                  <v-text-field v-model="societyName" label="Name"></v-text-field>
                </v-col>
              </v-row>
            </v-container>
            <p style="margin-left: 16px">New societies are free for 30 days. No credit card required.</p>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn text @click="dialog = false">
              Cancel
            </v-btn>
            <v-btn color="blue darken-1" text @click="createSociety">
              Save
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </div>
    <div v-else>
      <p>Introductory text goes here</p>
      <v-btn color="primary" @click="login">Register or Log in</v-btn>
    </div>
  </v-container>
</template>

<script>
import Auth from "@/services/Auth";
import Vue from "vue";
import { mapState } from "vuex";
import store from "@/store";
import NProgress from "nprogress";

export default {
  name: "Home",
  beforeRouteEnter(routeTo, routeFrom, next) {
    let code = routeTo.query ? routeTo.query.code : "";
    if (!code) {
      code = Vue.$cookies.get("invitation-code");
    }
    if (code) {
      next("/invitation/" + code);
    } else {
      next();
    }
  },
  created() {
    if (store.getters.userIsLoggedIn) {
      this.loadSocietiesForUser();
    }
  },
  data() {
    return {
      dialog: false,
      societyName: ""
    };
  },
  computed: mapState(["users", "societySummaries"]),
  methods: {
    loadSocietiesForUser() {
      store.dispatch("societySummariesGetAll");
    },
    // Log the user in
    login() {
      Auth.login();
    },
    // Log the user out
    logout() {
      Auth.logout();
    },
    openSociety(id) {
      this.$router.push({
        name: "society-home",
        params: { society: id }
      });
    },
    createSociety() {
      NProgress.start();
      this.$store
        .dispatch("societiesCreate", { name: this.societyName })
        .then(result => {
          console.log("createSociety result", result);
          this.openSociety(result.id);
        })
        .catch(() => {
          NProgress.done();
        });
    }
  }
};
</script>
