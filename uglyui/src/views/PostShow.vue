<template>
  <div class="posts-show">
    <h1>Show Post</h1>
    <p>
      <strong>{{ post.name }}</strong>
    </p>
    <p>Status: {{ post.recordsStatus }}</p>
    <BaseButton
      v-if="post.recordsStatus === 'Draft'"
      v-on:click="publish"
      class="submit-button"
      buttonClass="-fill-gradient"
      >Publish</BaseButton
    >
  </div>
</template>

<script>
import NProgress from "nprogress";

export default {
  props: {
    post: {
      type: Object,
      required: true
    }
  },
  methods: {
    publish() {
      this.post.recordsStatus = "Published";
      NProgress.start();
      this.$store
        .dispatch("postsUpdate", this.post)
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
