<template>
  <v-container class="collections-list">
    <h1>Collections</h1>
    <v-btn small color="primary" class="mt-2" to="/collections/create">
      Create a new collection
    </v-btn>
    <v-row fluid>
      <v-col cols="12" md="5" class="pt-0">
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="Search for a collection or category"
          single-line
          hide-details
        ></v-text-field>
      </v-col>
      <v-col cols="12">
        <v-data-table
          :items="getCollections()"
          :headers="headers"
          sortable
          sort-by="name"
          :search="search"
          :footer-props="{ 'items-per-page-options': [10, 25, 50] }"
          :items-per-page="25"
          @click:row="rowClicked"
          dense
          v-columns-resizable
        >
          <template v-slot:[`item.icon`]="{ item }">
            <v-btn icon small :to="{ name: 'collection-edit', params: { cid: item.id } }">
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
    Promise.all([
      store.dispatch("categoriesGetAll"),
      store.dispatch("collectionsGetAll"),
      store.dispatch("postsGetAll")
    ])
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
        { text: "Location", value: "location" },
        { text: "# Posts", value: "postsCount" },
        { text: "Categories", value: "categoryNames" },
        { text: "", value: "icon", align: "right", width:"15px" }
      ],
      search: ""
    };
  },
  computed: mapState(["categories", "collections", "posts"]),
  methods: {
    getCollections() {
      return this.collections.collectionsList.map(c => {
        return {
          id: c.id,
          name: c.name,
          location: c.location,
          postsCount: this.posts.postsList.filter(post => post.collection === c.id).length,
          categoryNames: this.categories.categoriesList
            .filter(cat => c.categories.includes(cat.id))
            .map(cat => cat.name)
            .join(", ")
        };
      });
    },
    rowClicked(coll) {
      this.$router.push({
        name: "collection-edit",
        params: { cid: coll.id }
      });
    }
  }
};
</script>
