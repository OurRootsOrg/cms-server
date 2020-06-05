<template>
  <div class="result">
    <div>Given: {{ result._source.given }}</div>
    <div>Surname: {{ result._source.surname }}</div>
    <div>Collection: {{ result._source.collection }}</div>
  </div>
</template>

<script>
export default {
  props: {
    result: Object
  }
};
</script>

<style scoped>
.result {
  margin: 16px;
}
</style>
