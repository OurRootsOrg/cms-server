<template>
  <div class="categories-list">
    <h1>List Categories</h1>
    <Category
      v-for="category in categories.categoriesList"
      :key="category.id"
      :category="category"
      :coll-count="collectionsForCategory(category.id).length"
    >
      <a href="" @click.prevent="del(category.id)" :class="{ disabled: collectionsForCategory(category.id).length > 0 }"
        >(del)</a
      >
    </Category>
  </div>
</template>

<script>
import Category from "@/components/Category.vue";
import { mapState } from "vuex";
import store from "@/store";
import NProgress from "nprogress";

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
      return id => {
        return this.collections.collectionsList.filter(coll => coll.categories.includes(id));
      };
    },
    ...mapState(["categories", "collections"])
  },
  methods: {
    del(id) {
      if (this.collectionsForCategory(id).length > 0) {
        return;
      }
      NProgress.start();
      this.$store
        .dispatch("categoriesDelete", id)
        .then(() => {
          NProgress.done();
        })
        .catch(() => {
          NProgress.done();
        });
    }
  }
};
</script>

<style scoped>
.disabled {
  cursor: not-allowed;
  color: gray;
}
</style>
