<template>
  <div class="categories-list">
    <h1>List Categories</h1>
    <Category
      v-for="category in categories.categoriesList"
      :key="category.id"
      :category="category"
      :coll-count="collectionsForCategory(category).length"
    />
  </div>
</template>

<script>
import Category from "@/components/Category.vue";
import { mapState } from "vuex";
import store from "@/store";

export default {
  components: {
    Category
  },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([store.dispatch("categoriesGetAll"), store.dispatch("collectionsGetAll")]).then(() => {
      next();
    });
  },
  computed: {
    collectionsForCategory() {
      return category => {
        return this.collections.collectionsList.filter(coll => coll.category.id === category.id);
      };
    },
    ...mapState(["categories", "collections"])
  }
};
</script>

<style scoped></style>
