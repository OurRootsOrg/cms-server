<template>
  <v-container fluid class="home">
    <h2>A database management system for genealogy societies!</h2>
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
      <h3 style="margin-top: 48px">Now that you've signed in, watch this tutorial</h3>
      <iframe
        style="margin: 24px 0"
        width="560"
        height="315"
        src="https://www.youtube.com/embed/WsJGYQk_lB4"
        frameborder="0"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
        allowfullscreen
      ></iframe>
    </div>
    <div v-else>
      <p class="intro">
        Upload your genealogy records and media and make it searchable on your wordpress-based website.
      </p>
      <p class="intro">
        <strong>Privacy:</strong> You can either make your records public, limit search to members only, or make your
        records publicly searchable but limit record details or media to members only. It's your choice. Non-society
        members wanting to view members-only content get directed to your society registration page.
      </p>
      <p class="intro">
        <strong>Pricing:</strong> You can have up to 500,000 records and 20Gb of media for $15/month. Additional media
        storage is $0.10/Gb/month, and additional records are $1/500,000 records/month.
      </p>
      <p class="intro">
        <strong>Free trial:</strong> The first 30 days are free. Go ahead and register below. There's no obligation.
      </p>
      <p class="intro">
        <strong>Who we are:</strong> We are a group of software engineers interested in helping genealogy societies. Do
        you have development skills? If so, please join us! OurRoots is completely
        <a href="https://github.com/OurRootsOrg/cms-server">open source</a>. You can host it yourself if you want. We
        welcome volunteers!
      </p>
      <div class="btn">
        <v-btn color="primary" @click="login">Register or Log in</v-btn>
      </div>
      <h3>Quick overview of the OurRoots database management system</h3>
      <p class="intro">
        <iframe
          style="margin-bottom: 16px"
          width="560"
          height="315"
          src="https://www.youtube.com/embed/obsMYYCCbag"
          frameborder="0"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen
        ></iframe>
      </p>
    </div>
    <div class="intro">
      <strong>More information</strong>
      <ul>
        <li><a href="https://www.facebook.com/groups/546537972928277">Facebook group</a></li>
        <li><a href="https://ourroots.org/knowledge-base/">Knowledge base</a></li>
        <li><a href="https://ourroots.org/ticket-desk/">Support desk</a></li>
      </ul>
    </div>
    <v-footer paddless fixed style="background-color: #fafafa">
      <v-col cols="12">
        <a style="padding-right: 8px;" href="/static/privacy.html">Privacy policy</a>
        <a style="padding-right: 8px;" href="/static/terms.html">Terms of service</a>
        <a style="padding-right: 8px;" href="/static/cookie.html">Cookie policy</a>
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
