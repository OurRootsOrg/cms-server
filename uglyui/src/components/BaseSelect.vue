<template>
  <div>
    <label v-if="label">{{ label }}</label>
    <select class="select" :value="value" @change="updateValue" v-bind="$attrs">
      <option v-for="option in options" :value="option.id" :key="option.id" :selected="option.id === value">{{
        option.name
      }}</option>
    </select>
  </div>
</template>
<script>
import { formFieldMixin } from "../mixins/formFieldMixin";
export default {
  mixins: [formFieldMixin],
  props: {
    options: {
      type: Array,
      required: true
    }
  }
};
</script>

<style scoped>
.select {
  margin-left: 8px;
}
</style>
