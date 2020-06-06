<template>
  <div class="search">
    <h1>Search</h1>
    <form @submit.prevent="go">
      <BaseInput label="Given" v-model="query.given" type="text" placeholder="Given name" class="field" />
      <BaseInput label="Surname" v-model="query.surname" type="text" placeholder="Surname" class="field" />
      <BaseButton type="submit" class="submit-button" buttonClass="-fill-gradient -size-small">Go</BaseButton>
    </form>
    <div v-if="search.searchTotal > 0">
      <p>Showing 1 - {{ search.searchList.length }} of {{ search.searchTotal }}</p>
      <SearchResult v-for="(result, $ix) in search.searchList" :key="$ix" :result="result" />
    </div>
  </div>
</template>

<script>
import SearchResult from "../components/SearchResult.vue";
import NProgress from "nprogress";
import { mapState } from "vuex";

export default {
  components: {
    SearchResult
  },
  data() {
    return {
      query: {}
    };
  },
  computed: mapState(["search"]),
  methods: {
    go() {
      NProgress.start();
      this.$store
        .dispatch("search", {
          given: this.query.given,
          surname: this.query.surname
        })
        .then(() => {
          NProgress.done();
        })
        .catch(() => {
          NProgress.done();
        });
    }
  }
};
</script>

<style scoped>
.field {
  margin: 8px;
}
</style>
