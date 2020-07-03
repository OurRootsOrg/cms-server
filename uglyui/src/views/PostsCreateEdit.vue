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
        <v-select
          label="Status"
          :items="getRecordsStatusOptions()"
          item-text="name"
          item-value="id"
          v-model="post.recordsStatus"
          :class="{ error: $v.post.recordsStatus.$error }"
          @input="touch('recordsStatus')"
        ></v-select>
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

      <v-btn type="submit" color="primary" :disabled="$v.$anyError || !$v.$anyDirty">Save </v-btn>

      <input
        v-if="post.id && post.recordsStatus === 'Draft'"
        type="button"
        id="importData"
        :value="post.recordsKey ? 'Replace data' : 'Import data'"
        @click.prevent="importData"
      />
    </v-form>

    <v-btn v-if="post.recordsStatus === 'Draft'" @click="del" class="warning">Delete Post </v-btn>

    <Tabulator
      v-if="post.id && post.recordsKey && post.recordsStatus !== 'Loading'"
      :data="records.recordsList.map(r => r.data)"
      :columns="getRecordColumns()"
    />
  </div>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import { required } from "vuelidate/lib/validators";
import Tabulator from "../components/Tabulator";
import FlatfileImporter from "flatfile-csv-importer";
import config from "../flatfileConfig.js";
import Server from "@/services/Server.js";
import NProgress from "nprogress";
import moment from "moment";
import lodash from "lodash";

FlatfileImporter.setVersion(2);

function setup() {
  Object.assign(this.post, this.posts.post);
  this.metadata.splice(0, 1, Object.assign({}, this.posts.post.metadata));
}

async function uploadData(store, post, contentType, data) {
  let postRequestResult = await Server.contentPostRequest(contentType);
  await Server.contentPut(postRequestResult.data.putURL, contentType, data.validData);
  post.recordsKey = postRequestResult.data.key;
  let postPostResult = await store.dispatch("postsUpdate", post);
  return postPostResult;
}

function dateEditor(cell, onRendered, success, cancel) {
  //cell - the cell component for the editable cell
  //onRendered - function to call when the editor has been rendered
  //success - function to call to pass the successfuly updated value to Tabulator
  //cancel - function to call to abort the edit and return to a normal cell

  //create and style input
  var cellValue = moment(cell.getValue(), "DD MMM YYYY").format("YYYY-MM-DD"),
    input = document.createElement("input");

  input.setAttribute("type", "date");

  input.style.padding = "4px";
  input.style.width = "100%";
  input.style.boxSizing = "border-box";

  input.value = cellValue;

  onRendered(function() {
    input.focus();
    input.style.height = "100%";
  });

  function onChange() {
    if (input.value !== cellValue) {
      success(moment(input.value, "YYYY-MM-DD").format("DD MMM YYYY"));
    } else {
      cancel();
    }
  }

  //submit new value on blur or change
  input.addEventListener("blur", onChange);

  //submit new value on enter
  input.addEventListener("keydown", function(e) {
    if (e.keyCode === 13) {
      onChange();
    }

    if (e.keyCode === 27) {
      cancel();
    }
  });

  return input;
}

function getMetadataColumn(pf) {
  switch (pf.type) {
    case "string":
      return {
        title: pf.name,
        minWidth: 200,
        widthGrow: 2,
        field: pf.name,
        tooltip: pf.tooltip,
        editor: "input"
      };
    case "number":
      return {
        title: pf.name,
        minWidth: 75,
        widthGrow: 1,
        field: pf.name,
        tooltip: pf.tooltip,
        editor: "number"
      };
    case "date":
      return {
        title: pf.name,
        minWidth: 140,
        widthGrow: 1,
        field: pf.name,
        tooltip: pf.tooltip,
        hozAlign: "center",
        editor: dateEditor
      };
    case "boolean":
      return {
        title: pf.name,
        minWidth: 75,
        widthGrow: 1,
        field: pf.name,
        tooltip: pf.tooltip,
        hozAlign: "center",
        formatter: "tickCross",
        editor: true
      };
  }
}

export default {
  components: { Tabulator },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    let routes = [store.dispatch("settingsGet")];
    if (routeTo.params && routeTo.params.pid) {
      routes.push(store.dispatch("postsGetOne", routeTo.params.pid));
      routes.push(store.dispatch("recordsGetForPost", routeTo.params.pid));
    } else {
      routes.push(store.dispatch("collectionsGetAll"));
    }
    Promise.all(routes).then(() => {
      if (routeTo.params && routeTo.params.pid) {
        store.dispatch("collectionsGetOne", store.state.posts.post.collection).then(() => {
          next();
        });
      } else {
        next();
      }
    });
  },
  created() {
    if (this.$route.params && this.$route.params.pid) {
      setup.bind(this)();
    }
  },
  data() {
    return {
      post: {},
      metadata: [{}]
    };
  },
  computed: mapState(["collections", "posts", "records", "settings"]),
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
      return this.collections.collection.fields.map(f => {
        return { title: f.header, field: f.header };
      });
    },
    getRecordsStatusOptions() {
      let opts = [];
      if (this.post.id) {
        opts.push({ id: this.post.recordsStatus, name: this.post.recordsStatus });
        if (this.post.recordsStatus === "Draft" && this.post.recordsKey) {
          opts.push({ id: "Published", name: "Published" });
        } else if (this.post.recordsStatus === "Published") {
          opts.push({ id: "Draft", name: "Draft" });
        }
      }
      return opts;
    },
    getMetadataColumns() {
      return this.settings.settings.postMetadata.map(pf => getMetadataColumn(pf));
    },
    getPostFromForm() {
      let post = Object.assign({}, this.post);
      post.collection = +post.collection; // convert to a number
      post.metadata = this.metadata[0];
      return post;
    },
    save() {
      let post = this.getPostFromForm();
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
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.btn {
  margin: 24px 0;
}
#importData {
  margin: 24px 0;
  border: none;
  border-radius: 4px;
  color: #fff;
  display: block;
  padding: 0 64px;
  height: 44px;
  cursor: pointer;
  transition: background-color 0.2s;
  font-size: 15px;
  font-weight: 500;
  outline: 0;
  background-color: #4a90e2;
}
#importData:focus,
#importData:hover {
  background-color: #2171ce;
}
#importData:active {
  background-color: #1d62b4;
}
</style>
