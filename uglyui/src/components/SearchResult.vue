<template>
  <div class="result">
    <div>
      Name:
      <router-link :to="{ name: 'search-detail', params: { rid: result.id } }">{{ result.person.name }}</router-link>
    </div>
    <div>Role: {{ result.person.role }}</div>
    <div>Collection: {{ result.collectionName }}</div>
    <div v-for="(event, $ix) in result.person.events" :key="$ix">
      <span>{{ event.type }}: {{ event.date }}, {{ event.place }}</span>
    </div>
    <div v-for="(relationship, $ix) in result.person.relationships" :key="$ix">
      <span>{{ relationship.type }}: {{ relationship.name }}</span>
    </div>
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
