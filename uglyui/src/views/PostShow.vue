<template>
  <div class="posts-show">
    <h1>Show Post</h1>
    <p>
      <strong>{{ posts.post.name }}</strong>
    </p>
    <p>Status: {{ posts.post.recordsStatus }}</p>
    <BaseButton
      v-if="posts.post.recordsStatus === 'Draft'"
      v-on:click="publish"
      class="submit-button"
      buttonClass="-fill-gradient"
      >Publish</BaseButton
    >
    <Tabulator :data="records.recordsList.map(r => r.data)" :columns="columns" />
  </div>
</template>

<script>
import NProgress from "nprogress";
import Tabulator from "../components/Tabulator";
import { mapState } from "vuex";
import store from "@/store";

export default {
  components: { Tabulator },
  data() {
    return {
      columns: [
        { title: "Given", field: "given" },
        { title: "Surname", field: "surname" }
      ]
    };
  },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([
      store.dispatch("postsGetOne", routeTo.params.pid),
      store.dispatch("recordsGetForPost", routeTo.params.pid)
    ]).then(() => {
      next();
    });
  },
  computed: mapState(["posts", "records"]),
  methods: {
    publish() {
      this.posts.post.recordsStatus = "Published";
      NProgress.start();
      this.$store
        .dispatch("postsUpdate", this.posts.post)
        .then(() => {
          this.$router.push({
            name: "posts-list"
          });
        })
        .catch(() => {
          NProgress.done();
        });
    }
  }
};
</script>

<style scoped></style>
