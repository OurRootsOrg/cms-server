<template>
  <div class="categories-create">
    <h1>{{ category.id ? "Edit" : "Create" }} Category</h1>
    <form @submit.prevent="save">
      <h3>Give your category a name</h3>
      <BaseInput
        label="Name"
        v-model="category.name"
        type="text"
        placeholder="Name"
        class="field"
        :class="{ error: $v.category.name.$error }"
        @blur="$v.category.name.$touch()"
      />

      <template v-if="$v.category.name.$error">
        <p v-if="!$v.category.name.required" class="errorMessage">
          Name is required.
        </p>
      </template>

      <BaseButton type="submit" class="submit-button" buttonClass="-fill-gradient" :disabled="$v.$anyError"
        >Save</BaseButton
      >
      <p v-if="$v.$anyError" class="errorMessage">
        Please fill out the required field(s).
      </p>
    </form>
    <BaseButton
      v-if="category.id"
      class="btn"
      buttonClass="danger"
      @click="del()"
      :disabled="collectionsForCategory.length > 0"
      >Delete Category</BaseButton
    >
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
  </div>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import Tabulator from "../components/Tabulator";
import NProgress from "nprogress";
import { required } from "vuelidate/lib/validators";

function setup() {
  Object.assign(this.category, this.categories.category);
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
      category: {},
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

<style scoped>
.submit-button {
  margin-top: 32px;
}
.btn {
  margin: 24px 0;
}
</style>
