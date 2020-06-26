<template>
  <div class="collections-list">
    <h1>List Collections</h1>
    <Collection
      v-for="collection in collections.collectionsList"
      :key="collection.id"
      :collection="collection"
      :categories="categoriesForCollection(collection)"
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
    categoriesForCollection() {
      return collection => {
        return this.categories.categoriesList.filter(cat => collection.categories.includes(cat.id));
      };
    },
    ...mapState(["categories", "collections"])
  }
};
</script>

<style scoped></style>
