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
            <v-img :contain="true" :src="thumbURL" :max-width="thumbWidth"></v-img>
          </router-link>
        </div>
        <router-link
          :to="{
            name: 'image',
            params: { pid: search.searchResult.post, path: search.searchResult.imagePath }
          }"
        >
          Larger image
        </router-link>
      </v-col>
    </v-row>
    <v-row v-if="search.searchResult.citation">
      <v-col>
        <h4>Citation</h4>
        <div>{{ search.searchResult.citation }}</div>
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
    if (this.search.searchResult.imagePath) {
      Server.postsGetImage(this.search.searchResult.post, this.search.searchResult.imagePath, 0, this.thumbWidth).then(
        result => {
          console.log("searchResult", this.search.searchResult);
          this.thumbURL = result.data.url;
          console.log("thumbURL", this.thumbURL);
        }
      );
    }
  },
  data() {
    return {
      thumbWidth: 160,
      thumbURL: ""
    };
  },
  computed: mapState(["search"])
};
</script>

<style scoped></style>
