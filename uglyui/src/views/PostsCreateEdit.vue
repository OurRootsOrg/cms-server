<template>
  <v-container class="posts-create">
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
            <p>{{ post.recordsStatus }}</p>
          </div>

          <div v-if="settings.settings.postMetadata.length > 0">
            <h3>Custom fields (metadata) for this post 
                <v-tooltip bottom>
                  <template v-slot:activator="{ on, attrs }">
                    <v-icon
                      small
                      v-bind="attrs"
                      v-on="on"
                    >mdi-information</v-icon>
                  </template>
                  <span>Information about/specific to <em>this particular post</em> such as the transcription date, translator, etc. which might be different from other posts in this collection</span>
                </v-tooltip>
            </h3>
            <!-- <Tabulator
              :data="metadata"
              :columns="getMetadataColumns()"
              layout="fitColumns"
              :resizable-columns="true"
              @cellEdited="metadataEdited"
            /> -->

            <v-data-table
              :items="metadata"
              :headers="getMetadataColumns()"
              dense
            >
            </v-data-table>   

          </div>


          <p v-if="$v.$anyError" class="errorMessage">
            Please fill out the required field(s).
          </p>

          <v-row>
            <v-col class="d-flex justify-space-between">
              <v-btn
               type="submit" 
               color="primary" 
               :disabled="$v.$anyError || !$v.$anyDirty"
              >
                Save
              </v-btn>
              <v-btn
                v-if="isPublishable"
                @click="publish"
                color="primary"
                title="Publish the post to make it searchable"
              >
                Publish Post
              </v-btn>
              <v-btn
                v-if="isUnpublishable"
                @click="unpublish"
                color="primary"
                title="Unpublish the post to remove it from the index"
              >
                Unpublish Post
              </v-btn>
              <v-btn
                v-if="isImportable"
                id="importData"
                @click="importData"
                color="primary"
                title="Upload or replace records"
              >
                {{ post.recordsKey ? "Replace data" : "Import data" }}
              </v-btn>
              <v-btn :disabled="!isDeletable" @click="del" class="warning">Delete Post</v-btn>
            </v-col>
          </v-row>
        </v-form>

        <!-- <Tabulator
          v-if="post.id && post.recordsKey && post.recordsStatus !== 'Loading'"
          layout="fitColumns"
          :data="records.recordsList.map(r => r.data)"
          :columns="getRecordColumns()"
        /> -->
      <v-row class="pt-5">
        <v-col>
          <h3 v-if="post.id && post.recordsKey && post.recordsStatus !== 'Loading'" class="pl-1">Post data</h3>
          <v-data-table
            v-if="post.id && post.recordsKey && post.recordsStatus !== 'Loading'"
            :items="records.recordsList.map(r => r.data)"
            :headers="getRecordColumns()"
            dense
            sortable
            :footer-props="{
              'items-per-page-options': [10, 25, 50]
            }"
            :items-per-page="25"
          >
          </v-data-table>
        </v-col>
      </v-row>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import { required } from "vuelidate/lib/validators";
import { getMetadataColumnForEditing } from "../utils/metadata";
// import Tabulator from "../components/Tabulator";
import FlatfileImporter from "flatfile-csv-importer";
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
  // components: { Tabulator },
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
  //added this watch for the crud metadata table
  watch: {
    dialog (val) {
      val || this.close()
    },
  },  
  data() {
    return {
      //for the crud table: dialog, edited index, edited item, default item
      dialog: false,
      editedIndex: -1,
      editedItem: {},
      defaultItem: {},
      post: { id: null, name: null, collection: null, recordsStatus: null, recordsKey: null },
      metadata: [{}],
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
    ...mapState(["collections", "posts", "records", "settings"])
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
      //Tabulator:v-data-table translation is title:text and field:value (rename "title" as "text" and "field" as "value")
      return this.collections.collection.fields.map(f => {
        return { text: f.header, value: f.header };
      });
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
        // TODO set to real user
        importer.setCustomer({ userId: 1, email: "dallan@gmail.com" });
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
    },
    //methods for the custom fields table
      editCustomFieldItem (item) {
        this.editedIndex = this.metadata.indexOf(item)
        this.editedItem = Object.assign({}, item)
        this.dialog = true
      },
      close () {
        this.dialog = false
        this.$nextTick(() => {
          this.editedItem = Object.assign({}, this.defaultItem)
          this.editedIndex = -1
        })
      },
      saveCustomField () {
        if (this.editedIndex > -1) {
          Object.assign(this.metadata[this.editedIndex], this.editedItem)
        } else {
          this.metadata.push(this.editedItem)
        }
        this.close()
      },
  }
};
</script>

