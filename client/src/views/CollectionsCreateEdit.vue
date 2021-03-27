<template>
  <v-container class="collections-create">
    <h1>{{ collection.id ? "Edit" : "Create" }} Collection</h1>
    <v-btn
      v-if="collection.id"
      title="Use this collection as a template for a new collection"
      outlined
      color="primary"
      class="ml-4 mt-4 mb-4"
      :to="{ name: 'collection-edit', query: { clone: true } }"
    >
      Clone collection
    </v-btn>
    <v-form @submit.prevent="save">
      <h3>Give your collection a name (step 1 of 8)</h3>
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

      <h3>Select one or more categories (step 2 of 8)</h3>
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

      <h3 style="margin-top: 24px;">
        Privacy level
        <v-tooltip bottom maxWidth="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
          </template>
          <span>Do you want search results, record details, or images to be available to members only?</span>
        </v-tooltip>
        (step 3 of 8)
      </h3>
      <div class="citation">
        <v-select
          label="Privacy level"
          :items="privacyLevels"
          item-text="name"
          item-value="id"
          v-model="collection.privacyLevel"
          @input="touch('privacyLevel')"
        ></v-select>
      </div>

      <h3 style="margin-top: 16px;">What location does this collection cover? (step 4 of 8)</h3>
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

      <v-row no-gutters>
        <v-col cols="12">
          <h3 style="margin-top: 16px;">
            Enter spreadsheet columns
            <v-tooltip bottom maxWidth="600px">
              <template v-slot:activator="{ on, attrs }">
                <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
              </template>
              <span>"Spreadsheet headers" are the names of the columns in the CSV you will be uploading.</span>
              >
            </v-tooltip>
            (step 5 of 8)
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
                    <v-btn class="secondary primary--text mr-3" v-bind="attrs" v-on="on" small>Add a column</v-btn>
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
                            <v-text-field
                              dense
                              v-model="editedMappingItem.header"
                              label="Spreadsheet header"
                              placeholder="Column title in your spreadsheet"
                              @change="spreadsheetHeaderChanged"
                            >
                            </v-text-field>
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
                      <v-btn color="primary" @click="saveMapping">Save</v-btn>
                    </v-card-actions>
                  </v-card>
                </v-dialog>
              </v-toolbar>
            </template>
          </v-data-table>
          <!--end draggable mapping-->
        </v-col>
      </v-row>

      <h3 class="my-4">
        Column containing image file names
        <v-tooltip bottom maxWidth="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
          </template>
          <span>If the collection does not contain images, leave this blank</span>
        </v-tooltip>
        (step 6 of 8)
      </h3>
      <v-select
        outlined
        label="Image filename column"
        v-model="collection.imagePathHeader"
        :items="headers"
        @change="touch('imagePathHeader')"
        dense
      ></v-select>
      <div v-if="!isHeader(collection.imagePathHeader)" class="errorMessage">
        Column containing image file names no longer appears in the list of spreadsheet columns.
      </div>

      <h3 class="my-4">
        Columns containing household information
        <v-tooltip bottom maxWidth="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
          </template>
          <div>
            If the collection does not have households (multiple records that should be displayed together on the record
            detail page), leave these fields blank.
          </div>
          <div>
            <strong>Household number:</strong> Column containing the household number. All records with the same
            household number will be displayed together on the record detail page.
          </div>
          <div>
            <strong>Relationship to head:</strong> Column containing the record's relationship to the "head" of the
            household. Leave this blank if the records in the household are not in the same family. If the values in
            this column are "head", "father", "mother", "spouse", "husband", "wife", "child", "son", "daughter", then
            the corresponding relationships will be created when indexing the records.
          </div>
          <div>
            <strong>Gender:</strong> Column containing the gender. This is used to determine whether a father or mother
            relationship should be created when indexing the "head" record of a son, daughter, or child.
          </div>
        </v-tooltip>
        (step 7 of 8)
      </h3>
      <v-select
        outlined
        label="Household number column"
        v-model="collection.householdNumberHeader"
        :items="headers"
        @change="householdNumberChanged"
        dense
      ></v-select>
      <div v-if="!isHeader(collection.householdNumberHeader)" class="errorMessage">
        Column containing household numbers no longer appears in the list of spreadsheet columns.
      </div>
      <v-select
        outlined
        label="Relationship to head column"
        v-model="collection.householdRelationshipHeader"
        :items="headers"
        @change="touch('householdRelationshipHeader')"
        :disabled="!collection.householdNumberHeader"
        dense
      ></v-select>
      <div v-if="!isHeader(collection.householdRelationshipHeader)" class="errorMessage">
        Column containing relationship to head no longer appears in the list of spreadsheet columns.
      </div>
      <v-select
        outlined
        label="Gender column"
        v-model="collection.genderHeader"
        :items="headers"
        @change="touch('genderHeader')"
        :disabled="!collection.householdNumberHeader"
        dense
      ></v-select>
      <div v-if="!isHeader(collection.genderHeader)" class="errorMessage">
        Column containing gender no longer appears in the list of spreadsheet columns.
      </div>

      <h3>
        Citation template
        <v-tooltip bottom maxWidth="600px">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small>mdi-information</v-icon>
          </template>
          <span>You can include html and <span>{{</span>Spreadsheet header<span>}}</span> references</span>
        </v-tooltip>
        (step 8 of 8)
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

      <div v-if="warnChanges" class="errorMessage" style="margin-bottom: 16px">
        The changes made to the collection will not affect record sets that have already been uploaded, even if the
        record sets have not yet been published. If you want your changes to affect record sets that have already been
        uploaded, you need to delete the record sets and re-upload them after changing the collection.
      </div>
      <div class="d-flex justify-space-between">
        <v-btn
          type="submit"
          color="primary"
          :disabled="$v.$anyError || collection.mappings.length === 0 || !$v.$anyDirty"
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
          :title="
            postsForCollection.length > 0 ? 'Collections with record sets cannot be deleted' : 'Cannot be undone!'
          "
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
        <h3 class="mt-4" v-if="collection.id">Record sets</h3>
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

        <v-btn v-if="collection.id" outlined color="primary" class="mt-4" :to="{ name: 'posts-create' }">
          Create a new record set
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

