<template>
  <div class="society-home">
    <h2>{{ societySummaries.societySummary.name }}</h2>
    <p style="margin-top: 32px">Created: {{ getCreationDate }}</p>
    <p style="margin-top: 32px">Your membership level in this society: {{ getLevel }}</p>
  </div>
</template>

<script>
import { mapState } from "vuex";
import { getAuthLevelName } from "@/utils/authLevels";

const months = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December"
];

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
    getCreationDate() {
      let d = new Date(this.societySummaries.societySummary.insert_time);
      return `${d.getDate()} ${months[d.getMonth()]} ${d.getFullYear()}`;
    },
    getLevel() {
      return getAuthLevelName(this.societyUsers.societyUserCurrent.level);
    },
    ...mapState(["societySummaries", "societyUsers"])
  }
};
</script>

<style scoped></style>
