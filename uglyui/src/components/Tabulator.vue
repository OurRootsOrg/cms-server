<template>
  <div ref="table"></div>
</template>

<script>
import Tabulator from "tabulator-tables";
import "tabulator-tables/dist/css/tabulator.min.css";

export default {
  props: {
    data: {
      type: Array,
      required: true
    },
    columns: {
      type: Array,
      required: true
    },
    layout: {
      type: String
    },
    movableRows: {
      type: Boolean
    },
    movableColumns: {
      type: Boolean
    },
    resizableColumns: {
      type: Boolean
    },
    virtualDom: {
      type: Boolean
    }
  },
  data() {
    return {
      tabulator: null
    };
  },
  watch: {
    data: {
      handler(newData) {
        this.tabulator.replaceData(newData);
      },
      deep: true
    }
  },
  mounted() {
    let self = this;
    this.tabulator = new Tabulator(this.$refs.table, {
      data: this.data,
      columns: this.columns,
      headerSort: false,
      selectable: false,
      layout: this.layout || "fitData",
      movableRows: this.movableRows,
      movableColumns: this.movableColumns,
      resizableColumns: this.resizableColumns,
      virtualDom: this.virtualDom,
      rowMoved: function() {
        self.$emit("updated", self.tabulator.getData());
      }
    });
  }
};
</script>

<style scoped></style>
