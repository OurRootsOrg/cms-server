<template>
  <div>
    <div>
      <router-link class="post-link" :to="{ name: 'collection-edit', params: { cid: collection.id } }">{{
        collection.name
      }}</router-link>
      <!--      (in <span v-for="(category, $ix) in categories" :key="$ix">{{ category.name }}</span-->
      (in {{ categories.map(cat => cat.name).join(", ") }})
    </div>
  </div>
</template>

<script>
export default {
  props: {
    collection: Object,
    categories: Array
  }
};
</script>

<style scoped></style>
