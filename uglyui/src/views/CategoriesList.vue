<template>
  <div class="categories-list">
    <h1>Categories</h1>
    <Tabulator
      :data="getCategories()"
      :columns="categoryColumns"
      layout="fitColumns"
      :header-sort="true"
      :selectable="true"
      :resizable-columns="true"
      @rowClicked="rowClicked"
    />
    <v-btn color="primary" class="mt-4" to="/categories/create">
      Create a new category
    </v-btn>
  </div>
</template>

<script>
import { mapState } from "vuex";
import store from "@/store";
import Tabulator from "../components/Tabulator";

export default {
  components: { Tabulator },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([store.dispatch("categoriesGetAll"), store.dispatch("collectionsGetAll")]).then(() => {
      next();
    });
  },
  data() {
    return {
      categoryColumns: [
        {
          title: "Name",
          field: "name",
          headerFilter: "input",
          sorter: "string"
        },
        {
          title: "# Collections",
          field: "collectionsCount",
          headerFilter: "number",
          sorter: "number"
        }
      ]
    };
  },
  computed: mapState(["categories", "collections"]),
  methods: {
    getCategories() {
      return this.categories.categoriesList.map(c => {
        return {
          id: c.id,
          name: c.name,
          collectionsCount: this.collections.collectionsList.filter(coll => coll.categories.includes(c.id)).length
        };
      });
    },
    rowClicked(cat) {
      this.$router.push({
        name: "category-edit",
        params: { cid: cat.id }
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
.create {
  margin-top: 8px;
}
</style>
