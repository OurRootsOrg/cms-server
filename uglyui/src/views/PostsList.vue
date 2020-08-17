<template>
  <v-container class="posts-list">
    <h1>Posts</h1>
    <v-btn small color="primary" class="mt-2" to="/posts/create">
      Create a new post
    </v-btn>
    <v-row class="d-flex justify-end">
      <!-- <v-col class="mt-1">
        <Tabulator
          :data="getPosts()"
          :columns="getPostColumns()"
          layout="fitColumns"
          :header-sort="true"
          :selectable="true"
          :resizable-columns="true"
          @rowClicked="rowClicked"
        />
      </v-col> -->
      <v-col cols="12" md="2">  
        <v-select
          v-model="recordsStatusFilter"
          :items="recordsStatusOptions"
          label="Status"
          multiple
        ></v-select>
      </v-col>
      <v-col cols="12" md="2">
        <v-select
          v-model="hasDataFilter"
          :items="hasDataOptions"
          label="Has data?"
          multiple
        ></v-select>
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="Search"
          single-line
          hide-details
        ></v-text-field>   
      </v-col>      
    </v-row>
    <v-row>  
      <v-col cols="12">
        <v-data-table
        :items="getPosts()"
        :headers="getPostColumns()"
        sortable
        sort-by='name'
        :search="search"
        :footer-props="{
          'items-per-page-options': [10, 25, 50]
        }"
        :items-per-page="25"
        @click:row="rowClicked"
        dense
        class="rowHover postsTable"
        >
        <!-- <template v-slot:body.prepend>
          I took the filters out of here because it throws the widths all out of whack
        </template> -->
          <template v-slot:item.hasData="{ item }">
            <v-icon v-if="item.hasData" class="green--text">mdi-checkbox-marked</v-icon>
            <v-icon v-else class="red--text">mdi-close-circle</v-icon>
          </template>
          <template v-slot:item.icon="{ item }">
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
// import Tabulator from "../components/Tabulator";

export default {
  // components: { Tabulator },
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
      cols: [
        { text: "Name", value: "name" },
        { text: "Status", value: "recordsStatus"},
        { text: "Has Data", value: "hasData", align:"center"},
        { text: "Collection", value: "collectionName"}, 
        // { text: "Source", value: "source" },
        // { text: "Date Transcribed", value: "transcriptionDate"},               
        { text: "", value: "icon", align:"right" }        
      ],
      search: "",
      status: "",
      recordsStatusFilter: [],
      hasDataFilter: [],
      recordsStatusOptions: ["Published", "Publishing", "Draft"],
      hasDataOptions: [{value: true, text: "Has data"}, {value: false, text: "No data"}]
    }
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
      //Tabulator:v-data-table translation is title:text and field:value (rename "title" as "text" and "field" as "value")
      let cols = [
        {
          text: "Name",
          value: "name",
          headerFilter: "input",
          sorter: "string"
        },
        {
          text: "Status",
          value: "recordsStatus",
          filter: value => {
            // console.log("value = ", value, "recordsStatusFilter = ", this.recordsStatusFilter)
              if (this.recordsStatusFilter.length === 0) return true;
              return this.recordsStatusFilter.includes(value);
            },
          headerFilter: "select",
          headerFilterParams: {
            values: true
          },
          sorter: "string"
        },
        {
          text: "Has Data",
          value: "hasData",
          filter: value => {
            // console.log("value = ", value, "hasDataFilter = ", this.hasDataFilter);
              if (this.hasDataFilter.length === 0) return true;
              return this.hasDataFilter.includes(value);
          },
          align: "center", //used to be hozAlign: 'center'
          formatter: "tickCross",
          headerFilter: "tickCross",
          sorter: "boolean",
        },
        {
          text: "Collection",
          value: "collectionName",
          headerFilter: "input",
          sorter: "string"
        },
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

.postsTable >>> table > tbody > tr > td:nth-child(1),
.postsTable >>> table > thead > tr > th:nth-child(1) {
  left: 0;
}
.postsTable >>> table > tbody > tr > td:nth-child(2),
.postsTable >>> table > thead > tr > th:nth-child(2) {
  left: 50px;
}
.postsTable >>> table > tbody > tr > td:nth-child(3),
.postsTable >>> table > thead > tr > th:nth-child(3) {
  left: 140px;
}
.postsTable >>> table > tbody > tr > td:nth-child(4),
.postsTable >>> table > thead > tr > th:nth-child(4) {
  left: 260px;
}
.postsTable >>> table > thead > tr > th:nth-child(1)
/* .postsTable >>> table > thead > tr > th:nth-child(2) */
 {
  position: sticky !important;
  position: -webkit-sticky !important;
  /* z-index: 9999; */
  background: white;
}
.postsTable >>> table > tbody > tr > td:nth-child(1)
/* .postsTable >>> table > tbody > tr > td:nth-child(2) */
 {
  position: sticky !important;
  position: -webkit-sticky !important;
  /* z-index: 9998; */
  background: white;
}
.postsTable >>> table > tbody > tr > td:nth-child(1):hover {
  background-color:#efefef;
}

.postsTable >>>  table > tbody > tr > td {
  padding: 0 8px;
}
.postsTable >>> thead .text-start {
  vertical-align: top;
  text-align: left;
  padding-left:8px;
}
.postsTable >>> thead .sortable {
  vertical-align: top;
  text-align: left;
  padding-left:8px;
}
.postsTable >>> .table-header-group {
  vertical-align: top;
  text-align:left;
  padding-left:8px;
}
</style>
