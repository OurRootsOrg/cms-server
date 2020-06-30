<template>
  <div class="posts-list">
    <h1>Posts</h1>
    <Tabulator
      :data="getPosts()"
      :columns="getPostColumns()"
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

function getMetadataColumn(pf) {
  switch (pf.type) {
    case "string":
      return {
        title: pf.name,
        field: pf.name,
        tooltip: pf.tooltip,
        headerFilter: "input",
        sorter: "string"
      };
    case "number":
      return {
        title: pf.name,
        field: pf.name,
        tooltip: pf.tooltip,
        headerFilter: "number",
        sorter: "number"
      };
    case "date":
      return {
        title: pf.name,
        field: pf.name,
        hozAlign: "center",
        tooltip: pf.tooltip,
        headerFilter: "input",
        sorter: "date",
        sorterParams: {
          format: "DD MMM YYYY",
          alignEmptyValues: "top"
        }
      };
    case "boolean":
      return {
        title: pf.name,
        field: pf.name,
        tooltip: pf.tooltip,
        hozAlign: "center",
        formatter: "tickCross",
        headerFilter: "tickCross",
        sorter: "boolean"
      };
  }
}

export default {
  components: { Tabulator },
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([
      store.dispatch("collectionsGetAll"),
      store.dispatch("postsGetAll"),
      store.dispatch("settingsGet")
    ]).then(() => {
      next();
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
</style>
