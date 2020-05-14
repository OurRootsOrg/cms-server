<template>
  <div class="collections-list">
    <h1>List Collections</h1>
    <Collection
      v-for="collection in collections.collectionsList"
      :key="collection.id"
      :collection="collection"
      :category="categoryForCollection(collection)"
    />
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
    Promise.all([store.dispatch("categoriesGetAll"), store.dispatch("collectionsGetAll")]).then(() => {
      next();
    });
  },
  computed: {
    categoryForCollection() {
      return collection => {
        return this.categories.categoriesList.find(cat => cat.id === collection.category);
      };
    },
    ...mapState(["categories", "collections"])
  }
};
</script>

<style scoped></style>
