<template>
  <div>
    <div>
      <strong>{{ post.name }}</strong> (in {{ collection.name }}) Status: {{ post.recordsStatus }}
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
