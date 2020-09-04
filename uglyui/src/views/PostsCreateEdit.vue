<template>
  <div class="posts-create">
    <h1>{{ post.id ? "Edit" : "Create" }} Post</h1>
    <v-form @submit.prevent="save">
      <h3>Give your post a name</h3>
      <v-text-field
        label="Post Name"
        v-model="post.name"
        type="text"
        placeholder="Name"
        class="field"
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

      <div v-if="post.id">
        <h3>Post status</h3>
        <p>
          <span>{{ post.recordsStatus }}</span
          ><span v-if="post.imagesStatus === 'Loading'"> - loading images</span>
          <span v-if="!!post.imagesKeys && post.imagesKeys.length > 0 && post.imagesStatus === 'Draft'">
            - with images</span
          >
        </p>
      </div>

      <div v-if="settings.settings.postMetadata.length > 0">
        <h3>Custom fields</h3>
        <Tabulator
          :data="metadata"
          :columns="getMetadataColumns()"
          layout="fitColumns"
          :resizable-columns="true"
          @cellEdited="metadataEdited"
        />
      </div>

      <p v-if="$v.$anyError" class="errorMessage">
        Please fill out the required field(s).
      </p>

      <v-row>
        <v-btn type="submit" color="primary" class="btn mt-4" :disabled="$v.$anyError || !$v.$anyDirty">Save </v-btn>
        <v-btn
          v-if="isPublishable"
          @click="publish"
          color="primary"
          class="btn mt-4"
          title="Publish the post to make it searchable"
          >Publish Post</v-btn
        >
        <v-btn
          v-if="isUnpublishable"
          @click="unpublish"
          color="primary"
          class="btn mt-4"
          title="Unpublish the post to remove it from the index"
          >Unpublish Post</v-btn
        >
        <v-btn
          v-if="isImportable"
          id="importData"
          @click="importData"
          color="primary"
          class="btn mt-4"
          title="Upload or replace records"
        >
          {{ post.recordsKey ? "Replace data" : "Import data" }}
        </v-btn>
        <v-dialog
          v-if="isImportable && collections.collection.imagePathHeader"
          v-model="importImagesDlg"
          persistent
          max-width="320"
        >
          <template v-slot:activator="{ on, attrs }">
            <v-btn color="primary" v-bind="attrs" v-on="on" class="btn mt-4" :disabled="post.imagesStatus !== 'Draft'">
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
                :multiple="false"
                :size="1024 * 1024 * 1024 * 10"
                v-model="imageFiles"
                @input-filter="imagesInputFilter"
                ref="upload"
              >
                <v-btn class="btn primary" :disabled="imagesUploading">Select ZIP file</v-btn>
              </file-upload>
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
              <v-btn text v-if="!imageFiles.find(f => f.success)" @click="cancelImagesUpload($refs.upload)"
                >Cancel</v-btn
              >
              <v-btn
                color="primary"
                text
                v-if="!imageFiles.find(f => f.success)"
                :disabled="!$refs.upload || $refs.upload.active"
                @click="startImagesUpload($refs.upload)"
                >Start Upload</v-btn
              >
              <v-btn color="primary" text v-if="imageFiles.find(f => f.success)" @click="endImagesUpload()"
                >Upload Successful!</v-btn
              >
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-row>
    </v-form>

    <v-btn :disabled="!isDeletable" @click="del" class="warning mt-2 mb-4">Delete Post </v-btn>

    <Tabulator
      v-if="post.id && post.recordsKey && post.recordsStatus !== 'Loading' && post.imagesStatus !== 'Loading'"
      layout="fitColumns"
      :data="records.recordsList.map(r => r.data)"
      :columns="getRecordColumns()"
    />
  </div>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import { required } from "vuelidate/lib/validators";
import { getMetadataColumnForEditing } from "../utils/metadata";
import Tabulator from "../components/Tabulator";
import FlatfileImporter from "flatfile-csv-importer";
import FileUpload from "vue-upload-component";
import config from "../utils/flatfileConfig.js";
import Server from "@/services/Server.js";
import NProgress from "nprogress";
import lodash from "lodash";

FlatfileImporter.setVersion(2);

function setup() {
  this.post = {
    ...this.posts.post
  };
  this.metadata.splice(0, 1, { ...this.posts.post.metadata });
}

async function uploadData(store, post, contentType, data) {
  let postRequestResult = await Server.contentPostRequest(contentType);
  await Server.contentPut(postRequestResult.data.putURL, contentType, data.validData);
  post.recordsKey = postRequestResult.data.key;
  let postPostResult = await store.dispatch("postsUpdate", post);
  return postPostResult;
}

