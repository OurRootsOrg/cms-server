<template>
  <v-container class="collections-create">
    <h1>{{ collection.id ? "Edit" : "Create" }} Collection</h1>
    <v-form @submit.prevent="save">
      <h3>Give your collection a name (step 1 of 7)</h3>
      <v-text-field
        label="Collection Name"
        v-model="collection.name"
        type="text"
        placeholder="Name"
        :class="{ error: $v.collection.name.$error }"
        @change="touch('name')"
      ></v-text-field>

      <template v-if="$v.collection.name.$error">
        <p v-if="!$v.collection.name.required" class="errorMessage">
          Name is required.
        </p>
      </template>

      <h3>Select one or more categories (step 2 of 7)</h3>
      <multiselect
        v-model="collection.categories"
        :options="categories.categoriesList"
        :multiple="true"
        :searchable="true"
        :close-on-select="false"
        :clear-on-select="true"
        :preserve-search="false"
        :show-labels="true"
        :allow-empty="true"
        track-by="id"
        label="name"
        placeholder="Search or add a category"
        tag-placeholder="Add this category"
        :class="{ error: $v.collection.categories.$error }"
        @close="touch('categories')"
      ></multiselect>
      <template v-if="$v.collection.categories.$error">
        <p v-if="!$v.collection.categories.required" class="errorMessage">
          At least one category is required.
        </p>
      </template>

      <h3 style="margin-top: 16px;">What location does this collection cover? (step 3 of 7)</h3>
      <div class="location">
        <v-autocomplete
          outlined
          dense
          v-model="collection.location"
          :loading="locationLoading"
          :items="locationItems"
          :search-input.sync="locationSearch"
          no-filter
          auto-select-first
          flat
          hide-no-data
          hide-details
          solo
          @change="touch('location')"
          placeholder="Location"
          class="ma-0 mb-n2"
        ></v-autocomplete>
      </div>

      <h3>
        Citation template
        <v-tooltip bottom maxWidth="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
          </template>
          <span>You can include html and <span>{{</span>Spreadsheet header<span>}}</span> references</span>
        </v-tooltip>
        (step 4 of 7)
      </h3>
      <div class="citation">
        <v-textarea
          outlined
          name="input-7-4"
          v-model="collection.citation_template"
          @change="touch('citation_template')"
          placeholder="Citation template"
        ></v-textarea>
      </div>

      <v-row no-gutters class="mt-5">
        <v-col cols="12" class="mb-0">
          <h3>
            Define spreadsheet columns
            <v-tooltip bottom maxWidth="600px">
              <template v-slot:activator="{ on, attrs }">
                <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
              </template>
              <span
                >"Spreadsheet headers" are the names of the columns in the Excel or CSV you will be uploading.
                "Validation rules" are expressions defining the requirements for the spreadsheet data, and "Validation
                messages" are the error messages you want to show if the data does not meet the validation rules.</span
              >
            </v-tooltip>
            (step 5 of 7)
          </h3>
          <v-data-table
            :headers="fieldColumns"
            :items="collection.fields"
            item-key="id"
            :show-select="false"
            :disable-pagination="true"
            dense
            v-columns-resizable
          >
            <template v-slot:body>
              <draggable :list="collection.fields" tag="tbody" @change="columnDefsDrag">
                <tr v-for="(field, index) in collection.fields" :key="index">
                  <td><v-icon small>mdi-drag-horizontal-variant</v-icon></td>
                  <td>{{ field.header }}</td>
                  <td>
                    <span v-if="field.required"
                      ><v-icon class="green--text" small>mdi-check-circle</v-icon> Required</span
                    >
                  </td>
                  <td>{{ field.regex }}</td>
                  <td>{{ field.regexError }}</td>
                  <td>
                    <v-icon small @click="editColumnDefs(field)" class="mr-3">mdi-pencil</v-icon>
                    <v-icon small @click="deleteColumnDefs(field)">mdi-delete</v-icon>
                  </td>
                </tr>
              </draggable>
            </template>
            <template v-slot:footer>
              <v-toolbar flat class="ml-n3">
                <v-dialog v-model="dialogColumnDefs" max-width="600px">
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn class="secondary primary--text mr-3" v-bind="attrs" v-on="on" small>Add a row</v-btn>
                    <span v-if="collection.fields.length === 0"
                      >(you need at least one row defining at least one column in your spreadsheet)</span
                    >
                  </template>
                  <v-card>
                    <v-card-title class="pb-5 mb-0"> {{ formTitle }}</v-card-title>
                    <v-card-text>
                      <v-container class="pl-0">
                        <v-row>
                          <v-col cols="12" sm="7">
                            <v-text-field
                              dense
                              v-model="editedItem.header"
                              label="Spreadsheet header"
                              placeholder="Column title in your spreadsheet"
                            ></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="5">
                            <v-checkbox
                              dense
                              class="pt-0 mt-1"
                              v-model="editedItem.required"
                              label="Required"
                            ></v-checkbox>
                          </v-col>
                          <v-col cols="12">
                            <v-textarea
                              dense
                              outlined
                              rows="2"
                              v-model="editedItem.regex"
                              label="Validation rule (optional)"
                              placeholder="Regex to validate data. For help with regular expressions see http://regex101.com/"
                            ></v-textarea>
                          </v-col>
                          <v-col cols="12">
                            <v-text-field
                              dense
                              v-model="editedItem.regexError"
                              label="Validation message (optional)"
                              placeholder="Error message if validation fails"
                            ></v-text-field>
                          </v-col>
                        </v-row>
                      </v-container>
                    </v-card-text>
                    <v-card-actions class="pb-5 pr-5">
                      <v-spacer></v-spacer>
                      <v-btn color="primary" text @click="closeColumnDefs" class="mr-5">Cancel</v-btn>
                      <v-btn color="primary" @click="saveColumnDefs">Save</v-btn>
                    </v-card-actions>
                  </v-card>
                </v-dialog>
              </v-toolbar>
            </template>
          </v-data-table>
        </v-col>
      </v-row>

      <v-row no-gutters>
        <v-col class="mt-0">
          <h3>
            Define how spreadsheet data is displayed and indexed
            <v-tooltip bottom maxWidth="600px">
              <template v-slot:activator="{ on, attrs }">
                <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
              </template>
              <span
                >This mapping determines how your spreadsheet's columns and data will be indexed and shown in search
                results.</span
              >
            </v-tooltip>
            (step 6 of 7)
          </h3>

          <!--draggable mappings-->
          <p class="caption">
            Hint: use the <v-icon small>mdi-drag-horizontal-variant</v-icon> handles to drag rows up and down the list
            to put them in whatever order you want them to show on the record detail page
          </p>
          <v-data-table
            :headers="mappingColumns"
            :items="collection.mappings"
            item-key="id"
            :show-select="false"
            :disable-pagination="true"
            dense
            v-columns-resizable
          >
            <template v-slot:body>
              <draggable :list="collection.mappings" tag="tbody" @change="mappingDrag">
                <tr v-for="(field, index) in collection.mappings" :key="index">
                  <td><v-icon small>mdi-drag-horizontal-variant</v-icon></td>
                  <td>{{ field.header }}</td>
                  <td>{{ field.dbField }}</td>
                  <td>{{ ixRoleMap[field.ixRole] }}</td>
                  <td>{{ ixFieldMap[field.ixField] }}</td>
                  <td>
                    <v-icon small @click="editMapping(field)" class="mr-3">mdi-pencil</v-icon>
                    <v-icon small @click="deleteMapping(field)">mdi-delete</v-icon>
                  </td>
                </tr>
              </draggable>
            </template>
            <template v-slot:footer>
              <v-toolbar flat class="ml-n3">
                <v-dialog v-model="dialogMapping" max-width="600px">
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn class="secondary primary--text mr-3" v-bind="attrs" v-on="on" small>Add a row</v-btn>
                    <span v-if="collection.mappings.length === 0"
                      >(you need at least one row defining at least one column in your spreadsheet)</span
                    >
                  </template>
                  <v-card>
                    <v-card-title class="pb-5 mb-0"> {{ formTitle }}</v-card-title>
                    <v-card-text>
                      <v-container class="pl-0">
                        <v-row>
                          <v-col cols="12" sm="6">
                            <v-select
                              v-model="editedMappingItem.header"
                              label="Spreadsheet header"
                              :items="spreadsheetColumnHeaders"
                              dense
                            >
                            </v-select>
                          </v-col>
                          <v-col cols="12" sm="6">
                            <v-text-field
                              dense
                              v-model="editedMappingItem.dbField"
                              label="Label shown on record detail page"
                              placeholder="Label for record detail page"
                            >
                              <v-tooltip bottom slot="append" maxWidth="600px">
                                <template v-slot:activator="{ on, attrs }">
                                  <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
                                </template>
                                <span>Leave blank to omit this field from the record detail page.</span>
                              </v-tooltip>
                            </v-text-field>
                          </v-col>
                          <v-col cols="12">
                            <v-select
                              v-model="editedMappingItem.ixRole"
                              label="Relationship of this information to the primary search person"
                              :items="ixRoleMapOptions"
                              @change="ixRoleChanged"
                            >
                              <v-tooltip bottom slot="append" maxWidth="600px">
                                <template v-slot:activator="{ on, attrs }">
                                  <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
                                </template>
                                <span
                                  >Select the relationship of this field to the primary person (principal) in the
                                  record. This affects how the information will be indexed for search.</span
                                >
                              </v-tooltip>
                            </v-select>
                          </v-col>
                          <v-col cols="12">
                            <v-select
                              v-model="editedMappingItem.ixField"
                              label="Index field"
                              :items="ixFieldMapOptions"
                              :disabled="editedMappingItem.ixRole === 'na'"
                            ></v-select>
                          </v-col>
                        </v-row>
                      </v-container>
                    </v-card-text>
                    <v-card-actions class="pb-5 pr-5">
                      <v-spacer></v-spacer>
                      <v-btn color="primary" text @click="closeMapping" class="mr-5">Cancel</v-btn>
                      <v-btn
                        color="primary"
                        @click="saveMapping"
                        :disabled="editedMappingItem.ixRole !== 'na' && editedMappingItem.ixField === 'na'"
                        >Save</v-btn
                      >
                    </v-card-actions>
                  </v-card>
                </v-dialog>
              </v-toolbar>
            </template>
          </v-data-table>
          <!--end draggable mapping-->
        </v-col>
      </v-row>

      <h3 class="mt-4">
        Column containing image file names
        <v-tooltip bottom maxWidth="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
          </template>
          <span>If the collection does not contain images, leave this blank</span>
        </v-tooltip>
        (step 7 of 7)
      </h3>
      <v-select
        outlined
        v-model="collection.imagePathHeader"
        :items="headers"
        @change="touch('imagePathHeader')"
      ></v-select>
      <div v-if="!isHeader(collection.imagePathHeader)" class="errorMessage">
        Column containing image file names no longer appears in the list of spreadsheet columns.
      </div>

      <div class="d-flex justify-space-between">
        <v-btn
          type="submit"
          color="primary"
          :disabled="
            $v.$anyError || collection.fields.length === 0 || collection.mappings.length === 0 || !$v.$anyDirty
          "
        >
          <v-icon left small>
            mdi-alert
          </v-icon>
          Important: Save all changes
        </v-btn>
        <v-btn
          v-if="collection.id"
          class="mt-2"
          buttonClass="danger"
          :title="postsForCollection.length > 0 ? 'Collections with posts cannot be deleted' : 'Cannot be undone!'"
          @click="del()"
          :disabled="postsForCollection.length > 0"
          >Delete Collection
        </v-btn>
        <span v-if="$v.$anyError" class="red--text">
          Please fill out the required field(s).
        </span>
      </div>
    </v-form>

    <v-row class="pt-5">
      <v-col>
        <h3 class="mt-4" v-if="collection.id">Posts</h3>
        <v-data-table
          v-if="collection.id"
          :items="postsForCollection"
          :headers="getPostColumns()"
          @click:row="postRowClicked"
          :footer-props="{
            'items-per-page-options': [10, 25, 50]
          }"
          :items-per-page="25"
          dense
          class="rowHover"
        >
          <template v-slot:[`item.hasData`]="{ item }">
            <v-icon v-if="item.hasData" class="green--text">mdi-checkbox-marked</v-icon>
            <v-icon v-else class="red--text">mdi-close-circle</v-icon>
          </template>
        </v-data-table>

        <v-btn v-if="collection.id" outlined color="primary" class="mt-4" to="/posts/create">
          Create a new post
        </v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import NProgress from "nprogress";