function getContent(cid, next) {
  let routes = [store.dispatch("categoriesGetAll")];
  if (cid) {
    routes.push(store.dispatch("collectionsGetOne", cid));
    routes.push(store.dispatch("collectionsGetAll"));
    routes.push(store.dispatch("postsGetAll"));
  }
  Promise.all(routes)
    .then(() => {
      next();
    })
    .catch(() => {
      next("/");
    });
}

function setup(query) {
  this.collection = {
    ...this.collections.collection,
    categories: this.collections.collection.categories.map(catId =>
      this.categories.categoriesList.find(cat => cat.id === catId)
    ),
    mappings: lodash.cloneDeep(this.collections.collection.mappings)
  };
  if (this.collection.location) {
    this.locationItems = [this.collection.location];
  }
  this.collection.imagePathHeader = this.collection.imagePathHeader || "";
  this.collection.householdNumberHeader = this.collection.householdNumberHeader || "";
  this.collection.householdRelationshipHeader = this.collection.householdRelationshipHeader || "";
  this.collection.genderHeader = this.collection.genderHeader || "";
  this.editedFields.length = 0;
  if (query.clone) {
    this.collection.id = 0;
    this.collection.name += " (copy)";
  }
}

export default {
  components: { Multiselect, draggable },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    console.log("collectionsCreateEdit.beforeRouteEnter");
    getContent(routeTo.params.cid, next);
  },
  beforeRouteUpdate: function(routeTo, routeFrom, next) {
    console.log("collectionsCreateEdit.beforeRouteUpdate");
    getContent(routeTo.params.cid, next);
  },
  created() {
    if (this.$route.params && this.$route.params.cid) {
      setup.bind(this)(this.$route.query);
    }
    this.editedMappingItem = Object.assign({}, this.defaultMappingItem);
  },
  data() {
    return {
      dialogMapping: false,
      editedIndex: -1,
      editedMappingItem: {},
      defaultMappingItem: {
        header: "",
        dbField: "",
        ixRole: "principal",
        ixField: "na"
      },
      privacyLevels: [
        { id: 0, name: "Public" },
        { id: 1, name: "Public search results and record details; Members-only media" },
        { id: 3, name: "Public search results; Members-only record details and media" },
        { id: 7, name: "Members-only search results, record details, and media" }
      ],
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
        otherPlace: "Other Place",
        keywords: "Keywords"
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
        { value: "otherPlace", text: "Other Place" },
        { value: "keywords", text: "Keywords" }
      ],
      //end of data for table
      collection: {
        id: null,
        name: null,
        privacyLevel: 0,
        location: null,
        citation_template: null,
        categories: [],
        mappings: []
      },
      locationLoading: false,
      locationTimeout: null,
      locationItems: [],
      locationSearch: "",
      editedFields: [],
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
      ],
      warnChangesFields: [
        "categories",
        "privacyLevel",
        "location",
        "mappings",
        "imagePathHeader",
        "householdNumberHeader",
        "householdRelationshipHeader",
        "genderHeader"
      ]
    };
  },
  watch: {
    dialogMapping(val) {
      val || this.closeMapping();
    },
    locationSearch(val) {
      val && val !== this.collection.location && this.doLocationSearch(val);
    }
  },
  computed: {
    warnChanges() {
      return this.postsForCollection.length > 0 && this.warnChangesFields.some(fld => this.editedFields.includes(fld));
    },
    headers() {
      let headers = [{ text: "N/A", value: "" }].concat(
        this.collection.mappings.map(f => {
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
            postStatus: p.postStatus,
            recordsStatus: p.recordsKey ? p.recordsStatus || "Loaded" : "Missing",
            imagesStatus: !this.collection.imagePathHeader
              ? "N/A"
              : !!p.imagesKeys && p.imagesKeys.length > 0
              ? p.imagesStatus || "Loaded"
              : "Missing",
            collectionName: this.collections.collectionsList.find(coll => coll.id === p.collection).name,
            ...p.metadata
          };
        });
    },
    formTitle() {
      return this.editedIndex === -1 ? "New Spreadsheet Item" : "Edit Spreadsheet Item";
    },
    ...mapState(["collections", "categories", "posts", "societySummaries"])
  },
  validations: {
    collection: {
      name: { required },
      privacyLevel: {},
      location: {},
      citation_template: {},
      categories: { required },
      mappings: {},
      imagePathHeader: {},
      householdNumberHeader: {},
      householdRelationshipHeader: {},
      genderHeader: {}
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
    householdNumberChanged() {
      this.collection.householdRelationshipHeader = "";
      this.collection.genderHeader = "";
      this.touch("householdNumberHeader");
    },
    touch(attr) {
      this.editedFields.push(attr);
      if (this.$v.collection[attr].$dirty) {
        return;
      }
      let value = this.collection[attr];
      if (attr === "categories") {
        value = value.map(v => v.id);
      }
      if (attr === "mappings" || !this.collection.id || !lodash.isEqual(value, this.collections.collection[attr])) {
        this.$v.collection[attr].$touch();
      }
    },
    getPostColumns() {
      let cols = [
        {
          text: "Name",
          value: "name"
        },
        {
          text: "Status",
          value: "postStatus"
        },
        {
          text: "Records",
          value: "recordsStatus",
          align: "center"
        }
      ];
      if (this.collection.imagePathHeader) {
        cols.push({
          text: "Images",
          value: "imagesStatus",
          align: "center"
        });
      }
      cols.push({
        title: "Collection",
        field: "collectionName"
      });
      cols.push(...this.societySummaries.societySummary.postMetadata.map(pf => getMetadataColumn(pf)));
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
      this.collection.mappings = this.collection.mappings.filter(f => f.header);
      if (this.collection.mappings.length === 0) {
        return;
      }
      this.collection.fields = this.collection.mappings.map(m => ({ header: m.header }));
      if (
        !this.isHeader(this.imagePathHeader) ||
        !this.isHeader(this.householdNumberHeader) ||
        !this.isHeader(this.householdRelationshipHeader) ||
        !this.isHeader(this.genderHeader)
      ) {
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
    // methods to touch fields after they've been dragged
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
        this.touch("mappings");
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
      if (this.editedMappingItem.ixRole !== "na" && this.editedMappingItem.ixField === "na") {
        this.editedMappingItem.ixRole = "na";
      }
      if (this.editedIndex > -1) {
        Object.assign(this.collection.mappings[this.editedIndex], this.editedMappingItem);
      } else {
        this.collection.mappings.push(this.editedMappingItem);
      }
      this.touch("mappings");
      this.closeMapping();
    },
    ixRoleChanged() {
      if (this.editedMappingItem.ixRole === "na") {
        this.editedMappingItem.ixField = "na";
      }
    },
    spreadsheetHeaderChanged() {
      if (!this.editedMappingItem.dbField) {
        this.editedMappingItem.dbField = this.editedMappingItem.header;
      }
    }
  }
};
</script>

<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
<style>
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
.multiselect__option.multiselect__option--selected.multiselect__option--highlight {
  background: #b2ebf2;
  outline: none;
  color: #006064;
}
.multiselect__option--highlight:after {
  content: attr(data-select);
  background: #b2ebf2;
  color: #006064;
}
</style>
<style scoped>
.location {
  margin-bottom: 24px;
}
</style>
