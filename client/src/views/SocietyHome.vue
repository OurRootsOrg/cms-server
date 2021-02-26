<template>
  <div class="society-home">
    <h2>{{ societySummaries.societySummary.name }}</h2>
    <p>Your membership level in this society: {{ getLevel }}</p>
  </div>
</template>

<script>
import { mapState } from "vuex";

const levels = ["Guest", "Reader", "Contributor", "Editor", "Admin"];

export default {
  name: "SocietyHome",
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    console.log("societyHome.beforeRouteEnter");
    next();
  },
  beforeRouteUpdate: function(routeTo, routeFrom, next) {
    console.log("societyHome.beforeRouteUpdate");
    next();
  },
  created() {
    console.log("societyHome.create", this.$route.fullPath);
  },
  computed: {
    getLevel() {
      let level = this.societyUsers.societyUserCurrent.level;
      return levels[level];
    },
    ...mapState(["societySummaries", "societyUsers"])
  }
};
</script>

<style scoped></style>
