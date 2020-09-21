<template>
  <div class="search-detail">
    <v-row>
      <v-col>
        <h1>Record detail</h1>
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="8">
        <v-row v-if="search.searchResult.person.role !== 'principal'">
          <v-col>
            <div>Name: {{ search.searchResult.person.name }}</div>
            <div>Role: {{ search.searchResult.person.role }}</div>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <h4>
              {{ search.searchResult.collectionName }}
              {{ search.searchResult.collectionLocation ? "in " + search.searchResult.collectionLocation : "" }}
            </h4>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <div v-for="(lv, $ix) in search.searchResult.record" :key="$ix">{{ lv.label }}: {{ lv.value }}</div>
          </v-col>
        </v-row>
      </v-col>
      <v-col>
        <div v-if="thumbURL">
          <router-link
            :to="{
              name: 'image',
              params: {
                pid: search.searchResult.post,
                path: search.searchResult.imagePath
              }
            }"
          >
            <v-img :contain="true" :src="thumbURL" max-width="160"></v-img>
          </router-link>
          <router-link
            :to="{
              name: 'image',
              params: { pid: search.searchResult.post, path: search.searchResult.imagePath }
            }"
          >
            Larger image
          </router-link>
        </div>
      </v-col>
    </v-row>
    <v-row v-if="search.searchResult.citation">
      <v-col>
        <h4>Citation</h4>
        <div>{{ search.searchResult.citation }}</div>
      </v-col>
    </v-row>
    <v-row v-if="search.searchResult.household && search.searchResult.household.length > 0">
      <v-col>
        <h3>Household</h3>
        <table>
          <thead>
            <td v-for="(header, ix) in householdHeaders" :key="ix">
              {{ header }}
            </td>
          </thead>
          <tbody>
            <tr v-for="(record, i) in search.searchResult.household" :key="i">
              <td v-for="(header, j) in householdHeaders" :key="j">
                {{ getRecordValue(record, header) }}
              </td>
            </tr>
          </tbody>
        </table>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { mapState } from "vuex";
import store from "@/store";
import Server from "@/services/Server.js";

export default {
  beforeRouteEnter(routeTo, routeFrom, next) {
    store
      .dispatch("searchGetResult", routeTo.params.rid)
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
      Server.postsGetImage(this.search.searchResult.post, this.search.searchResult.imagePath, true).then(result => {
        this.thumbURL = result.data.url;
      });
    }
    // get household headers
    if (this.search.searchResult.household && this.search.searchResult.household.length > 0) {
      for (let record of this.search.searchResult.household) {
        for (let lv of record) {
          if (!this.householdHeaders.includes(lv.label)) {
            this.householdHeaders.push(lv.label);
          }
        }
      }
    }
    console.log("householdHeaders", this.householdHeaders);
  },
  data() {
    return {
      thumbURL: "",
      householdHeaders: []
    };
  },
  computed: mapState(["search"]),
  methods: {
    getRecordValue(record, header) {
      let lv = record.find(lv => lv.label === header);
      return lv ? lv.value : "";
    }
  }
};
</script>

<style scoped></style>
