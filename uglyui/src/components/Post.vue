<template>
  <div>
    <div>
      <router-link class="post-link" :to="{ name: 'post-show', params: { pid: post.id } }">{{ post.name }}</router-link>
      (in {{ collection.name }}) Status: {{ post.recordsStatus }}
    </div>
  </div>
</template>

<script>
export default {
  props: {
    post: Object,
    collection: Object
  }
};
</script>

<style scoped></style>
