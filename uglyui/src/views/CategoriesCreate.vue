<template>
  <div class="categories-create">
    <h1>Create Category</h1>
    <form @submit.prevent="createCategory">
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
        >Submit</BaseButton
      >
      <p v-if="$v.$anyError" class="errorMessage">
        Please fill out the required field(s).
      </p>
    </form>
  </div>
</template>

<script>
import NProgress from "nprogress";
import { required } from "vuelidate/lib/validators";

export default {
  data() {
    return {
      category: {}
    };
  },
  validations: {
    category: {
      name: { required }
    }
  },
  methods: {
    createCategory() {
      this.$v.$touch();
      if (!this.$v.$invalid) {
        NProgress.start();
        this.$store
          .dispatch("categoriesCreate", this.category)
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
  }
};
</script>

<style scoped>
.submit-button {
  margin-top: 32px;
}
</style>
