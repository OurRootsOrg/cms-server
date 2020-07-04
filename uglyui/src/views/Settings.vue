<template>
  <v-container class="settings">
    <v-layout row>
      <v-flex>
        <h1>Settings</h1>
        <v-btn class="mt-2 mb-4" color="primary" href="" @click.prevent="addPostMetadata">Add a custom field</v-btn>
      </v-flex>
    </v-layout>
    <v-layout row class="mt-4">
      <form @submit.prevent="save">
        <h3>Define custom post fields</h3>
        <Tabulator
          :data="settingsObj.postMetadata"
          :columns="postMetadataColumns"
          layout="fitColumns"
          :movable-rows="true"
          :resizable-columns="true"
          @rowMoved="postMetadataMoved"
          @cellEdited="postMetadataEdited"
        />
        <v-row class="pl-3">
          <v-btn class="mt-4" type="submit" :disabled="$v.$anyError || !$v.$anyDirty">Save </v-btn>
          <p v-if="$v.$anyError" class="errorMessage">
            Please fill out the required field(s).
          </p>
        </v-row>
      </form>
    </v-layout>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import Tabulator from "../components/Tabulator";
import NProgress from "nprogress";
import lodash from "lodash";
//import {required} from "vuelidate/lib/validators";

const postMetadataTypes = {
  string: "Text",
  number: "Numeric",
  date: "Date",
  boolean: "Checkbox"
};

function setup() {
  Object.assign(this.settingsObj, this.settings.settings);
  // deep-clone arrays
  this.settingsObj.postMetadata = lodash.cloneDeep(this.settings.settings.postMetadata);
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
      settingsObj: { postMetadata: [] },
      postMetadataColumns: [
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
          formatterParams: postMetadataTypes,
          editor: "select",
          editorParams: {
            values: postMetadataTypes,
            defaultValue: "string"
          },
          validator: ["required"]
        },
        {
          title: "Tooltip",
          minWidth: 200,
          widthGrow: 2,
          field: "tooltip",
          tooltip: "tooltip for field (optional)",
          editor: "input"
        },
        {
          title: "Delete",
          formatter: "buttonCross",
          hozAlign: "center",
          width: 55,
          minWidth: 55,
          cellClick: (e, cell) => {
            this.postMetadataDelete(cell.getRow().getPosition());
          }
        }
      ]
    };
  },
  computed: mapState(["settings"]),
  validations: {
    settingsObj: {
      postMetadata: {}
    }
  },
  methods: {
    touch(attr) {
      if (this.$v.settingsObj[attr].$dirty) {
        return;
      }
      if (!lodash.isEqual(this.settingsObj[attr], this.settings.settings[attr])) {
        this.$v.settingsObj[attr].$touch();
      }
    },
    addPostMetadata() {
      this.settingsObj.postMetadata.push({ type: "string" });
    },
    postMetadataMoved(data) {
      this.settingsObj.postMetadata = data;
      this.touch("postMetadata");
    },
    postMetadataEdited() {
      this.touch("postMetadata");
    },
    postMetadataDelete(ix) {
      this.settingsObj.postMetadata.splice(ix, 1);
    },
    save() {
      this.settingsObj.postMetadata = this.settingsObj.postMetadata.filter(f => f.name && f.type);
      this.$v.$touch();
      if (!this.$v.$invalid) {
        NProgress.start();
        this.$store
          .dispatch("settingsUpdate", this.settingsObj)
          .then(() => {
            setup.bind(this)();
            this.$v.$reset();
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
