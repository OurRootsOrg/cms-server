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
    this.tabulator = new Tabulator(this.$refs.table, {
      data: this.data,
      columns: this.columns
    });
  }
};
</script>

<style scoped></style>
