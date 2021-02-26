<template>
  <v-container class="settings">
    <h1>Society Settings</h1>
    <v-form @submit.prevent="save">
      <h3 style="margin-top: 16px;">
        ID
        <v-tooltip bottom maxWidth="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
          </template>
          <span>Copy this field into your wordpress plugin</span>
        </v-tooltip>
      </h3>
      <div style="margin-top: 8px;">{{ society.id }}</div>

      <h3 style="margin-top: 16px;">Name</h3>
      <v-text-field
        label="Society Name"
        v-model="society.name"
        type="text"
        placeholder="Name"
        @blur="touch('name')"
      ></v-text-field>
      <template v-if="$v.society.name.$error">
        <p v-if="!$v.society.name.required" class="errorMessage">
          Name is required.
        </p>
      </template>

      <h3 style="margin-top: 16px;">
        Secret key
        <v-tooltip bottom maxWidth="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
          </template>
          <span>Copy this field into your wordpress plugin</span>
        </v-tooltip>
      </h3>
      <v-text-field
        label="Secret key"
        v-model="society.secretKey"
        type="text"
        placeholder="Secret key"
        @blur="touch('secretKey')"
      ></v-text-field>
      <template v-if="$v.society.secretKey.$error">
        <p v-if="!$v.society.secretKey.required" class="errorMessage">
          Secret key is required.
        </p>
      </template>

      <h3 style="margin-top: 16px;">
        Login URL
        <v-tooltip bottom maxWidth="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
          </template>
          <span>If record details or media are private, send users to this URL to join your society</span>
        </v-tooltip>
      </h3>
      <v-text-field
        label="Login URL"
        v-model="society.loginURL"
        type="text"
        placeholder="Login URL"
        @blur="touch('loginURL')"
      ></v-text-field>
      <template v-if="$v.society.loginURL.$error">
        <p v-if="!$v.society.loginURL.required" class="errorMessage">
          Login url must be a URL
        </p>
      </template>

      <v-row>
        <v-col cols="12">
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

          <v-data-table
            :headers="postMetadataColumns"
            :items="society.postMetadata"
            item-key="id"
            :disable-pagination="true"
            dense
            v-columns-resizable
          >
            <template v-slot:body>
              <draggable :list="society.postMetadata" tag="tbody" @change="metadataDrag">
                <tr v-for="(item, index) in society.postMetadata" :key="index">
                  <td><v-icon small class="page__grab-icon">mdi-drag-horizontal-variant</v-icon></td>
                  <td>{{ item.name }}</td>
                  <td>{{ typeOptions.find(x => x.value === item.type).text }}</td>
                  <td>{{ item.tooltip }}</td>
                  <td>
                    <v-icon small @click="editItem(item)" class="mr-3">mdi-pencil</v-icon>
                    <v-icon small @click="deleteItem(item)">mdi-delete</v-icon>
                  </td>
                </tr>
              </draggable>
            </template>
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
                      <v-btn
                        color="blue darken-1"
                        text
                        @click="saveField"
                        :disabled="!editedItem.name || !editedItem.type"
                        >Save</v-btn
                      >
                    </v-card-actions>
                  </v-card>
                </v-dialog>
              </v-toolbar>
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
    </v-form>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import NProgress from "nprogress";
import { required, url } from "vuelidate/lib/validators";
import lodash from "lodash";
import draggable from "vuedraggable";

function getContent(next) {
  store
    .dispatch("societiesGetCurrent")
    .then(() => {
      next();
    })
    .catch(() => {
      next("/");
    });
}

function setup() {
  console.log("!!! setup", this.societies.society);
  this.society = {
    ...this.societies.society,
    postMetadata: lodash.cloneDeep(this.societies.society.postMetadata)
  };
}

export default {
  components: { draggable },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    getContent(next);
  },
  beforeRouteUpdate: function(routeTo, routeFrom, next) {
    getContent(next);
  },
  created() {
    setup.bind(this)();
  },
  data() {
    return {
      society: {},
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
    ...mapState(["societies"])
  },
  validations: {
    society: {
      name: { required },
      secretKey: { required },
      loginURL: { url },
      postMetadata: {}
    }
  },
  methods: {
    touch(attr) {
      if (this.$v.society[attr].$dirty) {
        return;
      }
      if (!lodash.isEqual(this.society[attr], this.societies.society[attr])) {
        this.$v.society[attr].$touch();
      }
    },
    postMetadataMoved(data) {
      this.society.postMetadata = data;
      this.touch("postMetadata");
    },
    postMetadataEdited() {
      this.touch("postMetadata");
    },
    save() {
      this.society.postMetadata = this.society.postMetadata.filter(f => f.name && f.type);
      this.$v.$touch();
      if (!this.$v.$invalid) {
        NProgress.start();
        this.$store
          .dispatch("societiesUpdate", this.society)
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
      this.editedIndex = this.society.postMetadata.indexOf(item);
      this.editedItem = Object.assign({}, item);
      this.dialog = true;
    },
    deleteItem(item) {
      const index = this.society.postMetadata.indexOf(item);
      confirm("Are you sure you want to delete this item?") && this.society.postMetadata.splice(index, 1);
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
        Object.assign(this.society.postMetadata[this.editedIndex], this.editedItem);
      } else {
        this.society.postMetadata.push(this.editedItem);
      }
      this.touch("postMetadata");
      this.close();
    },
    metadataDrag() {
      this.touch("postMetadata");
    }
  }
};
</script>
