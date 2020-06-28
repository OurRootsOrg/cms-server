<template>
  <div class="posts-show">
    <h1>Post</h1>
    <p>
      <strong>{{ posts.post.name }}</strong>
    </p>
    <p>Status: {{ posts.post.recordsStatus }}</p>
    <BaseButton v-if="posts.post.recordsStatus === 'Draft'" @click="publish" class="btn" buttonClass="-fill-gradient"
      >Publish Post</BaseButton
    >
    <BaseButton v-if="posts.post.recordsStatus === 'Published'" @click="unpublish" class="btn" buttonClass="danger"
      >Unpublish Post</BaseButton
    >
    <BaseButton v-if="posts.post.recordsStatus === 'Draft'" @click="del" class="btn" buttonClass="danger"
      >Delete Post</BaseButton
    >
    <Tabulator :data="records.recordsList.map(r => r.data)" :columns="getColumns()" />
  </div>
</template>

<script>
import NProgress from "nprogress";
import Tabulator from "../components/Tabulator";
import { mapState } from "vuex";
import store from "@/store";

export default {
  components: { Tabulator },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([
      store.dispatch("postsGetOne", routeTo.params.pid),
      store.dispatch("recordsGetForPost", routeTo.params.pid)
    ]).then(() => {
      store.dispatch("collectionsGetOne", store.state.posts.post.collection).then(() => {
        next();
      });
    });
  },
  computed: mapState(["collections", "posts", "records"]),
  methods: {
    getColumns() {
      return this.collections.collection.fields.map(f => {
        return { title: f.header, field: f.header };
      });
    },
    publish() {
      this.update("Published");
    },
    unpublish() {
      this.update("Draft");
    },
    update(status) {
      let post = Object.assign({}, this.posts.post);
      post.recordsStatus = status;
      NProgress.start();
      this.$store
        .dispatch("postsUpdate", post)
        .then(() => {
          this.$router.push({
            name: "posts-list"
          });
        })
        .catch(() => {
          NProgress.done();
        });
    },
    del() {
      NProgress.start();
      this.$store
        .dispatch("postsDelete", this.posts.post.id)
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

<style scoped>
.btn {
  margin-bottom: 24px;
}
</style>
