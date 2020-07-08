<template>
  <v-container class="categories-create">
    <h1>{{ category.id ? "Edit" : "Create" }} Category</h1>
    <v-form @submit.prevent="save">
      <h3 class="mb-4">Give your category a name</h3>
      <v-text-field
        label="Category Name"
        v-model="category.name"
        type="text"
        placeholder="Name"
        class="field"
        :class="{ error: $v.category.name.$error }"
        @blur="touch('name')"
      >
      </v-text-field>

      <template v-if="$v.category.name.$error">
        <p v-if="!$v.category.name.required" class="errorMessage">
          Name is required.
        </p>
      </template>

      <v-layout row>
        <v-flex justify-center>
          <v-btn class="ml-2 mb-4" color="primary" type="submit" :disabled="$v.$anyError || !$v.$anyDirty">Save</v-btn>
          <span v-if="$v.$anyError" class="red--text">
            Please fill out the required field(s).
          </span>
        </v-flex>
      </v-layout>
    </v-form>

    <v-btn
      class="mt-2 mb-4"
      v-if="category.id"
      color="warning"
      @click="del()"
      :title="collectionsForCategory.length > 0 ? 'Categories with collections cannot be deleted' : 'Cannot be undone!'"
      :disabled="collectionsForCategory.length > 0"
      >Delete Category
    </v-btn>

    <h3 v-if="category.id">Collections</h3>
    <Tabulator
      v-if="category.id"
      :data="collectionsForCategory"
      :columns="collectionColumns"
      layout="fitColumns"
      :header-sort="true"
      :selectable="true"
      :resizable-columns="true"
      @rowClicked="collectionRowClicked"
    />
    <v-btn outlined color="primary" class="mt-4" to="/collections/create">
      Create a new collection
    </v-btn>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import Tabulator from "../components/Tabulator";
import NProgress from "nprogress";
import { required } from "vuelidate/lib/validators";
import lodash from "lodash";

function setup() {
  this.category = {
    ...this.categories.category
  };
}

export default {
  components: { Tabulator },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    let routes = [];
    if (routeTo.params && routeTo.params.cid) {
      routes.push(store.dispatch("categoriesGetOne", routeTo.params.cid));
      routes.push(store.dispatch("categoriesGetAll"));
      routes.push(store.dispatch("collectionsGetAll"));
      routes.push(store.dispatch("postsGetAll"));
    }
    Promise.all(routes).then(() => {
      next();
    });
  },
  created() {
    if (this.$route.params && this.$route.params.cid) {
      setup.bind(this)();
    }
  },
  data() {
    return {
      category: { id: null, name: null },
      collectionsList: [],
      collectionColumns: [
        {
          title: "Name",
          field: "name",
          headerFilter: "input",
          sorter: "string"
        },
        {
          title: "# Posts",
          field: "postsCount",
          headerFilter: "number",
          sorter: "number"
        },
        {
          title: "Categories",
          field: "categoryNames",
          headerFilter: "input",
          sorter: "string"
        }
      ]
    };
  },
  computed: {
    collectionsForCategory() {
      return this.collections.collectionsList
        .filter(coll => coll.categories.includes(this.category.id))
        .map(c => {
          return {
            id: c.id,
            name: c.name,
            postsCount: this.posts.postsList.filter(post => post.collection === c.id).length,
            categoryNames: this.categories.categoriesList
              .filter(cat => c.categories.includes(cat.id))
              .map(cat => cat.name)
              .join(", ")
          };
        });
    },
    ...mapState(["collections", "categories", "posts"])
  },
  validations: {
    category: {
      name: { required }
    }
  },
  methods: {
    touch(attr) {
      if (this.$v.category[attr].$dirty) {
        return;
      }
      if (!this.category.id || !lodash.isEqual(this.category[attr], this.categories.category[attr])) {
        this.$v.category[attr].$touch();
      }
    },
    collectionRowClicked(coll) {
      this.$router.push({
        name: "collection-edit",
        params: { cid: coll.id }
      });
    },
    save() {
      let category = Object.assign({}, this.category);
      this.$v.$touch();
      if (!this.$v.$invalid) {
        NProgress.start();
        this.$store
          .dispatch(category.id ? "categoriesUpdate" : "categoriesCreate", category)
          .then(result => {
            if (category.id) {
              setup.bind(this)();
              this.$v.$reset();
              NProgress.done();
            } else {
              this.$router.push({
                name: "category-edit",
                params: { cid: result.id }
              });
            }
          })
          .catch(() => {
            NProgress.done();
          });
      }
    },
    del() {
      if (this.collectionsForCategory.length > 0) {
        return;
      }
      NProgress.start();
      this.$store
        .dispatch("categoriesDelete", this.category.id)
        .then(() => {
          this.$router.push({
            name: "categories-list"
          });
        })
        .catch(() => {
          NProgress.done();
        });
    }
  }
};
</script>
