<template>
  <v-container fluid class="home">
    <h1>Accept Invitation</h1>
    <p style="margin: 32px 0;">
      {{ invitations.invitation.name }}, you've been invited to be an {{ getLevel(invitations.invitation.level) }} for
      the <em>{{ invitations.invitation.societyName }}</em
      >.
    </p>
    <div v-if="users.user">
      <p>To accept your invitation, click the button below</p>
      <v-btn color="primary" @click="accept">Accept Invitation</v-btn>
    </div>
    <div v-else>
      <p>Before you can accept the invitation, you need to register or log in</p>
      <v-btn color="primary" @click="login">Register or Log in</v-btn>
    </div>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import { getAuthLevelName } from "@/utils/authLevels";
import Auth from "@/services/Auth";
import NProgress from "nprogress";

function getContent(code, next) {
  return store
    .dispatch("invitationGetForCode", code)
    .then(() => {
      next();
    })
    .catch(() => {
      next("/");
    });
}

export default {
  name: "Home",
  beforeRouteEnter(routeTo, routeFrom, next) {
    getContent(routeTo.params.code, next);
  },
  beforeRouteUpdate(routeTo, routeFrom, next) {
    getContent(routeTo.params.code, next);
  },
  created() {
    if (store.getters.userIsLoggedIn) {
      this.delCookie();
    } else {
      this.setCookie();
    }
  },
  data() {
    return {};
  },
  computed: mapState(["users", "invitations"]),
  methods: {
    setCookie() {
      // set cookie for 15 minutes
      this.$cookies.set("invitation-code", this.invitations.invitation.code, 60 * 15);
    },
    delCookie() {
      this.$cookies.remove("invitation-code");
    },
    getLevel(level) {
      return getAuthLevelName(level);
    },
    login() {
      Auth.login();
    },
    accept() {
      let societyId = this.invitations.invitation.societyId;
      NProgress.start();
      this.delCookie();
      this.$store
        .dispatch("invitationAccept", this.invitations.invitation.code)
        .then(() => {
          this.$router.push({
            name: "society-home",
            params: { society: societyId }
          });
        })
        .catch(() => {
          NProgress.done();
        });
    }
  }
};
</script>
