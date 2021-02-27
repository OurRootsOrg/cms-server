<template>
  <div class="society-home">
    <h2>{{ societySummaries.societySummary.name }}</h2>
    <p>Your membership level in this society: {{ getLevel }}</p>
  </div>
</template>

<script>
import { mapState } from "vuex";
import { getAuthLevelName } from "@/utils/authLevels";

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
      return getAuthLevelName(this.societyUsers.societyUserCurrent.level);
    },
    ...mapState(["societySummaries", "societyUsers"])
  }
};
</script>

<style scoped></style>
