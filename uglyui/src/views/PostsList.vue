<template>
  <v-container class="posts-list">
    <v-layout row>
      <v-flex>
        <h1>Posts</h1>
        <v-btn small color="primary" class="mt-2 mb-5" to="/posts/create">
          Create a new post
        </v-btn>
      </v-flex>
    </v-layout>
    <v-layout row>
      <v-flex class="mt-1">
        <Tabulator
          :data="getPosts()"
          :columns="getPostColumns()"
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
import { getMetadataColumn } from "../utils/metadata";
import Tabulator from "../components/Tabulator";

export default {
  components: { Tabulator },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([store.dispatch("collectionsGetAll"), store.dispatch("postsGetAll"), store.dispatch("settingsGet")])
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  computed: mapState(["collections", "posts", "settings"]),
  methods: {
    getPosts() {
      return this.posts.postsList.map(p => {
        return {
          id: p.id,
          name: p.name,
          recordsStatus: p.recordsStatus,
          hasData: !!p.recordsKey,
          collectionName: this.collections.collectionsList.find(coll => coll.id === p.collection).name,
          ...p.metadata
        };
      });
    },
    getPostColumns() {
      let cols = [
        {
          title: "Name",
          field: "name",
          headerFilter: "input",
          sorter: "string"
        },
        {
          title: "Status",
          field: "recordsStatus",
          headerFilter: "select",
          headerFilterParams: {
            values: true
          },
          sorter: "string"
        },
        {
          title: "Has Data",
          field: "hasData",
          hozAlign: "center",
          formatter: "tickCross",
          headerFilter: "tickCross",
          sorter: "boolean"
        },
        {
          title: "Collection",
          field: "collectionName",
          headerFilter: "input",
          sorter: "string"
        }
      ];
      cols.push(...this.settings.settings.postMetadata.map(pf => getMetadataColumn(pf)));
      return cols;
    },
    rowClicked(post) {
      this.$router.push({
        name: "post-edit",
        params: { pid: post.id }
      });
    }
  }
};
</script>

<style scoped>
.tabulator {
  width: 750px;
}
.create {
  margin-top: 8px;
}
</style>
