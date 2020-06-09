<template>
  <div class="result">
    <div>Name: {{ result.person.name }}</div>
    <div>Role: {{ result.person.role }}</div>
    <div>Collection: {{ result.collectionName }}</div>
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
