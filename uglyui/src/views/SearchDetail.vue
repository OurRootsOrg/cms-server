<template>
  <div class="search-detail">
    <h1>Record detail</h1>
    <div v-if="search.searchResult.person.role !== 'principal'">
      <div>Name: {{ search.searchResult.person.name }}</div>
      <div>Role: {{ search.searchResult.person.role }}</div>
    </div>
    <h4>{{ search.searchResult.collectionName }}</h4>
    <div v-for="(lv, $ix) in search.searchResult.record" :key="$ix">{{ lv.label }}: {{ lv.value }}</div>
  </div>
</template>

<script>
import { mapState } from "vuex";
import store from "@/store";

export default {
  beforeRouteEnter(routeTo, routeFrom, next) {
    store
      .dispatch("searchGetResult", routeTo.params.rid)
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  computed: mapState(["search"])
};
</script>

<style scoped></style>
