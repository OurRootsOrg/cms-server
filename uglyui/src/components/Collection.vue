<template>
  <div>
    <div>
      <router-link class="post-link" :to="{ name: 'collection-edit', params: { cid: collection.id } }">{{
        collection.name
      }}</router-link>
      (in {{ category.name }})
    </div>
  </div>
</template>

<script>
export default {
  props: {
    collection: Object,
    category: Object
  }
};
</script>

<style scoped></style>
