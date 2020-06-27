<template>
  <div class="collections-list">
    <h1>List Collections</h1>
    <Collection
      v-for="collection in collections.collectionsList"
      :key="collection.id"
      :collection="collection"
      :categories="categoriesForCollection(collection)"
      :posts="postsForCollection(collection.id)"
    >
    </Collection>
  </div>
</template>

<script>
import Collection from "@/components/Collection.vue";
import { mapState } from "vuex";
import store from "@/store";

export default {
  components: {
    Collection
  },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([
      store.dispatch("categoriesGetAll"),
      store.dispatch("collectionsGetAll"),
      store.dispatch("postsGetAll")
    ]).then(() => {
      next();
    });
  },
  computed: {
    categoriesForCollection() {
      return collection => {
        return this.categories.categoriesList.filter(cat => collection.categories.includes(cat.id));
      };
    },
    postsForCollection() {
      return id => {
        return this.posts.postsList.filter(post => post.collection === id);
      };
    },
    ...mapState(["categories", "collections", "posts"])
  }
};
</script>

<style scoped>
.disabled {
  cursor: not-allowed;
  color: gray;
}
</style>
