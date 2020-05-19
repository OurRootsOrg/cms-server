<template>
  <div class="posts-create">
    <h1>Create Post</h1>
    <form @submit.prevent="createPost">
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

      <input type="button" id="launch" value="Import data" @click.prevent="launch" />
      <p v-if="$v.$anyError" class="errorMessage">
        Please fill out the required field(s).
      </p>
    </form>
    <textarea id="raw_output" :value="results"></textarea>
  </div>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import { required } from "vuelidate/lib/validators";
import FlatfileImporter from "flatfile-csv-importer";
import config from "../flatfileConfig.js";
import Server from "@/services/Server.js";

FlatfileImporter.setVersion(2);

const importer = new FlatfileImporter(config.license, config.config);

// importer.registerRecordHook(record => {
//   let out = {};
//   if (record.name.includes(" ")) {
//     out.name = {
//       value: record.name.replace(" ", "_"),
//       info: [
//         {
//           message: "No spaces allowed. Replaced all spaces with underscores",
//           level: "warning"
//         }
//       ]
//     };
//   }
//   return out;
// });

async function uploadData(store, post, contentType, data) {
  console.log("uploadData", contentType, data, data.validData);
  let postRequestResult = await Server.contentPostRequest(contentType);
  console.log("postRequestResult", postRequestResult, postRequestResult.data.putURL);
  let putResult = await Server.contentPut(postRequestResult.data.putURL, contentType, data.validData);
  console.log("putResult", putResult);
  post.key = postRequestResult.data.key;
  console.log("upload post", post);
  let postPostResult = await store.dispatch("postsCreate", post);
  console.log("postPostResult", postPostResult);
  return postPostResult;
}

export default {
  beforeRouteEnter(routeTo, routeFrom, next) {
    store.dispatch("collectionsGetAll").then(() => {
      next();
    });
  },
  data() {
    return {
      post: {},
      results: "Your raw output will appear here."
    };
  },
  computed: mapState(["collections"]),
  validations: {
    post: {
      name: { required },
      collection: { required }
    }
  },
  methods: {
    launch() {
      let post = this.post;
      let store = this.$store;
      console.log("post", post, "store", store);
      this.$v.$touch();
      if (!this.$v.$invalid) {
        importer.setCustomer({ userId: 1, email: "dallan@gmail.com" });
        importer
          .requestDataFromUser()
          .then(results => {
            importer.displayLoader();
            console.log("post again", post);
            uploadData(store, post, "application/json", results) // use application/json for records
              .then(() => {
                console.log("SUCCESS!");
                importer.displaySuccess("Success!");
                this.$router.push({
                  name: "posts-list"
                });
              });
          })
          .catch(function(error) {
            console.info(error);
          });
      }
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.download {
  position: relative;
  left: 50%;
  transform: translateX(-50%);
  text-align: center;
}
.download a {
  color: #3c4151;
  text-decoration: none;
}
.download a:hover {
  color: #1d62b4;
}
#launch {
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
#launch:focus,
#launch:hover {
  background-color: #2171ce;
}
#launch:active {
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
