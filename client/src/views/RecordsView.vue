<template>
  <div class="records-view">
    <v-row>
      <v-col>
        <h1>Record detail</h1>
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="8">
        <v-row>
          <v-col>
            <div v-for="(label, $ix) in records.record.labels.filter(l => l.label)" :key="$ix">
              {{ label.label }}: <span v-html="sanitize(getRecordValue(records.record, label.header))"></span>
            </div>
            <h4 v-if="records.record.labels.filter(l => !l.label).length > 0" class="omitted-labels">
              Omitted from search results page
              <v-tooltip bottom maxWidth="600px">
                <template v-slot:activator="{ on, attrs }">
                  <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
                </template>
                <span>The record detail page label for these fields is empty on the Collection page</span>
              </v-tooltip>
            </h4>
            <div v-for="label in records.record.labels.filter(l => !l.label)" :key="label.header">
              {{ label.header }}: <span v-html="sanitize(getRecordValue(records.record, label.header))"></span>
            </div>
          </v-col>
        </v-row>
      </v-col>
      <v-col cols="4">
        <div v-if="thumbURL">
          <router-link
            :to="{
              name: 'image',
              params: {
                pid: records.record.post,
                path: records.record.imagePath
              }
            }"
          >
            <v-img :contain="true" :src="thumbURL" max-width="160"></v-img>
          </router-link>
          <router-link
            :to="{
              name: 'image',
              params: { pid: records.record.post, path: records.record.imagePath }
            }"
          >
            Larger image
          </router-link>
        </div>
      </v-col>
    </v-row>
    <v-row v-if="records.record.citation">
      <v-col>
        <h4>Citation</h4>
        <div v-html="sanitize(records.record.citation)"></div>
      </v-col>
    </v-row>
    <v-row v-if="records.record.household && records.record.household.length > 0">
      <v-col>
        <h3>Household</h3>
        <table>
          <thead>
            <td v-for="(label, ix) in records.record.labels.filter(l => l.label)" :key="ix">
              {{ label.label }}
            </td>
          </thead>
          <tbody>
            <tr v-for="(record, i) in records.record.household" :key="i">
              <td
                v-for="(label, j) in records.record.labels.filter(l => l.label)"
                :key="j"
                v-html="sanitize(getRecordValue(record, label.header))"
              ></td>
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

function getContent(rid, next) {
  store
    .dispatch("recordsGetDetail", rid)
    .then(() => {
      next();
    })
    .catch(() => {
      next("/");
    });
}

export default {
  beforeRouteEnter(routeTo, routeFrom, next) {
    console.log("recordsView.beforeRouteEnter");
    getContent(routeTo.params.rid, next);
  },
  beforeRouteUpdate(routeTo, routeFrom, next) {
    console.log("recordsView.beforeRouteUpdate");
    getContent(routeTo.params.rid, next);
  },
  created() {
    console.log("recordsView", this.records.record);
    // get image path
    if (this.records.record.imagePath) {
      Server.postsGetImage(
        store.getters.currentSocietyId,
        this.records.record.post,
        this.records.record.imagePath,
        true
      ).then(result => {
        this.thumbURL = result.data.url;
      });
    }
  },
  data() {
    return {
      thumbURL: ""
    };
  },
  computed: mapState(["records"]),
  methods: {
    sanitize(value) {
      return this.$sanitize(value);
    },
    getRecordValue(record, header) {
      return record.data[header] || "";
    }
  }
};
</script>

<style scoped>
.records-view {
  width: 100%;
}
.omitted-labels {
  margin-top: 10px;
}
</style>
