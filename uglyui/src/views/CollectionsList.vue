<template>
  <div class="collections-list">
    <h1>Collections</h1>
    <Tabulator
      :data="getCollections()"
      :columns="collectionColumns"
      layout="fitColumns"
      :header-sort="true"
      :selectable="true"
      :resizable-columns="true"
      @rowClicked="rowClicked"
    />
  </div>
</template>

<script>
import { mapState } from "vuex";
import store from "@/store";
import Tabulator from "../components/Tabulator";

export default {
  components: { Tabulator },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([
      store.dispatch("categoriesGetAll"),
      store.dispatch("collectionsGetAll"),
      store.dispatch("postsGetAll")
    ]).then(() => {
      next();
    });
  },
  data() {
    return {
      collectionColumns: [
        {
          title: "Name",
          field: "name",
          headerFilter: "input",
          sorter: "string"
        },
        {
          title: "# Posts",
          field: "postsCount",
          headerFilter: "number",
          sorter: "number"
        },
        {
          title: "Categories",
          field: "categoryNames",
          headerFilter: "input",
          sorter: "string"
        }
      ]
    };
  },
  computed: mapState(["categories", "collections", "posts"]),
  methods: {
    getCollections() {
      return this.collections.collectionsList.map(c => {
        return {
          id: c.id,
          name: c.name,
          postsCount: this.posts.postsList.filter(post => post.collection === c.id).length,
          categoryNames: this.collections.collectionsList
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

<style scoped>
.disabled {
  cursor: not-allowed;
  color: gray;
}
</style>
