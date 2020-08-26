<template>
  <v-container class="settings">
    <v-row>
      <v-col cols="12">
        <h1>Settings</h1>
      </v-col>
    </v-row>
    <v-row class="mt-4">
      <v-col cols="12">
        <form @submit.prevent="save">
          <h3>
            Define custom post fields
            <v-tooltip bottom maxWidth="600px">
              <template v-slot:activator="{ on, attrs }">
                <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
              </template>
              <span
                >The fields you add here will be available for post metadata (data about the data within the post).
                Metadata <em>does not</em> appear in search results.</span
              >
            </v-tooltip>
          </h3>
          <v-row>
            <v-col cols="12">
              <v-data-table :headers="postMetadataColumns" :items="settingsObj.postMetadata" dense>
                <template v-slot:footer>
                  <v-toolbar flat color="white">
                    <v-dialog v-model="dialog" max-width="600px">
                      <template v-slot:activator="{ on, attrs }">
                        <v-btn class="secondary primary--text ml-n3" v-bind="attrs" v-on="on">New Custom Field</v-btn>
                      </template>
                      <v-card>
                        <v-card-title>
                          <span class="headline">{{ formTitle }}</span>
                        </v-card-title>
                        <v-card-text>
                          <v-container>
                            <v-row>
                              <v-col cols="12">
                                <v-text-field v-model="editedItem.name" label="Name"></v-text-field>
                              </v-col>
                              <v-col cols="12">
                                <v-select v-model="editedItem.type" :items="typeOptions" label="Field type"></v-select>
                              </v-col>
                              <v-col cols="12">
                                <v-text-field v-model="editedItem.tooltip" label="Tooltip"></v-text-field>
                              </v-col>
                            </v-row>
                          </v-container>
                        </v-card-text>
                        <v-card-actions>
                          <v-spacer></v-spacer>
                          <v-btn color="blue darken-1" text @click="close">Cancel</v-btn>
                          <v-btn color="blue darken-1" text @click="saveField">Save</v-btn>
                        </v-card-actions>
                      </v-card>
                    </v-dialog>
                  </v-toolbar>
                </template>
                <template v-slot:[`item.type`]="{ item }">
                  {{ typeOptions.find(x => x.value === item.type).text }}
                </template>
                <template v-slot:[`item.actions`]="{ item }">
                  <v-icon small class="mr-2" @click="editItem(item)">mdi-pencil</v-icon>
                  <v-icon small @click="deleteItem(item)">mdi-delete</v-icon>
                </template>
                <template v-slot:item.handle>
                  <v-btn icon small>
                    <v-icon left>mdi-drag-horizontal-variant</v-icon>
                  </v-btn>
                </template>
              </v-data-table>
            </v-col>
          </v-row>
          <v-row class="pl-3">
            <v-btn class="mt-4" type="submit" color="primary" :disabled="$v.$anyError || !$v.$anyDirty">
              <v-icon left>mdi-alert</v-icon>
              Important: Save all changes
            </v-btn>
            <p v-if="$v.$anyError" class="errorMessage">
              Please fill out the required field(s).
            </p>
          </v-row>
        </form>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import NProgress from "nprogress";
import lodash from "lodash";

function setup() {
  this.settingsObj = {
    ...this.settings.settings,
    postMetadata: lodash.cloneDeep(this.settings.settings.postMetadata)
  };
}

export default {
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    store
      .dispatch("settingsGet")
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  created() {
    setup.bind(this)();
  },
  data() {
    return {
      settingsObj: { postMetadata: [] },
      dialog: false,
      editedIndex: -1,
      editedItem: {
        name: "",
        type: "",
        tooltip: ""
      },
      defaultItem: {
        name: "",
        type: "",
        tooltip: ""
      },
      typeOptions: [
        { value: "string", text: "Text" },
        { value: "number", text: "Numeric" },
        { value: "date", text: "Date" },
        { value: "boolean", text: "Checkbox" }
      ],
      postMetadataColumns: [
        {
          text: "",
          value: "handle",
          width: 30
        },
        {
          text: "Name",
          value: "name"
        },
        {
          text: "Type",
          width: 80,
          value: "type"
        },
        {
          text: "Tooltip",
          value: "tooltip"
        },
        { title: "", value: "actions" }
      ]
    };
  },
  computed: {
    formTitle() {
      return this.editedIndex === -1 ? "New custom field" : "Edit custom field";
    },
    ...mapState(["settings"])
  },
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
    postMetadataMoved(data) {
      this.settingsObj.postMetadata = data;
      this.touch("postMetadata");
    },
    postMetadataEdited() {
      this.touch("postMetadata");
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
    },
    //methods for the CRUD table
    editItem(item) {
      this.editedIndex = this.settingsObj.postMetadata.indexOf(item);
      this.editedItem = Object.assign({}, item);
      this.dialog = true;
    },
    deleteItem(item) {
      const index = this.settingsObj.postMetadata.indexOf(item);
      confirm("Are you sure you want to delete this item?") && this.settingsObj.postMetadata.splice(index, 1);
      this.touch("postMetadata");
    },
    close() {
      this.dialog = false;
      this.$nextTick(() => {
        this.editedItem = Object.assign({}, this.defaultItem);
        this.editedIndex = -1;
      });
    },
    saveField() {
      if (this.editedIndex > -1) {
        Object.assign(this.settingsObj.postMetadata[this.editedIndex], this.editedItem);
      } else {
        this.settingsObj.postMetadata.push(this.editedItem);
      }
      this.touch("postMetadata");
      this.close();
    }
  }
};
</script>
