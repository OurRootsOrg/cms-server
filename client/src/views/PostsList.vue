<template>
  <v-container class="posts-list">
    <h1>Posts</h1>
    <v-btn small color="primary" class="mt-2" to="/posts/create">
      Create a new post
    </v-btn>
    <v-row class="d-flex justify-end">
      <v-col cols="12" md="2">
        <v-select v-model="postStatusFilter" :items="postStatusOptions" label="Status" multiple></v-select>
      </v-col>
      <v-col cols="12" md="2">
        <v-select v-model="recordsStatusFilter" :items="recordsStatusOptions" label="Records" multiple></v-select>
      </v-col>
      <v-col cols="12" md="2">
        <v-select v-model="imagesStatusFilter" :items="imagesStatusOptions" label="Images" multiple></v-select>
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field v-model="search" append-icon="mdi-magnify" label="Search" single-line hide-details></v-text-field>
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="12">
        <v-data-table
          :items="getPosts()"
          :headers="getPostColumns()"
          sortable
          sort-by="name"
          :search="search"
          :footer-props="{
            'items-per-page-options': [10, 25, 50]
          }"
          :items-per-page="25"
          @click:row="rowClicked"
          dense
          class="rowHover postsTable"
          v-columns-resizable
        >
          <template v-slot:[`item.icon`]="{ item }">
            <v-btn icon small :to="{ name: 'post-edit', params: { pid: item.id } }">
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
import { getMetadataColumn } from "../utils/metadata";

export default {
  beforeRouteEnter(routeTo, routeFrom, next) {
    Promise.all([store.dispatch("collectionsGetAll"), store.dispatch("postsGetAll"), store.dispatch("settingsGet")])
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  data() {
    return {
      search: "",
      status: "",
      postStatusFilter: [],
      recordsStatusFilter: [],
      imagesStatusFilter: [],
      postStatusOptions: ["Published", "Draft"],
      recordsStatusOptions: ["Loaded", "Loading", "Error", "Missing"],
      imagesStatusOptions: ["Loaded", "Loading", "Error", "Missing", "N/A"]
    };
  },
  computed: mapState(["collections", "posts", "settings"]),
  methods: {
    getPosts() {
      return this.posts.postsList.map(p => {
        return {
          id: p.id,
          name: p.name,
          postStatus: p.postStatus,
          recordsStatus: p.recordsKey ? p.recordsStatus || "Loaded" : "Missing",
          imagesStatus: !this.collections.collectionsList.find(coll => coll.id === p.collection).imagePathHeader
            ? "N/A"
            : !!p.imagesKeys && p.imagesKeys.length > 0
            ? p.imagesStatus || "Loaded"
            : "Missing",
          collectionName: this.collections.collectionsList.find(coll => coll.id === p.collection).name,
          ...p.metadata
        };
      });
    },
    getPostColumns() {
      let cols = [
        {
          text: "Name",
          value: "name"
        },
        {
          text: "Status",
          value: "postStatus",
          filter: value => {
            return this.postStatusFilter.length === 0 || this.postStatusFilter.includes(value);
          }
        },
        {
          text: "Records",
          value: "recordsStatus",
          align: "center",
          filter: value => {
            return this.recordsStatusFilter.length === 0 || this.recordsStatusFilter.includes(value);
          }
        },
        {
          text: "Images",
          value: "imagesStatus",
          align: "center",
          filter: value => {
            return this.imagesStatusFilter.length === 0 || this.imagesStatusFilter.includes(value);
          }
        },
        {
          text: "Collection",
          value: "collectionName"
        }
      ];
      cols.push(...this.settings.settings.postMetadata.map(pf => getMetadataColumn(pf)));
      cols.push({ text: "", value: "icon", align: "right" });
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
/* freeze the first column (for post name) since we don't know if this will be an overly wide table with custom fields */
.postsTable >>> table > tbody > tr > td:nth-child(1),
.postsTable >>> table > thead > tr > th:nth-child(1) {
  left: 0;
}
.postsTable >>> table > thead > tr > th:nth-child(1) {
  position: sticky !important;
  position: -webkit-sticky !important;
  /* z-index: 9999; */
  /* background: white; */
}
.postsTable >>> table > tbody > tr > td:nth-child(1) {
  position: sticky !important;
  position: -webkit-sticky !important;
  /* z-index: 9998; */
  background: white;
}
.postsTable >>> table > tbody > tr > td:nth-child(1):hover {
  background-color: #efefef;
}
.postsTable >>> table > tbody > tr > td {
  padding: 0 8px;
}
.postsTable >>> thead .text-start {
  vertical-align: top;
  text-align: left;
  padding-left: 8px;
}
.postsTable >>> thead .sortable {
  vertical-align: top;
  text-align: left;
  padding-left: 8px;
}
.postsTable >>> .table-header-group {
  vertical-align: top;
  text-align: left;
  padding-left: 8px;
}
</style>