import { required } from "vuelidate/lib/validators";
import Multiselect from "vue-multiselect";
import lodash from "lodash";
import draggable from "vuedraggable";
import Server from "@/services/Server";
import { getMetadataColumn } from "../utils/metadata";

function setup() {
  this.collection = {
    ...this.collections.collection,
    categories: this.collections.collection.categories.map(catId =>
      this.categories.categoriesList.find(cat => cat.id === catId)
    ),
    fields: lodash.cloneDeep(this.collections.collection.fields),
    mappings: lodash.cloneDeep(this.collections.collection.mappings)
  };
  if (this.collection.location) {
    this.locationItems = [this.collection.location];
  }
  if (!this.collection.imagePathHeader) {
    this.collection.imagePathHeader = "";
  }
}

export default {
  components: { Multiselect, draggable },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    let routes = [store.dispatch("categoriesGetAll")];
    if (routeTo.params && routeTo.params.cid) {
      routes.push(store.dispatch("collectionsGetOne", routeTo.params.cid));
      routes.push(store.dispatch("collectionsGetAll"));
      routes.push(store.dispatch("postsGetAll"));
      routes.push(store.dispatch("settingsGet"));
    }
    Promise.all(routes)
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  created() {
    if (this.$route.params && this.$route.params.cid) {
      setup.bind(this)();
    }
  },
  data() {
    return {
      dialogColumnDefs: false,
      dialogMapping: false,
      editedIndex: -1,
      editedItem: {
        header: "",
        required: false,
        regex: "",
        regexError: ""
      },
      editedMappingItem: {
        header: "",
        dbField: "",
        ixRole: "",
        ixField: ""
      },
      defaultItem: {
        header: "",
        required: "",
        regex: "",
        regexError: ""
      },
      defaultMappingItem: {
        header: "",
        dbField: "",
        ixRole: "Don't index",
        ixField: "Don't index"
      },
      ixRoleMap: {
        na: "Don't index",
        principal: "Principal",
        father: "Father",
        mother: "Mother",
        spouse: "Spouse",
        bride: "Bride",
        groom: "Groom",
        brideFather: "Father of the bride",
        brideMother: "Mother of the bride",
        groomFather: "Father of the groom",
        groomMother: "Mother of the groom",
        other: "Other person"
      },
      ixFieldMap: {
        na: "Don't index",
        given: "Given name",
        surname: "Surname",
        birthDate: "Birth Date",
        birthPlace: "Birth Place",
        marriageDate: "Marriage Date",
        marriagePlace: "Marriage Place",
        deathDate: "Death Date",
        deathPlace: "Death Place",
        residenceDate: "Residence Date",
        residencePlace: "Residence Place",
        otherDate: "Other Date",
        otherPlace: "Other Place"
      },
      //do it like this [{value: true, text: "Has data"}, {value: false, text: "No data"}]
      ixRoleMapOptions: [
        { value: "na", text: "Don't index" },
        { value: "principal", text: "Principal" },
        { value: "father", text: "Father" },
        { value: "mother", text: "Mother" },
        { value: "spouse", text: "Spouse" },
        { value: "bride", text: "Bride" },
        { value: "groom", text: "Groom" },
        { value: "brideFather", text: "Father of the bride" },
        { value: "brideMother", text: "Mother of the bride" },
        { value: "groomFather", text: "Father of the groom" },
        { value: "groomMother", text: "Mother of the groom" },
        { value: "other", text: "Other person" }
      ],
      ixFieldMapOptions: [
        { value: "na", text: "Don't index" },
        { value: "given", text: "Given name" },
        { value: "surname", text: "Surname" },
        { value: "birthDate", text: "Birth Date" },
        { value: "birthPlace", text: "Birth Place" },
        { value: "marriageDate", text: "Marriage Date" },
        { value: "marriagePlace", text: "Marriage Place" },
        { value: "deathDate", text: "Death Date" },
        { value: "deathPlace", text: "Death Place" },
        { value: "residenceDate", text: "Residence Date" },
        { value: "residencePlace", text: "Residence Place" },
        { value: "otherDate", text: "Other Date" },
        { value: "otherPlace", text: "Other Place" }
      ],
      spreadsheetColumnOptions: [{ value: "get this from the columns", text: "Spreadsheet column" }],
      //end of data for experimental table; keep everything after this
      collection: {
        id: null,
        name: null,
        location: null,
        citation_template: null,
        categories: [],
        fields: [],
        mappings: []
      },
      locationLoading: false,
      locationTimeout: null,
      locationItems: [],
      locationSearch: "",
      fieldColumns: [
        {
          text: "",
          value: "handle",
          align: "left",
          width: 10
        },
        {
          text: "Spreadsheet header",
          value: "header"
        },
        {
          text: "Required?",
          value: "required",
          align: "center"
        },
        {
          text: "Validation rule",
          value: "regex"
        },
        {
          text: "Validation Message",
          value: "regexError"
        },
        {
          text: "",
          value: "actions",
          align: "right"
        }
      ],
      mappingColumns: [
        {
          text: "",
          value: "handle",
          align: "left",
          width: 20
        },
        {
          text: "Spreadsheet header",
          value: "header"
        },
        {
          text: "Record detail page label",
          value: "dbField"
        },
        {
          text: "Index Role",
          value: "ixRole"
        },
        {
          text: "Index Field",
          value: "ixField"
        },
        {
          text: "",
          value: "actions",
          align: "right"
        }
      ]
    };
  },
  watch: {
    dialogColumnDefs(val) {
      val || this.closeColumnDefs();
    },
    dialogMapping(val) {
      val || this.closeMapping();
    },
    locationSearch(val) {
      val && val !== this.collection.location && this.doLocationSearch(val);
    }
  },
  computed: {
    headers() {
      let headers = [{ text: "N/A", value: "" }].concat(
        this.collection.fields.map(f => {
          return {
            get text() {
              return f.header;
            },
            get value() {
              return f.header;
            }
          };
        })
      );
      return headers;
    },
    postsForCollection() {
      return this.posts.postsList
        .filter(p => p.collection === this.collection.id)
        .map(p => {
          return {
            id: p.id,
            name: p.name,
            recordsStatus: p.imagesStatus === "Loading" ? p.imagesStatus : p.recordsStatus,
            hasData: !!p.recordsKey,
            hasImages: !!p.imagesKeys && p.imagesKeys.length > 0,
            collectionName: this.collections.collectionsList.find(coll => coll.id === p.collection).name,
            ...p.metadata
          };
        });
    },
    spreadsheetColumnHeaders() {
      return this.collection.fields.map(f => f.header);
    },
    formTitle() {
      return this.editedIndex === -1 ? "New Spreadsheet Item" : "Edit Spreadsheet Item";
    },
    ...mapState(["collections", "categories", "posts", "settings"])
  },
  validations: {
    collection: {
      name: { required },
      location: {},
      citation_template: {},
      categories: { required },
      fields: {},
      mappings: {},
      imagePathHeader: {}
    }
  },
  methods: {
    doLocationSearch(text) {
      if (this.locationTimeout) {
        clearTimeout(this.locationTimeout);
      }
      this.locationLoading = true;
      this.locationTimeout = setTimeout(() => {
        this.locationTimeout = null;
        Server.placeSearch(text)
          .then(res => {
            this.locationItems = res.data.map(p => p.fullName);
          })
          .finally(() => {
            this.locationLoading = false;
          });
      }, 400);
    },
    touch(attr) {
      if (this.$v.collection[attr].$dirty) {
        return;
      }
      let value = this.collection[attr];
      if (attr === "categories") {
        value = value.map(v => v.id);
      }
      if (
        attr === "mappings" ||
        attr === "fields" ||
        !this.collection.id ||
        !lodash.isEqual(value, this.collections.collection[attr])
      ) {
        this.$v.collection[attr].$touch();
      }
    },
    syncFieldsMappings(newValue, oldValue) {
      if (newValue) {
        if (oldValue) {
          this.collection.mappings.forEach(m => {
            if (m.header === oldValue) {
              m.header = newValue;
            }
          });
        } else {
          this.collection.mappings.push({ header: newValue, dbField: newValue, ixRole: "na", ixField: "na" });
        }
      } else {
        this.collection.mappings = this.collection.mappings.filter(m => m.header !== oldValue);
      }
      this.touch("mappings");
    },
    getPostColumns() {
      let cols = [
        {
          text: "Name",
          value: "name"
        },
        {
          text: "Status",
          value: "recordsStatus"
        },
        {
          text: "Has Data",
          value: "hasData",
          align: "center"
        },
        {
          title: "Has Images",
          field: "hasImages",
          align: "center"
        },
        {
          title: "Collection",
          field: "collectionName"
        }
      ];
      cols.push(...this.settings.settings.postMetadata.map(pf => getMetadataColumn(pf)));
      return cols;
    },
    postRowClicked(post) {
      this.$router.push({
        name: "post-edit",
        params: { pid: post.id }
      });
    },
    isHeader(value) {
      if (!value) value = "";
      return this.headers.findIndex(h => h.value === value) >= 0;
    },
    save() {
      this.collection.fields = this.collection.fields.filter(f => f.header);
      if (this.collection.fields.length === 0) {
        return;
      }
      this.collection.mappings = this.collection.mappings.filter(f => f.header);
      if (this.collection.mappings.length === 0) {
        return;
      }
      if (!this.isHeader(this.imagePathHeader)) {
        return;
      }
      let collection = Object.assign({}, this.collection);
      collection.categories = collection.categories.map(cat => cat.id);
      this.$v.$touch();
      if (!this.$v.$invalid) {
        NProgress.start();
        this.$store
          .dispatch(collection.id ? "collectionsUpdate" : "collectionsCreate", collection)
          .then(result => {
            if (collection.id) {
              setup.bind(this)();
              this.$v.$reset();
              NProgress.done();
            } else {
              this.$router.push({
                name: "collection-edit",
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
      if (this.postsForCollection.length > 0) {
        return;
      }
      NProgress.start();
      this.$store
        .dispatch("collectionsDelete", this.collection.id)
        .then(() => {
          this.$router.push({
            name: "collections-list"
          });
        })
        .catch(() => {
          NProgress.done();
        });
    },
    //methods for the spreadsheet columns table
    editColumnDefs(item) {
      this.editedIndex = this.collection.fields.indexOf(item);
      this.editedItem = Object.assign({}, item);
      this.dialogColumnDefs = true;
    },
    deleteColumnDefs(item) {
      const index = this.collection.fields.indexOf(item);
      if (confirm("Are you sure you want to delete this item?")) {
        this.collection.fields.splice(index, 1);
        this.syncFieldsMappings(null, item.header);
      }
    },
    closeColumnDefs() {
      this.dialogColumnDefs = false;
      this.$nextTick(() => {
        this.editedItem = Object.assign({}, this.defaultItem);
        this.editedIndex = -1;
      });
    },
    saveColumnDefs() {
      if (this.editedIndex > -1) {
        if (this.editedItem.header !== this.collection.fields[this.editedIndex].header) {
          this.syncFieldsMappings(this.editedItem.header, this.collection.fields[this.editedIndex].header);
        }
        Object.assign(this.collection.fields[this.editedIndex], this.editedItem);
        this.touch("fields");
      } else {
        this.collection.fields.push(this.editedItem);
        this.syncFieldsMappings(this.editedItem.header, null);
      }
      this.closeColumnDefs();
    },
    // methods to touch fields after they've been dragged
    columnDefsDrag() {
      this.touch("fields");
    },
    mappingDrag() {
      this.touch("mappings");
    },
    //methods for the mappings table
    editMapping(item) {
      this.editedIndex = this.collection.mappings.indexOf(item);
      this.editedMappingItem = Object.assign({}, item);
      this.dialogMapping = true;
    },
    deleteMapping(item) {
      const index = this.collection.mappings.indexOf(item);
      if (confirm("Are you sure you want to delete this item?")) {
        this.collection.mappings.splice(index, 1);
        this.syncFieldsMappings(null, item.header);
      }
    },
    closeMapping() {
      this.dialogMapping = false;
      this.$nextTick(() => {
        this.editedMappingItem = Object.assign({}, this.defaultMappingItem);
        this.editedIndex = -1;
      });
    },
    saveMapping() {
      if (this.editedIndex > -1) {
        Object.assign(this.collection.mappings[this.editedIndex], this.editedMappingItem);
        this.touch("fields");
      } else {
        this.collection.mappings.push(this.editedMappingItem);
      }
      this.closeMapping();
    },
    ixRoleChanged() {
      if (this.editedMappingItem.ixRole === "na") {
        this.editedMappingItem.ixField = "na";
      }
    }
  }
};
</script>

<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
<style scoped>
.multiselect__tag {
  color: #006064;
  line-height: 1;
  background: #b2ebf2;
}
.multiselect__option--highlight {
  background: #b2ebf2;
  outline: none;
  color: #006064;
}
.multiselect__option--highlight:after {
  content: attr(data-select);
  background: #b2ebf2;
  color: #006064;
}
.spreadsheetColumnsTable >>> table > tbody > tr > td:nth-child(1),
.spreadsheetColumnsTable >>> table > thead > tr > th:nth-child(1) {
  left: 0;
}
.spreadsheetColumnsTable >>> table > tbody > tr > td:nth-child(2),
.spreadsheetColumnsTable >>> table > thead > tr > th:nth-child(2) {
  left: 50px;
}
.spreadsheetColumnsTable >>> table > tbody > tr > td:nth-child(3),
.spreadsheetColumnsTable >>> table > thead > tr > th:nth-child(3) {
  left: 140px;
}
.spreadsheetColumnsTable >>> table > tbody > tr > td:nth-child(4),
.spreadsheetColumnsTable >>> table > thead > tr > th:nth-child(4) {
  left: 260px;
}
.spreadsheetColumnsTable >>> table > thead > tr > th:nth-child(1)
/* .spreadsheetColumnsTable >>> table > thead > tr > th:nth-child(2) */
 {
  position: sticky !important;
  position: -webkit-sticky !important;
  /* z-index: 9999; */
  background: white;
}
.spreadsheetColumnsTable >>> table > tbody > tr > td:nth-child(1)
/* .spreadsheetColumnsTable >>> table > tbody > tr > td:nth-child(2) */
 {
  position: sticky !important;
  position: -webkit-sticky !important;
  /* z-index: 9998; */
  background: white;
}
.spreadsheetColumnsTable >>> table > tbody > tr > td:nth-child(1):hover {
  background-color: #efefef;
}

.spreadsheetColumnsTable >>> table > tbody > tr > td {
  padding: 0 8px;
}
.spreadsheetColumnsTable >>> thead .text-start {
  vertical-align: top;
  text-align: left;
  padding-left: 8px;
}
.spreadsheetColumnsTable >>> thead .sortable {
  vertical-align: top;
  text-align: left;
  padding-left: 8px;
}
.spreadsheetColumnsTable >>> .table-header-group {
  vertical-align: top;
  text-align: left;
  padding-left: 8px;
}
.location {
  margin-bottom: 24px;
}
</style>
<!--the original green hex #41b883 change to cyan lighten-3 #80DEEA or cyan lighten-4 #B2EBF2-->
