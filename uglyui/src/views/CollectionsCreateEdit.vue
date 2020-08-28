<template>
  <v-container class="collections-create">
    <v-layout>
      <h1>{{ collection.id ? "Edit" : "Create" }} Collection</h1>
    </v-layout>

    <v-form @submit.prevent="save">
      <h3>Give your collection a name</h3>
      <v-text-field
        label="Collection Name"
        v-model="collection.name"
        type="text"
        placeholder="Name"
        class="field"
        :class="{ error: $v.collection.name.$error }"
        @change="touch('name')"
      ></v-text-field>

      <template v-if="$v.collection.name.$error">
        <p v-if="!$v.collection.name.required" class="errorMessage">
          Name is required.
        </p>
      </template>

      <h3>What location does this collection cover?</h3>
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

      <h3>Citation template (html and <span>{{</span>Spreadsheet header<span>}}</span> references allowed)</h3>
      <div class="citation">
        <v-textarea
          outlined
          name="input-7-4"
          v-model="collection.citation_template"
          @change="touch('citation_template')"
          placeholder="Citation template"
        ></v-textarea>
      </div>

      <h3>Select one or more categories</h3>
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

      <h3 class="mt-4">Define spreadsheet columns</h3>
      <Tabulator
        :data="collection.fields"
        :columns="fieldColumns"
        layout="fitColumns"
        :movable-rows="true"
        :resizable-columns="true"
        @rowMoved="fieldsMoved"
        @cellEdited="fieldsEdited"
      />
      <v-btn small color="primary" class="mt-2" href="" @click.prevent="addField">Add a row</v-btn>
      <span v-if="collection.fields.length === 0">
        (you need at least one)
      </span>

      <h3 class="mt-4">Define how spreadsheet columns are displayed and indexed</h3>
      <Tabulator
        :data="collection.mappings"
        :columns="mappingColumns"
        layout="fitColumns"
        :movable-rows="true"
        :resizable-columns="true"
        @rowMoved="mappingMoved"
        @cellEdited="mappingEdited"
      />

      <v-btn small color="primary" class="mt-2" href="" @click.prevent="addMapping">Add a row</v-btn>
      <span v-if="collection.mappings.length === 0">
        (you need at least one)
      </span>

      <h3 class="mt-4">Column containing image file names (if any)</h3>
      <v-select
        outlined
        v-model="collection.imagePathHeader"
        :items="headers"
        @change="touch('imagePathHeader')"
      ></v-select>
      <div v-if="!isHeader(collection.imagePathHeader)" class="errorMessage">
        Column containing image file names no longer appears in the list of spreadsheet columns.
      </div>

      <v-layout row>
        <v-flex class="ma-3">
          <v-btn
            type="submit"
            color="primary"
            class="mt-4"
            :disabled="
              $v.$anyError || collection.fields.length === 0 || collection.mappings.length === 0 || !$v.$anyDirty
            "
            >Save
          </v-btn>
          <span v-if="$v.$anyError" class="red--text">
            Please fill out the required field(s).
          </span>
        </v-flex>
      </v-layout>
    </v-form>

    <v-btn
      v-if="collection.id"
      class="mt-2"
      buttonClass="danger"
      :title="postsForCollection.length > 0 ? 'Collections with posts cannot be deleted' : 'Cannot be undone!'"
      @click="del()"
      :disabled="postsForCollection.length > 0"
      >Delete Collection
    </v-btn>

    <h3 class="mt-4" v-if="collection.id">Posts</h3>
    <Tabulator
      v-if="collection.id"
      :data="postsForCollection"
      :columns="getPostColumns()"
      layout="fitColumns"
      :header-sort="true"
      :selectable="true"
      :resizable-columns="true"
      @rowClicked="postRowClicked"
    />
    <v-btn v-if="collection.id" outlined color="primary" class="mt-4" to="/posts/create">
      Create a new post
    </v-btn>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import Tabulator from "../components/Tabulator";
import { getMetadataColumn } from "../utils/metadata";
import NProgress from "nprogress";
import { required } from "vuelidate/lib/validators";
import Multiselect from "vue-multiselect";
import lodash from "lodash";
import Server from "@/services/Server";

const ixRoleMap = {
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
};

const ixFieldMap = {
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
};

