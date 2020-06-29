<template>
  <div class="settings">
    <h1>Settings</h1>
    <form @submit.prevent="save">
      <h3>Define custom post fields</h3>
      <Tabulator
        :data="settingsObj.postFields"
        :columns="postFieldColumns"
        layout="fitColumns"
        :movable-rows="true"
        :resizable-columns="true"
        @rowMoved="postFieldsMoved"
      />
      <a href="" @click.prevent="addPostField">Add a row</a>
      <BaseButton type="submit" class="submit-button" buttonClass="-fill-gradient" :disabled="$v.$anyError"
        >Save</BaseButton
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

const postFieldTypes = {
  string: "Text",
  number: "Numeric",
  date: "Date",
  boolean: "Checkbox",
  rating: "Star rating"
};

function setup() {
  Object.assign(this.settingsObj, this.settings.settings);
}

export default {
  components: { Tabulator },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    store.dispatch("settingsGet").then(() => {
      next();
    });
  },
  created() {
    setup.bind(this)();
  },
  data() {
    return {
      settingsObj: {},
      postFieldColumns: [
        {
          rowHandle: true,
          formatter: "handle",
          headerSort: false,
          frozen: true,
          width: 30,
          minWidth: 30
        },
        {
          title: "Name",
          minWidth: 200,
          widthGrow: 2,
          field: "name",
          tooltip: "custom field name",
          editor: "input",
          validator: ["unique"]
        },
        {
          title: "Type",
          width: 80,
          field: "type",
          tooltip: "type of data the field will hold",
          formatter: "lookup",
          formatterParams: postFieldTypes,
          editor: "select",
          editorParams: {
            values: postFieldTypes,
            defaultValue: "string"
          },
          validator: ["required"]
        },
        {
          title: "Delete",
          formatter: "buttonCross",
          hozAlign: "center",
          width: 55,
          minWidth: 55,
          cellClick: (e, cell) => {
            this.postFieldsDelete(cell.getRow().getPosition());
          }
        }
      ]
    };
  },
  computed: mapState(["settings"]),
  validations: {},
  methods: {
    addPostField() {
      this.settingsObj.postFields.push({ type: "string" });
    },
    postFieldsMoved(data) {
      this.settingsObj.postFields = data;
    },
    postFieldsDelete(ix) {
      this.settingsObj.postFields.splice(ix, 1);
    },
    save() {
      this.settingsObj.postFields = this.settingsObj.postFields.filter(f => f.name && f.type);
      this.$v.$touch();
      if (!this.$v.$invalid) {
        NProgress.start();
        this.$store
          .dispatch("settingsUpdate", this.settingsObj)
          .then(() => {
            setup.bind(this)();
            NProgress.done();
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
