<template>
  <v-container fluid class="home">
    <h2>A database management system for genealogy societies</h2>
    <div v-if="users.user">
      <p class="intro">Select a society</p>
      <p class="intro" v-if="societySummaries.societySummariesList.length == 1">
        Everyone is a member of the Sandbox Society. The Sandbox is a great place to get a feel for how things work
        before creating your own society.
      </p>
      <ul>
        <li v-for="society in societySummaries.societySummariesList" :key="society.id">
          <v-btn text @click="openSociety(society.id)">{{ society.name }}</v-btn>
        </li>
      </ul>
      <p class="intro">Or create a new society</p>
      <v-dialog v-model="dialog" persistent max-width="600px">
        <template v-slot:activator="{ on, attrs }">
          <div class="btn">
            <v-btn color="primary" v-bind="attrs" v-on="on">Create a Society</v-btn>
          </div>
        </template>
        <v-card>
          <v-card-title>
            <span class="headline">Create a Society</span>
          </v-card-title>
          <v-card-text>
            <v-container>
              <v-row v-if="isEmailConfirmed">
                <v-col cols="12">
                  <v-text-field v-model="societyName" label="Name"></v-text-field>
                </v-col>
              </v-row>
              <v-row v-else>
                <p style="margin-left: 16px">Before creating a new society, you need to confirm your email.</p>
              </v-row>
            </v-container>
            <p style="margin-left: 16px">New societies are free for 30 days. No credit card required.</p>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn text @click="dialog = false">
              Cancel
            </v-btn>
            <v-btn v-if="isEmailConfirmed" color="blue darken-1" text @click="createSociety">
              Save
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
      <h3 style="margin-top: 48px">Now that you've signed in, watch these videos</h3>
      <p style="margin-top: 24px">About the sandbox society</p>
      <div>
        <iframe
          style="margin: 0"
          width="560"
          height="315"
          src="https://www.youtube.com/embed/vQ5wM5QZhCU"
          frameborder="0"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen
        ></iframe>
      </div>
      <p style="margin-top: 24px">Make vital records searchable</p>
      <div>
        <iframe
          style="margin: 0"
          width="560"
          height="315"
          src="https://www.youtube.com/embed/Hf3bazriQOY"
          frameborder="0"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen
        ></iframe>
      </div>
      <p style="margin-top: 24px">Make census records searchable</p>
      <div>
        <iframe
          style="margin: 0"
          width="560"
          height="315"
          src="https://www.youtube.com/embed/-FhajEuC3UU"
          frameborder="0"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen
        ></iframe>
      </div>
      <p style="margin-top: 24px">Make library catalogs searchable</p>
      <div>
        <iframe
          style="margin: 0"
          width="560"
          height="315"
          src="https://www.youtube.com/embed/s8qTPWZ9r9E"
          frameborder="0"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen
        ></iframe>
      </div>
    </div>
    <div v-else>
      <p class="intro">
        Upload your genealogy records, media, and books and make them searchable on your society website.
<!--        <a href="https://www.ourroots.org">Click here for more information.</a>-->
      </p>
      <div class="btn">
        <v-btn color="primary" @click="login">Log in</v-btn>
      </div>
    </div>
    <v-footer paddless fixed style="background-color: #fafafa">
      <v-col cols="12">
<!--        <a style="padding-right: 8px;" href="/static/privacy.html">Privacy policy</a>-->
<!--        <a style="padding-right: 8px;" href="/static/terms.html">Terms of service</a>-->
<!--        <a style="padding-right: 8px;" href="/static/cookie.html">Cookie policy</a>-->
      </v-col>
    </v-footer>
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
  computed: {
    isEmailConfirmed() {
      return this.users.user.email_confirmed;
    },
    ...mapState(["users", "societySummaries"])
  },
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
<style scoped>
.intro {
  margin: 16px 0;
  max-width: 600px;
}
.btn {
  margin: 32px 0;
}
</style>