const ixEmptyFieldMap = {
  na: "Don't index"
};

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
  components: { Tabulator, Multiselect },
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
          rowHandle: true,
          formatter: "handle",
          headerSort: false,
          frozen: true,
          width: 30,
          minWidth: 30
        },
        {
          title: "Spreadsheet header",
          widthGrow: 2,
          field: "header",
          tooltip: "spreadsheet column header (required)",
          editor: "input",
          validator: ["unique"]
        },
        {
          title: "Required?",
          field: "required",
          tooltip: "is this field required?",
          editor: "tickCross",
          hozAlign: "center",
          formatter: "tickCross",
          formatterParams: { allowEmpty: true }
        },
        {
          title: "Validation rule",
          widthGrow: 2,
          field: "regex",
          tooltip: "regular expression used to validate column values (optional)",
          editor: "input"
        },
        {
          title: "Validation Message",
          widthGrow: 2,
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
            this.fieldsDelete(cell.getRow().getPosition());
          }
        }
      ],
      mappingColumns: [
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
          widthGrow: 2,
          field: "header",
          tooltip: "spreadsheet column header from table above (required)",
          editor: "select",
          editorParams: () => {
            return {
              values: this.collection.fields.map(f => f.header),
              verticalNavigation: "table"
            };
          },
          validator: ["required"]
        },
        {
          title: "Record detail field label",
          widthGrow: 2,
          field: "dbField",
          tooltip: "name of the field when displaying the record detail (don't display if empty)",
          editor: "input"
        },
        {
          title: "Index Role",
          field: "ixRole",
          tooltip: "whether to index this field for the principal or another person in the record (optional)",
          formatter: "lookup",
          formatterParams: ixRoleMap,
          editor: "select",
          editorParams: {
            values: ixRoleMap,
            defaultValue: "na"
          },
          validator: ["required"]
        },
        {
          title: "Index Field",
          field: "ixField",
          tooltip: "how to index this field (optional)",
          formatter: "lookup",
          formatterParams: ixFieldMap,
          editor: "select",
          editorParams: cell => {
            let ixRole = cell
              .getRow()
              .getCell("ixRole")
              .getValue();
            return {
              values: !ixRole || ixRole === "na" ? ixEmptyFieldMap : ixFieldMap
            };
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
            this.mappingDelete(cell.getRow().getPosition());
          }
        }
      ]
    };
  },
  watch: {
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
            recordsStatus: p.recordsStatus,
            hasData: !!p.recordsKey,
            collectionName: this.collections.collectionsList.find(coll => coll.id === p.collection).name,
            ...p.metadata
          };
        });
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
      if (!this.collection.id || !lodash.isEqual(value, this.collections.collection[attr])) {
        this.$v.collection[attr].$touch();
      }
    },
    addField() {
      this.collection.fields.push({});
    },
    fieldsMoved(data) {
      this.collection.fields = data;
      this.touch("fields");
    },
    fieldsDelete(ix) {
      let header = this.collection.fields[ix].header;
      this.collection.fields.splice(ix, 1);
      this.syncFieldsMappings(null, header);
      this.touch("fields");
    },
    fieldsEdited(cell) {
      if (cell.getField() === "header") {
        this.syncFieldsMappings(cell.getValue(), cell.getOldValue());
      }
      this.touch("fields");
    },
    addMapping() {
      this.collection.mappings.push({ ixRole: "na", ixField: "na" });
    },
    mappingMoved(data) {
      this.collection.mappings = data;
      this.touch("mappings");
    },
    mappingDelete(ix) {
      this.collection.mappings.splice(ix, 1);
      this.touch("mappings");
    },
    mappingEdited(cell) {
      if (cell.getField() === "ixRole") {
        if (cell.getValue() === "" || cell.getValue() === "na") {
          cell
            .getRow()
            .getCell("ixField")
            .setValue("na", true);
        }
      }
      this.touch("mappings");
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
          title: "Name",
          field: "name",
          headerFilter: "input",
          sorter: "string"
        },
        {
          title: "Status",
          field: "recordsStatus",
          headerFilter: "select",
          headerFilterParams: {
            values: true
          },
          sorter: "string"
        },
        {
          title: "Has Data",
          field: "hasData",
          hozAlign: "center",
          formatter: "tickCross",
          headerFilter: "tickCross",
          sorter: "boolean"
        },
        {
          title: "Collection",
          field: "collectionName",
          headerFilter: "input",
          sorter: "string"
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
.location {
  margin-bottom: 24px;
}
</style>
<!--the original green hex #41b883 change to cyan lighten-3 #80DEEA or cyan lighten-4 #B2EBF2-->
