<template>
  <v-col class="search-detail" cols="12">
    <v-btn text class="pl-0 ml-0 primary--text" @click="$router.go(-1)"><v-icon>mdi-chevron-left</v-icon> Back</v-btn>
    <h1>Record detail for</h1>
    <h2>{{ search.searchResult.person.name }} in {{ search.searchResult.collectionName }}</h2>
    <v-row class="recordDetail d-flex">
      <!--image-->
      <v-col cols="12" md="3" class="" v-if="thumbURL">
        <div class="" no-gutters>
          <router-link
            :to="{
              name: 'image',
              params: {
                societyId: search.searchResult.societyId,
                pid: search.searchResult.post,
                path: search.searchResult.imagePath
              }
            }"
          >
            <v-img :contain="true" :src="thumbURL" max-width="160"></v-img>
          </router-link>
          <v-btn
            text
            class="primary--text mx-5"
            :to="{
              name: 'image',
              params: {
                societyId: search.searchResult.societyId,
                pid: search.searchResult.post,
                path: search.searchResult.imagePath
              }
            }"
          >
            Larger image<v-icon right>mdi-chevron-right</v-icon>
          </v-btn>
        </div>
      </v-col>
      <!--individual transcript-->
      <v-col cols="12" md="8" v-if="search.searchResult.private">
        <p>
          This record is available to members of the society.
          <a v-if="search.searchResult.loginURL" :href="search.searchResult.loginURL">Click here to become a member</a>
        </p>
      </v-col>
      <v-col cols="12" md="8" v-else>
        <v-row v-if="search.searchResult.person.role !== 'principal'" no-gutters>
          <v-col>
            <div>{{ nameLabel }}: {{ search.searchResult.person.name }}</div>
            <div>{{ roleLabel }}: {{ search.searchResult.person.role }}</div>
          </v-col>
        </v-row>
        <v-row no-gutters>
          <v-col>
            <h4>Collection: {{ search.searchResult.collectionName }}</h4>
            <h4>{{ search.searchResult.collectionLocation ? "In " + search.searchResult.collectionLocation : "" }}</h4>
          </v-col>
        </v-row>
        <v-row no-gutters class="mt-5 d-flex">
          <v-col cols="12">
            <h4 class="recordDetailSectionHead">Record Details</h4>
          </v-col>
        </v-row>
        <v-row v-for="(lv, $ix) in search.searchResult.record" :key="$ix">
          <v-col cols="3" class="d-flex justify-right flex-column recordDetailRow">{{ lv.label }}:</v-col>
          <v-col cols="9" class="recordDetailRow" v-html="sanitize(lv.value)"></v-col>
        </v-row>
      </v-col>
    </v-row>
    <!--household transcript table-->
    <v-row v-if="search.searchResult.household && search.searchResult.household.length > 0">
      <v-col>
        <h4 class="recordDetailSectionHead">Household Details</h4>
        <v-data-table
          :headers="householdHeaders"
          :items="householdRecords"
          :disable-pagination="true"
          dense
          v-columns-resizable
          hide-default-footer
        >
        </v-data-table>
      </v-col>
    </v-row>
    <!--citation-->
    <v-row v-if="search.searchResult.citation">
      <v-col cols="12">
        <v-card class="pa-5">
          <h4 class="recordDetailSectionHead mb-3">How to cite this record</h4>
          <div v-html="sanitize(search.searchResult.citation)"></div>
        </v-card>
      </v-col>
    </v-row>
  </v-col>
</template>

<script>
import { mapState } from "vuex";
import store from "@/store";
import Server from "@/services/Server.js";
// import draggable from "vuedraggable";

const surnameFirst = typeof window.ourroots.surnameFirst === "string" && window.ourroots.surnameFirst.length > 0;

export default {
  // components: { draggable },
  beforeRouteEnter(routeTo, routeFrom, next) {
    store
      .dispatch("searchGetResult", {id: routeTo.params.rid, surnameFirst: surnameFirst})
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  created() {
    console.log("searchResult", this.search.searchResult);
    // get image path
    if (this.search.searchResult.imagePath) {
      Server.postsGetImage(
        this.search.searchResult.societyId,
        this.search.searchResult.post,
        this.search.searchResult.imagePath,
        true
      ).then(result => {
        this.thumbURL = result.data.url;
      });
    }
    // get household headers
    if (this.search.searchResult.household && this.search.searchResult.household.length > 0) {
      for (let record of this.search.searchResult.household) {
        for (let lv of record) {
          if (this.householdHeaders.findIndex(header => header.text === lv.label) === -1) {
            this.householdHeaders.push({ text: lv.label, value: lv.label });
          }
        }
      }
    }
    console.log("householdHeaders", this.householdHeaders);
    this.householdRecords = (this.search.searchResult.household || []).map(record => {
      let result = {};
      for (let header of this.householdHeaders) {
        result[header.text] = this.getRecordValue(record, header.text);
      }
      console.log(result);
      return result;
    });
    console.log("householdRecords", this.householdRecords);
  },
  data() {
    return {
      thumbWidth: "200px",
      thumbURL: "",
      householdHeaders: [],
      householdRecords: []
    };
  },
  computed: {
    nameLabel() {
      return this.search.searchResult.collectionType === "Records" ? "Name" : "Title";
    },
    roleLabel() {
      return this.search.searchResult.collectionType === "Records" ? "Role" : "Author";
    },
    ...mapState(["search"])
  },
  methods: {
    sanitize(value) {
      return this.$sanitize(value);
    },
    getRecordValue(record, header) {
      let lv = record.find(lv => lv.label === header);
      return lv ? lv.value : "";
    }
  }
};
</script>

<style scoped>
.recordDetail {
  border-top: solid 1px #cccccc;
  margin-top: 12px;
}
.recordDetailSectionHead {
  text-transform: uppercase;
}
.recordDetailRow {
  padding-top: 0;
  padding-bottom: 0;
}
</style>
