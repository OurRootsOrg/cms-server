<template>
  <div class="posts-create">
    <h1>{{ post.id ? "Edit" : "Create" }} Post</h1>
    <form @submit.prevent="save">
      <h3>Give your post a name</h3>
      <BaseInput
        label="Name"
        v-model="post.name"
        type="text"
        placeholder="Name"
        class="field"
        :class="{ error: $v.post.name.$error }"
        @blur="$v.post.name.$touch()"
      />
      <template v-if="$v.post.name.$error">
        <p v-if="!$v.post.name.required" class="errorMessage">
          Name is required.
        </p>
      </template>

      <p v-if="post.id">Status: {{ posts.post.recordsStatus }}</p>

      <div v-if="post.id">
        <h3>Collection</h3>
        <p>{{ collections.collection.name }}</p>
      </div>
      <div v-else>
        <h3>Select a collection</h3>
        <BaseSelect
          label="Collection"
          :options="collections.collectionsList"
          v-model="post.collection"
          :class="{ error: $v.post.collection.$error }"
          @blur="$v.post.collection.$touch()"
        />
        <template v-if="$v.post.collection.$error">
          <p v-if="!$v.post.collection.required" class="errorMessage">
            Collection is required.
          </p>
        </template>
      </div>

      <div v-if="post.id">
        <h3>Post status</h3>
        <BaseSelect
          label="Status"
          :options="recordsStatusOptions"
          v-model="post.recordsStatus"
          :class="{ error: $v.post.recordsStatus.$error }"
          @blur="$v.post.recordsStatus.$touch()"
        />
      </div>

      <p v-if="$v.$anyError" class="errorMessage">
        Please fill out the required field(s).
      </p>

      <BaseButton type="submit" class="btn" buttonClass="-fill-gradient" :disabled="$v.$anyError">Save</BaseButton>

      <BaseButton v-if="post.recordsStatus === 'Draft'" @click="del" class="btn" buttonClass="danger"
        >Delete Post</BaseButton
      >

      <input
        v-if="post.id && post.recordsStatus === 'Draft'"
        type="button"
        id="importData"
        :value="post.recordsKey ? 'Replace data' : 'Import data'"
        @click.prevent="importData"
      />
    </form>

    <Tabulator
      v-if="post.id && post.recordsKey && post.recordsStatus !== 'Loading'"
      :data="records.recordsList.map(r => r.data)"
      :columns="getColumns()"
    />
  </div>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import { required } from "vuelidate/lib/validators";
import FlatfileImporter from "flatfile-csv-importer";
import config from "../flatfileConfig.js";
import Server from "@/services/Server.js";
import NProgress from "nprogress";

FlatfileImporter.setVersion(2);

function setup() {
  Object.assign(this.post, this.posts.post);
}

async function uploadData(store, post, contentType, data) {
  let postRequestResult = await Server.contentPostRequest(contentType);
  await Server.contentPut(postRequestResult.data.putURL, contentType, data.validData);
  post.recordsKey = postRequestResult.data.key;
  let postPostResult = await store.dispatch("postsUpdate", post);
  return postPostResult;
}

export default {
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
      recordsStatusOptions: []
    };
  },
  computed: mapState(["collections", "posts", "records", "settings"]),
  validations: {
    post: {
      name: { required },
      collection: { required }
    }
  },
  methods: {
    getColumns() {
      return this.collections.collection.fields.map(f => {
        return { title: f.header, field: f.header };
      });
    },
    save() {
      let post = Object.assign({}, this.post);
      post.collection = +post.collection; // convert to a number
      NProgress.start();
      this.$store
        .dispatch(post.id ? "postsUpdate" : "postsCreate", post)
        .then(result => {
          if (post.id) {
            setup.bind(this)();
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
      let post = Object.assign({}, this.posts.post);
      let collection = this.collections.collectionsList.find(coll => coll.id === post.collection);
      let store = this.$store;
      this.$v.$touch();
      if (!this.$v.$invalid) {
        const importer = new FlatfileImporter(config.license, this.getFlatFileOptions(collection));
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
              });
          })
          .catch(function(error) {
            console.info(error);
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
  margin-top: 24px;
}
#importData {
  position: relative;
  left: 50%;
  transform: translateX(-50%);
  margin: 96px 0 64px 0;
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
#raw_output {
  position: relative;
  left: 50%;
  transform: translateX(-50%);
  width: 50%;
  padding: 64px;
  margin: 32px 0;
  border: 1px solid #c1c6d1;
  border-radius: 4px;
  font-family: "Courier New", monospace;
  background-color: #3c4151;
  color: #fff;
  height: 200px;
}
</style>
