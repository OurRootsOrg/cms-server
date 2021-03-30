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
        <span>In {{ result.collectionName }} </span>
      </v-col>
      <v-col cols="12" md="3">
        <div v-for="(event, $ix) in result.person.events" :key="$ix">
          <p class="ma-0 pa-0">
            <span class="text-first-caps" v-if="event.type">{{ event.type }}:</span>
            {{ event.date }} <span v-if="event.date && event.place">,</span> {{ event.place }}
          </p>
        </div>
      </v-col>
      <v-col cols="12" md="4">
        <div v-for="(relationship, $ix) in result.person.relationships" :key="$ix">
          <span class="text-first-caps" v-if="relationship.type">{{ relationship.type }}:</span>
          {{ relationship.name }}
        </div>
      </v-col>
      <v-col cols="1" class="d-flex justify-center">
        <div class="view-column">
          <v-btn icon x-small class="primary--text" :to="{ name: 'search-detail', params: { rid: result.id } }"
            ><v-icon title="View record details">mdi-file-document</v-icon></v-btn
          >
          <v-btn
            v-if="result.imagePath"
            icon
            x-small
            class="primary--text view-camera"
            :to="{ name: 'image', params: { societyId: result.societyId, pid: result.post, path: result.imagePath } }"
            ><v-icon title="View image">mdi-camera</v-icon></v-btn
          >
        </div>
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
.view-camera {
  margin-left: 4px;
}
.view-column {
  margin-left: 4px;
  min-width: 48px;
}
.resultRole {
  font-size: 75%;
  color: #666;
  text-transform: uppercase;
  padding-bottom: 8px;
}
</style>
