<template>
  <v-container class="posts-create">
    <h1>{{ post.id ? "Edit" : "Create" }} Record set</h1>
    <v-form @submit.prevent="save">
      <h3>Give your record set a name</h3>
      <v-text-field
        label="Record-set Name"
        v-model="post.name"
        type="text"
        placeholder="Name"
        :class="{ error: $v.post.name.$error }"
        @blur="touch('name')"
      ></v-text-field>
      <template v-if="$v.post.name.$error">
        <p v-if="!$v.post.name.required" class="errorMessage">
          Name is required.
        </p>
      </template>
      <div v-if="post.id">
        <h3>Collection</h3>
        <p>
          <router-link :to="{ name: 'collection-edit', params: { cid: collections.collection.id } }">{{
            collections.collection.name
          }}</router-link>
        </p>
      </div>
      <div v-else>
        <h3>Select a collection</h3>
        <v-select
          label="Collection"
          :items="collections.collectionsList"
          item-text="name"
          item-value="id"
          v-model="post.collection"
          :class="{ error: $v.post.collection.$error }"
          @input="touch('collection')"
        ></v-select>
        <template v-if="$v.post.collection.$error">
          <p v-if="!$v.post.collection.required" class="errorMessage">
            Collection is required.
          </p>
        </template>
      </div>
      <div v-if="post.id" class="postStatusWrapper">
        <h3>Status</h3>
        <ul>
          <li>
            Records:
            <span v-if="post.recordsError"><strong>error</strong> - {{ cleanError(post.recordsError) }}</span>
            <span v-else-if="!post.recordsKey">Missing</span>
            <span v-else>
              <router-link :to="{ name: 'downloads', params: { key: post.recordsKey } }">
                {{ post.recordsStatus || "Loaded" }}
              </router-link>
            </span>
          </li>
          <li v-if="this.collections.collection.imagePathHeader">
            Images:
            <span v-if="post.imagesError"><strong>error</strong> - {{ cleanError(post.imagesError) }}</span>
            <span v-else-if="!post.imagesKeys || post.imagesKeys.length === 0">Missing</span>
            <span v-else>
              <router-link :to="{ name: 'downloads', params: { key: post.imagesKeys[0] } }">
                {{ post.imagesStatus || "Loaded" }}
              </router-link>
            </span>
          </li>
          <li>
            Visibility:
            <span v-if="post.postError"><strong>error</strong> - {{ cleanError(post.postError) }}</span>
            <span v-else>{{ post.postStatus }}</span>
          </li>
        </ul>
      </div>
      <div v-if="societySummaries.societySummary.postMetadata.length > 0">
        <h3>
          Custom fields (metadata) for this record set
          <v-tooltip bottom>
            <template v-slot:activator="{ on, attrs }">
              <v-icon small v-bind="attrs" v-on="on">mdi-information</v-icon>
            </template>
            <span
              >Information specific to <em>this particular record set</em> such as the transcription date, translator,
              etc. which might be different from other record sets in this collection</span
            >
          </v-tooltip>
        </h3>
        <v-row no-gutter v-for="(item, index) in societySummaries.societySummary.postMetadata" :key="index">
          <v-col cols="12" md="5" class="py-0" v-if="item.type === 'string'">
            <v-text-field
              :placeholder="item.tooltip"
              :label="item.name"
              v-model="post.metadata[item.name]"
              @change="touch('metadata')"
            ></v-text-field>
          </v-col>
          <v-col cols="12" md="5" class="py-0" v-if="item.type === 'boolean'">
            <v-checkbox :label="item.name" v-model="post.metadata[item.name]" @change="touch('metadata')">
              <v-tooltip slot="append" bottom>
                <template v-slot:activator="{ on }">
                  <v-icon v-on="on" small>mdi-information-outline</v-icon>
                </template>
                <span>{{ item.tooltip }}</span>
              </v-tooltip>
            </v-checkbox>
          </v-col>
          <v-col cols="12" md="5" class="py-0" v-if="item.type === 'number'">
            <v-text-field
              :placeholder="item.tooltip"
              :label="item.name"
              type="number"
              v-model="post.metadata[item.name]"
              @change="touch('metadata')"
            ></v-text-field>
          </v-col>
          <v-col cols="12" md="5" class="py-0" v-if="item.type === 'date'">
            <v-menu
              v-model="showPicker"
              :close-on-content-click="false"
              :nudge-right="40"
              transition="scale-transition"
              offset-y
              min-width="290px"
              max-width="290px"
            >
              <template v-slot:activator="{ on }">
                <v-text-field
                  :placeholder="item.tooltip"
                  :label="item.name"
                  v-model="post.metadata[item.name]"
                  prepend-icon="mdi-calendar-range"
                  readonly
                  v-on="on"
                ></v-text-field>
              </template>
              <v-date-picker
                v-model="post.metadata[item.name]"
                @input="showPicker = false"
                @change="touch('metadata')"
              ></v-date-picker>
            </v-menu>
          </v-col>
        </v-row>
      </div>
      <p v-if="$v.$anyError" class="errorMessage">
        Please fill out the required field(s).
      </p>

      <v-row>
        <v-col class="d-flex">
          <v-btn type="submit" color="primary" :disabled="$v.$anyError || !$v.$anyDirty">Save </v-btn>
          <v-btn
            v-if="isPublishable"
            @click="publish"
            color="primary"
            title="Publish the record set to make it searchable"
            class="ml-4"
            >Publish Record set</v-btn
          >
          <v-btn
            v-if="isUnpublishable"
            @click="unpublish"
            color="primary"
            title="Unpublish the record set to remove it from the index"
            class="ml-4"
            >Unpublish Record set</v-btn
          >
          <v-dialog v-if="isRecordsImportable" v-model="importRecordsDlg" persistent max-width="320">
            <template v-slot:activator="{ on, attrs }">
              <v-btn color="primary" v-bind="attrs" v-on="on" class="ml-4">
                {{ post.recordsKey ? "Replace data" : "Import data" }}
              </v-btn>
            </template>
            <v-card>
              <v-card-title class="headline">Select file to import</v-card-title>
              <v-card-text>
                <file-upload
                  class="btn btn-primary"
                  post-action="/"
                  extensions="csv"
                  accept="text/csv"
                  :headers="{ 'Content-Type': 'text/csv' }"
                  :multiple="false"
                  :size="1024 * 1024 * 1024"
                  v-model="recordFiles"
                  @input-filter="recordsInputFilter"
                  ref="uploadrecords"
                >
                  <v-btn class="btn primary" :disabled="recordsUploading">Select CSV file</v-btn>
                </file-upload>
                <div v-if="recordsError" class="errorMessage">{{ recordsError }}</div>
                <ul>
                  <li v-for="file in recordFiles" :key="file.id">
                    <span>{{ file.name }}</span> - <span>{{ file.size | formatSize }}</span>
                    <span v-if="file.error"> - {{ file.error }}</span>
                    <span v-else-if="file.success"> - success</span>
                    <span v-else-if="file.active"> - uploading</span>
                    <span v-else></span>
                  </li>
                </ul>
              </v-card-text>
              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn text v-if="!recordFiles.find(f => f.success)" @click="cancelRecordsUpload($refs.uploadrecords)"
                  >Cancel</v-btn
                >
                <v-btn
                  color="primary"
                  text
                  v-if="!recordFiles.find(f => f.success)"
                  :disabled="!$refs.uploadrecords || $refs.uploadrecords.active"
                  @click="startRecordsUpload($refs.uploadrecords)"
                  >Start Upload</v-btn
                >
                <v-btn color="primary" text v-if="recordFiles.find(f => f.success)" @click="endRecordsUpload()"
                  >Upload Successful!</v-btn
                >
              </v-card-actions>
            </v-card>
          </v-dialog>

          <v-dialog v-if="isImagesImportable" v-model="importImagesDlg" persistent max-width="320">
            <template v-slot:activator="{ on, attrs }">
              <v-btn color="primary" v-bind="attrs" v-on="on" class="ml-4">
                {{ !!post.imagesKeys && post.imagesKeys.length > 0 ? "Replace images" : "Import images" }}
              </v-btn>
            </template>
            <v-card>
              <v-card-title class="headline">Select file to import</v-card-title>
              <v-card-text>
                <file-upload
                  class="btn btn-primary"
                  post-action="/"
                  extensions="zip"
                  accept="application/zip"
                  :headers="{ 'Content-Type': 'application/zip' }"
                  :multiple="false"
                  :size="1024 * 1024 * 1024 * 10"
                  v-model="imageFiles"
                  @input-filter="imagesInputFilter"
                  ref="uploadimages"
                >
                  <v-btn class="btn primary" :disabled="imagesUploading">Select ZIP file</v-btn>
                </file-upload>
                <div v-if="imagesError" class="errorMessage">{{ imagesError }}</div>
                <ul>
                  <li v-for="file in imageFiles" :key="file.id">
                    <span>{{ file.name }}</span> - <span>{{ file.size | formatSize }}</span>
                    <span v-if="file.error"> - {{ file.error }}</span>
                    <span v-else-if="file.success"> - success</span>
                    <span v-else-if="file.active"> - uploading</span>
                    <span v-else></span>
                  </li>
                </ul>
              </v-card-text>
              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn text v-if="!imageFiles.find(f => f.success)" @click="cancelImagesUpload($refs.uploadimages)"
                  >Cancel</v-btn
                >
                <v-btn
                  color="primary"
                  text
                  v-if="!imageFiles.find(f => f.success)"
                  :disabled="!$refs.uploadimages || $refs.uploadimages.active"
                  @click="startImagesUpload($refs.uploadimages)"
                  >Start Upload</v-btn
                >
                <v-btn color="primary" text v-if="imageFiles.find(f => f.success)" @click="endImagesUpload()"
                  >Upload Successful!</v-btn
                >
              </v-card-actions>
            </v-card>
          </v-dialog>

          <v-spacer></v-spacer>

          <v-btn :disabled="!isDeletable" @click="del" class="warning">Delete Record set</v-btn>
        </v-col>
      </v-row>
    </v-form>

    <v-row class="pt-5" v-if="post.id && post.recordsKey && post.recordsStatus === ''">
      <v-col>
        <h3 class="pl-1">Records</h3>
        <v-data-table
          :items="
            records.recordsList.map(r => {
              return { __id: r.id, ...r.data };
            })
          "
          :headers="getRecordColumns()"
          dense
          @click:row="rowClicked"
          sortable
          :footer-props="{
            'items-per-page-options': [10, 25, 50]
          }"
          :items-per-page="25"
          v-columns-resizable
          class="rowHover"
        >
          <template v-slot:[`item.icon`]="{ item }">
            <v-btn icon small :to="{ name: 'records-view', params: { rid: item.__id } }">
              <v-icon right>mdi-chevron-right</v-icon>
            </v-btn>
          </template>
        </v-data-table>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import { required } from "vuelidate/lib/validators";
