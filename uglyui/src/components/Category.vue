<template>
  <div>
    <div>
      <strong>{{ category.name }}</strong> [ {{ collCount }} collections ]
      <slot></slot>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    category: Object,
    collCount: Number
  }
};
</script>

<style scoped></style>
