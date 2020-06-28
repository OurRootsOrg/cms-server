<template>
  <div>
    <div>
      <router-link class="post-link" :to="{ name: 'collection-edit', params: { cid: collection.id } }">{{
        collection.name
      }}</router-link>
      [ {{ posts.length }} posts ]
      <!--      (in <span v-for="(category, $ix) in categories" :key="$ix">{{ category.name }}</span-->
      in {{ categories.map(cat => cat.name).join(", ") }}
      <slot></slot>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    collection: Object,
    categories: Array,
    posts: Array
  }
};
</script>

<style scoped></style>
