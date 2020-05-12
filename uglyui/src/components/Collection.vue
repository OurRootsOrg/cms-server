<template>
  <div>
    <div>
      <strong>{{ collection.name }}</strong> (in {{ category.name }})
    </div>
  </div>
</template>

<script>
export default {
  props: {
    collection: Object,
    category: Object
  }
};
</script>

<style scoped></style>
