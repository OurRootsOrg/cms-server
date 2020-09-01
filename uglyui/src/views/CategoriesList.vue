<template>
  <v-container class="categories-list">
    <h1>Categories</h1>
    <v-btn small color="primary" class="mt-2" to="/categories/create">
      Create a new category
    </v-btn>
    <v-row fluid>
      <v-col cols="12" md="5" class="pt-0">
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="Search for a category"
          single-line
          hide-details
        ></v-text-field>
      </v-col>
      <v-col cols="12">
        <v-data-table
          :items="getCategories()"
          :headers="headers"
          sortable
          sort-by="name"
          :search="search"
          :footer-props="{
            'items-per-page-options': [10, 25, 50]
          }"
          :items-per-page="25"
          @click:row="rowClicked"
          dense
          class="rowHover"
          v-columns-resizable
        >
          <template v-slot:[`item.icon`]="{ item }">
            <v-btn icon small :to="{ name: 'category-edit', params: { cid: item.id } }">
              <v-icon right>mdi-chevron-right</v-icon>
            </v-btn>
          </template>
        </v-data-table>
      </v-col>
    </v-row>
  </v-container>
</template>
<script>
import { mapState } from "vuex";
import store from "@/store";

export default {
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
      headers: [
        { text: "Name", value: "name" },
        { text: "# Collections", value: "collectionsCount" },
        { text: "", value: "icon", align: "right", width:"15px" }
      ],
      search: ""
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
