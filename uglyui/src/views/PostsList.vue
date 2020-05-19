<template>
  <div class="posts-list">
    <h1>List Posts</h1>
    <Post v-for="post in posts.postsList" :key="post.id" :post="post" :collection="collectionForPost(post)" />
  </div>
</template>

<script>
import Post from "@/components/Post.vue";
import { mapState } from "vuex";
import store from "@/store";

export default {
  components: {
    Post
  },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([store.dispatch("collectionsGetAll"), store.dispatch("postsGetAll")]).then(() => {
      next();
    });
  },
  computed: {
    collectionForPost() {
      return post => {
        return this.collections.collectionsList.find(coll => coll.id === post.collection);
      };
    },
    ...mapState(["collections", "posts"])
  }
};
</script>

<style scoped></style>
