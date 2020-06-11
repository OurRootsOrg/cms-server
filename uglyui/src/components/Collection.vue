<template>
  <div>
    <div>
      <router-link class="post-link" :to="{ name: 'collection-edit', params: { cid: extractId(collection.id) } }">{{
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
