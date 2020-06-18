<template>
  <div class="collections-create">
    <h1>Create Collection</h1>
    <form @submit.prevent="createCollection">
      <h3>Give your collection a name</h3>
      <BaseInput
        label="Name"
        v-model="collection.name"
        type="text"
        placeholder="Name"
        class="field"
        :class="{ error: $v.collection.name.$error }"
        @blur="$v.collection.name.$touch()"
      />

      <template v-if="$v.collection.name.$error">
        <p v-if="!$v.collection.name.required" class="errorMessage">
          Name is required.
        </p>
      </template>

      <h3>Select a category</h3>
      <BaseSelect
        label="Category"
        :options="categories.categoriesList"
        v-model="collection.category"
        :class="{ error: $v.collection.category.$error }"
        @blur="$v.collection.category.$touch()"
      />
      <template v-if="$v.collection.category.$error">
        <p v-if="!$v.collection.category.required" class="errorMessage">
          Category is required.
        </p>
      </template>

      <BaseButton type="submit" class="submit-button" buttonClass="-fill-gradient" :disabled="$v.$anyError"
        >Submit</BaseButton
      >
      <p v-if="$v.$anyError" class="errorMessage">
        Please fill out the required field(s).
      </p>
    </form>
  </div>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import NProgress from "nprogress";
import { required } from "vuelidate/lib/validators";

export default {
  beforeRouteEnter(routeTo, routeFrom, next) {
    store.dispatch("categoriesGetAll").then(() => {
      next();
    });
  },
  data() {
    return {
      collection: {}
    };
  },
  computed: mapState(["categories"]),
  validations: {
    collection: {
      name: { required },
      category: { required }
    }
  },
  methods: {
    createCollection() {
      this.collection.category = +this.collection.category; // convert to a number
      this.$v.$touch();
      if (!this.$v.$invalid) {
        NProgress.start();
        this.$store
          .dispatch("collectionsCreate", this.collection)
          .then(() => {
            this.$router.push({
              name: "collections-list"
            });
          })
          .catch(() => {
            NProgress.done();
          });
      }
    }
  }
};
</script>

<style scoped>
.submit-button {
  margin-top: 32px;
}
</style>
