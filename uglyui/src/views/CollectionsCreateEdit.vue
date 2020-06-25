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
      Object.assign(this.collection, this.collections.collection);
    }
  },
  data() {
    return {
      collection: { fields: [], mappings: [] },
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
            console.log("select header", this.collection);
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
      category: { required }
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
      console.log("fieldsDelete", ix, header);
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
    createEditCollection() {
      let collection = this.collection;
      collection.fields = collection.fields.filter(f => f.header);
      if (collection.fields.length === 0) {
        return;
      }
      collection.mappings = collection.mappings.filter(f => f.header);
      if (collection.mappings.length === 0) {
        return;
      }
      collection.category = +collection.category; // convert to a number
      this.$v.$touch();
      if (!this.$v.$invalid) {
        NProgress.start();
        this.$store
          .dispatch(this.collection.id ? "collectionsUpdate" : "collectionsCreate", collection)
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
.tabulator {
  min-width: 640px;
}
</style>