export default {
  components: { Tabulator, FileUpload },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    let routes = [store.dispatch("settingsGet")];
    if (routeTo.params && routeTo.params.pid) {
      routes.push(store.dispatch("postsGetOne", routeTo.params.pid));
      routes.push(store.dispatch("recordsGetForPost", routeTo.params.pid));
    } else {
      routes.push(store.dispatch("collectionsGetAll"));
    }
    Promise.all(routes)
      .then(() => {
        if (routeTo.params && routeTo.params.pid) {
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
  },
  created() {
    if (this.$route.params && this.$route.params.pid) {
      setup.bind(this)();
    }
  },
  data() {
    return {
      importImagesDlg: false,
      imageFiles: [],
      imagesUploading: false,
      imagesPostRequestResultData: null,
      post: { id: null, name: null, collection: null, recordsStatus: null, recordsKey: null },
      metadata: [{}]
    };
  },
  computed: {
    isImportable() {
      return this.post.id && this.post.recordsStatus === "Draft";
    },
    isDeletable() {
      return !this.post.id || this.post.recordsStatus === "Draft";
    },
    isPublishable() {
      return this.post.id && this.post.recordsStatus === "Draft" && this.post.recordsKey;
    },
    isUnpublishable() {
      return this.post.id && this.post.recordsStatus === "Published";
    },
    ...mapState(["collections", "posts", "records", "settings", "user"])
  },
  validations: {
    post: {
      name: { required },
      collection: { required },
      recordsStatus: { required },
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
    metadataEdited() {
      this.touch("metadata");
    },
    getRecordColumns() {
      let cols = [];
      for (let f of this.collections.collection.fields) {
        cols.push({ title: f.header, field: f.header });
        if (
          this.collections.collection.mappings.find(
            m => m.header === f.header && (m.ixField.endsWith("Date") || m.ixField.endsWith("Place"))
          )
        ) {
          cols.push({ title: f.header + "_std", field: f.header + "_std" });
        }
      }
      return cols;
    },
    getMetadataColumns() {
      return this.settings.settings.postMetadata.map(pf => getMetadataColumnForEditing(pf));
    },
    getPostFromForm() {
      let post = Object.assign({}, this.post);
      post.collection = +post.collection; // convert to a number
      post.metadata = this.metadata[0];
      return post;
    },
    save() {
      let post = this.getPostFromForm();
      this.update(post);
    },
    publish() {
      let post = this.getPostFromForm();
      post.recordsStatus = "Published";
      this.update(post);
    },
    unpublish() {
      let post = this.getPostFromForm();
      post.recordsStatus = "Draft";
      this.update(post);
    },
    imagesInputFilter(newFile, oldFile, prevent) {
      if (newFile && !oldFile) {
        if (!newFile.name || !newFile.name.endsWith(".zip")) {
          return prevent();
        }
        return Server.contentPostRequest("application/zip").then(result => {
          console.log("contentPostRequest", result.data);
          this.imagesPostRequestResultData = result.data;
          newFile.putAction = result.data.putURL;
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
      let post = this.getPostFromForm();
      this.$v.$touch();
      NProgress.start();
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
    },
    importData() {
      let post = this.getPostFromForm();
      let store = this.$store;
      this.$v.$touch();
      if (!this.$v.$invalid) {
        const importer = new FlatfileImporter(config.license, this.getFlatFileOptions(this.collections.collection));
        console.log("setCustomer", this.user.user.id.toString(), this.user.user.email);
        importer.setCustomer({ userId: this.user.user.id.toString(), email: this.user.user.email });
        importer
          .requestDataFromUser()
          .then(results => {
            importer.displayLoader();
            uploadData(store, post, "application/json", results) // use application/json for records
              .then(() => {
                importer.displaySuccess("Success!");
                setup.bind(this)();
                this.$v.$reset();
              });
          })
          .catch(() => {
            // console.info(error);
          });
      }
    },
    getFlatFileOptions(coll) {
      return {
        type: "Record",
        allowInvalidSubmit: true,
        managed: true,
        allowCustom: false,
        disableManualInput: true,
        fields: coll.fields.map(fld => {
          let validators = [];
          if (fld.required) {
            validators.push({ validate: "required", error: "required field" });
          }
          if (fld.regex) {
            validators.push({
              validate: "regex_matches",
              regex: fld.regex,
              error: fld.regexError || "doesn't match validation rule"
            });
          }
          return {
            label: fld.header,
            key: fld.header,
            validators: validators
          };
        })
      };
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.btn {
  margin: 16px;
}
</style>
