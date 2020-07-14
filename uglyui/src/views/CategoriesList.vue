<template>
  <v-container class="categories-list">
    <v-layout row>
      <v-flex>
        <h1>Categories</h1>
        <v-btn small color="primary" class="mt-2 mb-5" to="/categories/create">
          Create a new category
        </v-btn>
      </v-flex>
    </v-layout>
    <v-layout row>
      <v-flex class="mt-1">
        <Tabulator
          :data="getCategories()"
          :columns="categoryColumns"
          layout="fitColumns"
          :header-sort="true"
          :selectable="true"
          :resizable-columns="true"
          @rowClicked="rowClicked"
        />
      </v-flex>
    </v-layout>
  </v-container>
</template>
<script>
import { mapState } from "vuex";
import store from "@/store";
import Tabulator from "../components/Tabulator";

export default {
  components: { Tabulator },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([store.dispatch("categoriesGetAll"), store.dispatch("collectionsGetAll")])
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
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
