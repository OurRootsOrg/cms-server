<template>
  <div class="collections-create">
    <h1>{{ collection.id ? "Edit" : "Create" }} Collection</h1>
    <form @submit.prevent="createEditCollection">
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

      <h3>Define spreadsheet columns</h3>
      <Tabulator
        :data="collection.fields"
        :columns="columns"
        :movable-rows="true"
        :resizable-columns="true"
        @updated="updateData"
      />
      <a href="#" @click="addField">+ Add a spreadsheet column</a>
      <span v-if="collection.fields.length === 0">
        (you need at least one)
      </span>

      <BaseButton
        type="submit"
        class="submit-button"
        buttonClass="-fill-gradient"
        :disabled="$v.$anyError || collection.fields.length === 0"
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
import Tabulator from "../components/Tabulator";
import NProgress from "nprogress";
import { required } from "vuelidate/lib/validators";

export default {
  components: { Tabulator },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    let routes = [store.dispatch("categoriesGetAll")];
    if (routeTo.params && routeTo.params.cid) {
      routes.push(store.dispatch("collectionsGetOne", routeTo.params.cid));
    }
    Promise.all(routes).then(() => {
      next();
    });
  },
  created() {
    if (this.$route.params && this.$route.params.cid) {
      this.collection = Object.assign({}, this.collections.collection);
    }
  },
  data() {
    return {
      collection: { fields: [] },
      columns: [
        {
          rowHandle: true,
          formatter: "handle",
          headerSort: false,
          frozen: true,
          width: 30,
          minWidth: 30
        },
        {
          title: "Spreadsheet header",
          field: "header",
          tooltip: "spreadsheet column header (required)",
          editor: "input",
          validator: ["required", "unique"]
        },
        {
          title: "Field name",
          field: "name",
          tooltip: "database field name (required)",
          editor: "input",
          validator: ["required", "unique"]
        },
        {
          title: "Required?",
          field: "required",
          tooltip: "is this field required?",
          editor: "tickCross",
          hozAlign: "center",
          formatter: "tickCross",
          formatterParams: { allowEmpty: true, crossElement: "&ndash;" }
        },
        {
          title: "Validation rule",
          field: "regex",
          tooltip: "regular expression used to validate column values (optional)",
          editor: "input"
        },
        {
          title: "Validation Message",
          field: "regexError",
          tooltip: "message to report if the value fails the validation rule (optional)",
          editor: "input"
        },
        {
          title: "Delete",
          formatter: "buttonCross",
          hozAlign: "center",
          width: 55,
          minWidth: 55,
          cellClick: (e, cell) => {
            cell.getRow().delete();
          }
        }
      ]
    };
  },
  computed: mapState(["collections", "categories"]),
  validations: {
    collection: {
      name: { required },
      category: { required }
    }
  },
  methods: {
    addField() {
      this.collection.fields.push({});
    },
    updateData(data) {
      this.collection.fields = data;
    },
    // cellEdited(cell) {
    //   if (cell.getField() === "header") {
    //     cell
    //       .getRow()
    //       .getCell("regex")
    //       .setValue(cell.getValue(), true);
    //   }
    // },
    createEditCollection() {
      this.collection.fields = this.collection.fields.filter(f => f.header && f.name);
      if (this.collection.fields.length === 0) {
        return;
      }
      this.collection.category = +this.collection.category; // convert to a number
      this.$v.$touch();
      if (!this.$v.$invalid) {
        NProgress.start();
        this.$store
          .dispatch(this.collection.id ? "collectionsUpdate" : "collectionsCreate", this.collection)
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
