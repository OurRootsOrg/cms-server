<template>
  <v-col cols="12">
    <v-row no-gutters class="no-underline">
      <v-col cols="12" md="4" class="d-flex flex-column">
        <strong
          ><router-link :to="{ name: 'search-detail', params: { rid: result.id } }">{{
            result.person.name
          }}</router-link></strong
        >
        <span class="text-first-caps resultRole">{{ result.person.role }}</span>
        <span
          >In: {{ result.collectionName }}
          <v-btn icon x-small class="primary--text" :to="{ name: 'search-detail', params: { rid: result.id } }"
            ><v-icon>mdi-chevron-right</v-icon></v-btn
          ></span
        >
      </v-col>
      <v-col cols="12" md="3">
        <div v-for="(event, $ix) in result.person.events" :key="$ix">
          <p class="ma-0 pa-0">
            <span class="text-first-caps">{{ event.type }}</span
            >: {{ event.date }} <span v-if="event.place">,</span> {{ event.place }}
          </p>
        </div>
      </v-col>
      <v-col cols="12" md="4">
        <div v-for="(relationship, $ix) in result.person.relationships" :key="$ix">
          <span class="text-first-caps">{{ relationship.type }}: {{ relationship.name }}</span>
        </div>
      </v-col>
      <v-col cols="1" class="d-flex justify-center">
        <v-btn icon x-small class="primary--text" :to="{ name: 'search-detail', params: { rid: result.id } }"
          ><v-icon>mdi-file-document</v-icon></v-btn
        >
        <!--TODO - need "has image" logic if there is an image associated with the record
        <v-row no-gutters class="ml-5">
          <router-link :to="{ name: 'search-detail', params: { rid: result.id } }">View image</router-link>
        </v-row>-->
      </v-col>
    </v-row>
  </v-col>
</template>

<script>
export default {
  props: {
    result: Object
  }
};
</script>

<style scoped>
.resultRole {
  font-size: 75%;
  color: #666;
  text-transform: uppercase;
  padding-bottom: 8px;
}
</style>
