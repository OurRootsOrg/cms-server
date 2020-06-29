<template>
  <div class="collections-create">
    <h1>{{ collection.id ? "Edit" : "Create" }} Collection</h1>
    <form @submit.prevent="save">
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

      <h3>Select one or more categories</h3>
      <label>Categories (select one or more)</label>
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
        @close="$v.collection.categories.$touch()"
      ></multiselect>
      <template v-if="$v.collection.categories.$error">
        <p v-if="!$v.collection.categories.required" class="errorMessage">
          At least one category is required.
        </p>
      </template>

      <h3>Define spreadsheet columns</h3>
      <Tabulator
        :data="collection.fields"
        :columns="fieldColumns"
        layout="fitColumns"
        :movable-rows="true"
        :resizable-columns="true"
        @rowMoved="fieldsMoved"
        @cellEdited="fieldsEdited"
      />
      <a href="" @click.prevent="addField">Add a row</a>
      <span v-if="collection.fields.length === 0">
        (you need at least one)
      </span>

      <h3>Define how spreadsheet columns are displayed and indexed</h3>
      <Tabulator
        :data="collection.mappings"
        :columns="mappingColumns"
        layout="fitColumns"
        :movable-rows="true"
        :resizable-columns="true"
        @rowMoved="mappingMoved"
        @cellEdited="mappingEdited"
      />
      <a href="" @click.prevent="addMapping">Add a row</a>
      <span v-if="collection.mappings.length === 0">
        (you need at least one)
      </span>

      <BaseButton
        type="submit"
        class="submit-button"
        buttonClass="-fill-gradient"
        :disabled="$v.$anyError || collection.fields.length === 0 || collection.mappings.length === 0"
        >Save</BaseButton
      >
      <p v-if="$v.$anyError" class="errorMessage">
        Please fill out the required field(s).
      </p>
    </form>
    <div v-if="collection.id" class="posts">Collection has {{ posts.length }} posts</div>
    <BaseButton class="btn" buttonClass="danger" @click="del()" :disabled="posts.length > 0"
      >Delete Collection</BaseButton
    >
  </div>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import Tabulator from "../components/Tabulator";
import NProgress from "nprogress";
import { required } from "vuelidate/lib/validators";
import Multiselect from "vue-multiselect";

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
  Object.assign(this.collection, this.collections.collection);
  this.collection.categories = this.collection.categories.map(catId =>
    this.$store.state.categories.categoriesList.find(cat => cat.id === catId)
  );
  this.posts = this.$store.state.posts.postsList.filter(p => p.collection === this.collection.id);
}

export default {
  components: { Tabulator, Multiselect },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    let routes = [store.dispatch("categoriesGetAll")];
    if (routeTo.params && routeTo.params.cid) {
      routes.push(store.dispatch("collectionsGetOne", routeTo.params.cid));
      routes.push(store.dispatch("postsGetAll"));
    }
    Promise.all(routes).then(() => {
      next();
    });
  },
  created() {
    if (this.$route.params && this.$route.params.cid) {
      setup.bind(this)();
    }
  },
  data() {
    return {
      collection: { categories: [], fields: [], mappings: [] },
      posts: [],
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
          formatterParams: { allowEmpty: true, crossElement: "&ndash;" }
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
  computed: mapState(["collections", "categories"]),
  validations: {
    collection: {
      name: { required },
      categories: { required }
    }
  },
  methods: {
    addField() {
      this.collection.fields.push({});
    },
    fieldsMoved(data) {
      this.collection.fields = data;
    },
    fieldsDelete(ix) {
      let header = this.collection.fields[ix].header;
      this.collection.fields.splice(ix, 1);
      this.syncFieldsMappings(null, header);
    },
    fieldsEdited(cell) {
      if (cell.getField() === "header") {
        this.syncFieldsMappings(cell.getValue(), cell.getOldValue());
      }
    },
    addMapping() {
      this.collection.mappings.push({ ixRole: "na", ixField: "na" });
    },
    mappingMoved(data) {
      this.collection.mappings = data;
    },
    mappingDelete(ix) {
      this.collection.mappings.splice(ix, 1);
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
          this.collection.mappings.push({ header: newValue, ixRole: "na", ixField: "na" });
        }
      } else {
        this.collection.mappings = this.collection.mappings.filter(m => m.header !== oldValue);
      }
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
      if (this.posts.length > 0) {
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
.submit-button {
  margin-top: 32px;
}
.tabulator {
  min-width: 640px;
}
.posts {
  margin-top: 32px;
}
.btn {
  margin: 32px 0;
}
</style>
