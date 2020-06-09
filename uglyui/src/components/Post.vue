<template>
  <div>
    <div>
      <router-link class="post-link" :to="{ name: 'post-show', params: { pid: extractId(post.id) } }">{{
        post.name
      }}</router-link>
      (in {{ collection.name }}) Status: {{ post.recordsStatus }}
    </div>
  </div>
</template>

<script>
export default {
  props: {
    post: Object,
    collection: Object
  },
  computed: {
    extractId() {
      return id => {
        return id.substring(id.lastIndexOf("/") + 1);
      };
    }
  }
};
</script>

<style scoped></style>