import FileUpload from "vue-upload-component";
import Server from "@/services/Server.js";
import NProgress from "nprogress";
import lodash from "lodash";

function getContent(pid, next) {
  let routes = [];
  if (pid) {
    routes.push(store.dispatch("postsGetOne", pid));
    routes.push(store.dispatch("recordsGetForPost", pid));
  } else {
    routes.push(store.dispatch("collectionsGetAll"));
  }
  Promise.all(routes)
    .then(() => {
      if (pid) {
        store.dispatch("collectionsGetOne", store.state.posts.post.collection).then(() => {
          next();
        });
      } else {
        next();
      }
    })
    .catch(() => {
      next("/");
    });
}

function setup() {
  this.post = {
    ...this.posts.post
  };
}

export default {
  components: { FileUpload },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    console.log("postsCreateEdit.beforeRouteEnter");
    getContent(routeTo.params.pid, next);
  },
  beforeRouteUpdate: function(routeTo, routeFrom, next) {
    console.log("postsCreateEdit.beforeRouteUpdate");
    getContent(routeTo.params.pid, next);
  },
  created() {
    if (this.$route.params && this.$route.params.pid) {
      setup.bind(this)();
    }
  },
  data() {
    return {
      post: {
        id: null,
        name: null,
        collection: null,
        postStatus: "Draft",
        recordsStatus: "",
        imagesStatus: "",
        recordsKey: null,
        metadata: {}
      },
      showPicker: false,
      importImagesDlg: false,
      imageFiles: [],
      imagesError: "",
      imagesUploading: false,
      imagesPostRequestResultData: null,
      importRecordsDlg: false,
      recordFiles: [],
      recordsError: "",
      recordsUploading: false,
      recordsPostRequestResultData: null
    };
  },
  computed: {
    isRecordsImportable() {
      return (
        this.post.id &&
        (this.post.recordsStatus === "" || this.post.recordsStatus === "Error") &&
        (this.post.postStatus === "Draft" || this.post.postStatus === "Error")
      );
    },
    isImagesImportable() {
      return (
        this.post.id &&
        this.collections.collection.imagePathHeader &&
        (this.post.imagesStatus === "" || this.post.imagesStatus === "Error") &&
        (this.post.postStatus === "Draft" || this.post.postStatus === "Error")
      );
    },
    isDeletable() {
      return (
        this.post.id &&
        (this.post.recordsStatus === "" || this.post.recordsStatus === "Error") &&
        (this.post.imagesStatus === "" || this.post.imagesStatus === "Error") &&
        (this.post.postStatus === "Draft" || this.post.postStatus === "Error")
      );
    },
    isPublishable() {
      return (
        this.post.id &&
        this.post.recordsKey &&
        this.post.recordsStatus === "" &&
        this.post.imagesStatus === "" &&
        (this.post.postStatus === "Draft" || this.post.postStatus === "Error")
      );
    },
    isUnpublishable() {
      return (
        this.post.id &&
        this.post.recordsKey &&
        this.post.recordsStatus === "" &&
        this.post.imagesStatus === "" &&
        this.post.postStatus === "Published"
      );
    },
    ...mapState(["collections", "posts", "records", "societySummaries", "users"])
  },
  validations: {
    post: {
      name: { required },
      collection: { required },
      metadata: {}
    }
  },
  methods: {
    touch(attr) {
      if (this.$v.post[attr].$dirty) {
        return;
      }
      if (!this.post.id || attr === "metadata" || !lodash.isEqual(this.post[attr], this.posts.post[attr])) {
        this.$v.post[attr].$touch();
      }
    },
    cleanError(err) {
      for (let errPrefix of ["[ERROR] ", "Errors:", "Error OTHER:", "Unknown error:"]) {
        if (err.startsWith(errPrefix)) {
          err = err.substr(errPrefix.length).trim();
        }
      }
      return err;
    },
    metadataEdited() {
      this.touch("metadata");
    },
    getRecordColumns() {
      let cols = [];
      for (let f of this.collections.collection.fields) {
        cols.push({ text: f.header, value: f.header });
        if (
          this.collections.collection.mappings.find(
            m => m.header === f.header && (m.ixField.endsWith("Date") || m.ixField.endsWith("Place"))
          )
        ) {
          cols.push({ text: f.header + "_std", value: f.header + "_std" });
        }
      }
      cols.push({ text: "", value: "icon", align: "right" });
      return cols;
    },
    rowClicked(record) {
      console.log("rowClicked", record);
      this.$router.push({
        name: "records-view",
        params: { rid: record.__id }
      });
    },
    getPostFromForm() {
      let post = Object.assign({}, this.post);
      post.collection = +post.collection; // convert to a number
      return post;
    },
    save() {
      let post = this.getPostFromForm();
      this.update(post);
    },
    publish() {
      let post = this.getPostFromForm();
      post.postStatus = "Publication Requested";
      this.update(post);
    },
    unpublish() {
      let post = this.getPostFromForm();
      post.postStatus = "Unpublication Requested";
      this.update(post);
    },
    imagesInputFilter(newFile, oldFile, prevent) {
      this.imagesError = "";
      if (newFile && !oldFile) {
        if (!newFile.name || !newFile.name.endsWith(".zip")) {
          this.imagesError = "Please put your images into a ZIP file; You need to select a file ending in .zip";
          return prevent();
        }
        return Server.contentPostRequest(store.getters.currentSocietyId, "application/zip").then(result => {
          console.log("contentPostRequest", result.data);
          this.imagesPostRequestResultData = result.data;
          newFile.putAction = result.data.signedURL;
        });
      }
    },
    cancelImagesUpload(upload) {
      upload.active = false;
      this.imagesUploading = false;
      this.imageFiles = [];
      this.importImagesDlg = false;
    },
    startImagesUpload(upload) {
      upload.active = true;
      this.imagesUploading = true;
    },
    endImagesUpload() {
      this.imageFiles = [];
      this.importImagesDlg = false;
      this.imagesUploading = false;
      this.$v.$touch();
      NProgress.start();
      // re-read post to avoid 409 conflict
      this.$store
        .dispatch("postsGetOne", this.post.id)
        .then(post => {
          post.imagesKeys = [this.imagesPostRequestResultData.key];
          console.log("emdImagesUpload", post, this.imagesPostRequestResultData);
          this.$store
            .dispatch("postsUpdate", post)
            .then(() => {
              setup.bind(this)();
              this.$v.$reset();
            })
            .finally(() => {
              NProgress.done();
            });
        })
        .catch(() => {
          NProgress.done();
        });
    },
    recordsInputFilter(newFile, oldFile, prevent) {
      this.recordsError = "";
      if (newFile && !oldFile) {
        if (!newFile.name || !newFile.name.endsWith(".csv")) {
          this.recordsError = "Please save your data as a CSV file; You need to select a file ending in .csv";
          return prevent();
        }
        return Server.contentPostRequest(store.getters.currentSocietyId, "text/csv").then(result => {
          console.log("contentPostRequest", result.data);
          this.recordsPostRequestResultData = result.data;
          newFile.putAction = result.data.signedURL;
        });
      }
    },
    cancelRecordsUpload(upload) {
      upload.active = false;
      this.recordsUploading = false;
      this.recordFiles = [];
      this.importRecordsDlg = false;
    },
    startRecordsUpload(upload) {
      upload.active = true;
      this.recordsUploading = true;
    },
    endRecordsUpload() {
      this.recordFiles = [];
      this.importRecordsDlg = false;
      this.recordsUploading = false;
      this.$v.$touch();
      NProgress.start();
      // re-read post to avoid 409 conflict
      this.$store
        .dispatch("postsGetOne", this.post.id)
        .then(post => {
          post.recordsKey = this.recordsPostRequestResultData.key;
          console.log("endRecordsUpload", post, this.recordsPostRequestResultData);
          this.$store
            .dispatch("postsUpdate", post)
            .then(() => {
              setup.bind(this)();
              this.$v.$reset();
            })
            .finally(() => {
              NProgress.done();
            });
        })
        .catch(() => {
          NProgress.done();
        });
    },
    update(post) {
      NProgress.start();
      this.$store
        .dispatch(post.id ? "postsUpdate" : "postsCreate", post)
        .then(result => {
          if (post.id) {
            setup.bind(this)();
            this.$v.$reset();
            NProgress.done();
          } else {
            this.$router.push({
              name: "post-edit",
              params: { pid: result.id }
            });
          }
        })
        .catch(() => {
          NProgress.done();
        });
    },
    del() {
      NProgress.start();
      this.$store
        .dispatch("postsDelete", this.posts.post.id)
        .then(() => {
          this.$router.push({
            name: "posts-list"
          });
        })
        .catch(() => {
          NProgress.done();
        });
    }
  }
};
</script>

<style scoped>
.postStatusWrapper {
  margin-bottom: 16px;
}
.postStatus {
  margin: 8px 0;
}
</style>
