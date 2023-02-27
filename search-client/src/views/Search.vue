<template>
  <v-row no-gutters class="pa-0 ma-0 search">
    <!--search results-->
    <v-col cols="12" md="8" order-md="last" v-if="searchPerformed">
      <v-row>
        <v-col cols="12" align="center">
          <h1>Search results</h1>
        </v-col>
      </v-row>
      <v-row v-if="searchPerformed && search.searchTotal === 0">
        <p>No results found</p>
      </v-row>
      <v-row no-gutters v-if="searchPerformed && search.searchTotal > 0" class="pl-5">
        <!--search results-->
        <v-col cols="12">
          <v-row v-if="searchPerformed && search.searchTotal > 0" no-gutters class="pl-3 pb-3">
            Showing results {{ formattedFrom + 1 }} - {{ formattedFrom + formattedLength }} of
            {{ formattedSearchTotal }}
          </v-row>
          <v-card>
            <v-row no-gutters class="no-underline pl-3 resultsHeader">
              <v-col cols="12" md="4">
                Name
              </v-col>
              <v-col cols="12" md="3">
                {{ eventsLabel }}
              </v-col>
              <v-col cols="12" md="4">
                {{ relationshipsLabel }}
              </v-col>
              <v-col cols="12" md="1">
                View
              </v-col>
            </v-row>
            <v-row v-for="(result, $ix) in search.searchList" :key="$ix" no-gutters class="result">
              <SearchResult :result="result" />
            </v-row>
          </v-card>
          <v-row v-if="searchPerformed && search.searchTotal > 0" no-gutters class="mt-3">
            <v-pagination v-model="page" :length="numPages" :total-visible="7" @input="pageChanged()"></v-pagination>
          </v-row>
        </v-col>
      </v-row>
    </v-col>
    <!--search facets and form-->
    <v-col cols="12" :md="searchPerformed ? 4 : 12" order-md="first" class="mt-7">
      <!--search result facets-->
      <v-row no-gutters class="ma-0 pa-0">
        <v-col cols="12" :md="searchPerformed ? 12 : 3" class="mt-4">
          <v-row no-gutters class="mt-0 mb-1">
            <h4>Categories</h4>
          </v-row>
          <v-row no-gutters class="ma-0 pa-0 no-underline">
            <v-col cols="12" class="pa-0 ma-0">
              <v-row v-if="query.category" no-gutters>
                <v-col cols="12">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('category', null) }"
                    x-small
                    icon
                    class="grey--text pr-2"
                    v-if="!defaultCategory"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                  <router-link :to="{ name: 'search', query: getQuery('category', null) }">{{
                    query.category
                  }}</router-link>
                </v-col>
              </v-row>
              <v-row v-if="query.collection" no-gutters>
                <v-col cols="11" class="offset-md-1">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('collection', null) }"
                    x-small
                    icon
                    class="grey--text pr-2"
                    v-if="!defaultCollection"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                  <router-link :to="{ name: 'search', query: getQuery('collection', null) }">{{
                    query.collection
                  }}</router-link>
                </v-col>
              </v-row>
              <v-row v-if="categoryFacet" no-gutters>
                <v-col v-for="(bucket, $ix) in categoryFacet.buckets" :key="$ix" cols="12">
                  <v-row no-gutters>
                    <v-col
                      :cols="!query.category ? 12 : 11"
                      class="d-flex flex-row"
                      :class="!query.category ? '' : 'offset-md-1'"
                    >
                      <v-btn
                        :to="{ name: 'search', query: getQuery(categoryFacet.key, bucket.label) }"
                        x-small
                        icon
                        class="grey--text pr-2"
                      >
                        <v-icon>mdi-chevron-right</v-icon>
                      </v-btn>
                      <router-link :to="{ name: 'search', query: getQuery(categoryFacet.key, bucket.label) }">{{
                        bucket.label
                      }}</router-link>
                      <v-spacer></v-spacer>
                      {{ bucket.count }}
                    </v-col>
                  </v-row>
                </v-col>
              </v-row>
            </v-col>
          </v-row>
          <v-row no-gutters class="mt-5 mb-1">
            <h4>Collection Location</h4>
          </v-row>
          <v-row no-gutters class="no-underline">
            <v-col cols="12" class="pa-0 ma-0">
              <v-row v-if="query.collectionPlace1" no-gutters>
                <v-col cols="12">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('collectionPlace1', null) }"
                    x-small
                    icon
                    class="grey--text pr-2"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                  <span>{{ query.collectionPlace1 }}</span>
                </v-col>
              </v-row>
              <v-row v-if="query.collectionPlace2" no-gutters>
                <v-col cols="11" class="offset-md-1">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('collectionPlace2', null) }"
                    x-small
                    icon
                    class="grey--text pr-2"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                  <span>{{ query.collectionPlace2 }}</span>
                </v-col>
              </v-row>
              <v-row v-if="query.collectionPlace3" no-gutters>
                <v-col cols="10" class="offset-md-2">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('collectionPlace3', null) }"
                    x-small
                    icon
                    class="grey--text pr-2"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                  <span>{{ query.collectionPlace3 }}</span>
                </v-col>
              </v-row>
              <v-row v-if="placeFacet" no-gutters>
                <v-col v-for="(bucket, $ix) in placeFacet.buckets" :key="$ix" cols="12">
                  <v-row no-gutters>
                    <v-col
                      :cols="!query.collectionPlace1 ? 12 : !query.collectionPlace2 ? 11 : 10"
                      :class="!query.collectionPlace1 ? '' : !query.collectionPlace2 ? 'offset-md-1' : 'offset-md-2'"
                      class="d-flex flex-row"
                    >
                      <v-btn
                        :to="{ name: 'search', query: getQuery(placeFacet.key, bucket.label) }"
                        x-small
                        icon
                        class="grey--text pr-2"
                      >
                        <v-icon>mdi-chevron-right</v-icon>
                      </v-btn>
                      <router-link :to="{ name: 'search', query: getQuery(placeFacet.key, bucket.label) }">{{
                        bucket.label
                      }}</router-link>
                      <v-spacer></v-spacer>
                      {{ bucket.count }}
                    </v-col>
                  </v-row>
                </v-col>
              </v-row>
            </v-col>
          </v-row>
        </v-col>
        <!--search form-->
        <v-col cols="12" :md="searchPerformed ? 12 : 8" :offset-md="searchPerformed ? 0 : 1" class="search pa-0">
          <v-divider v-if="searchPerformed" class="my-5"></v-divider>
          <h1 v-if="!searchPerformed">Search</h1>
          <h3 v-if="searchPerformed" class="pa-0 ma-0">Refine your search</h3>
          <v-form @submit.prevent="go">
            <!--Title-->
            <v-row no-gutters class="mt-4" v-if="hasField('title')">
              <v-col cols="3">
                <h4 class="mt-2">Title:</h4>
              </v-col>
              <v-col cols="9">
                <v-text-field
                  outlined
                  dense
                  v-model="query.title"
                  type="text"
                  placeholder="Title"
                  class="ma-0 mb-n2"
                ></v-text-field>
              </v-col>
            </v-row>
            <!--Author-->
            <v-row no-gutters class="mt-4" v-if="hasField('author')">
              <v-col cols="3">
                <h4 class="mt-2">Author:</h4>
              </v-col>
              <v-col cols="9">
                <v-text-field
                  outlined
                  dense
                  v-model="query.author"
                  type="text"
                  placeholder="Author"
                  class="ma-0 mb-n2"
                ></v-text-field>
              </v-col>
            </v-row>
            <!--name row-->
            <v-row no-gutters v-if="hasField('given') || hasField('surname')">
              <v-col cols="3" v-if="!hasField('surname')">
                <h4 class="mt-5">First name:</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 6"
                :class="searchPerformed ? 'mt-3' : 'pr-3 mt-3'"
                v-if="hasField('given')"
              >
                <h4 v-if="hasField('surname')">First &amp; Middle Name(s)</h4>
                <v-row no-gutters>
                  <v-text-field
                    outlined
                    dense
                    v-model="query.given"
                    type="text"
                    placeholder="First &amp; Middle Name(s)"
                  ></v-text-field>
                </v-row>
                <v-row no-gutters class="mt-n5" v-if="query.given">
                  <v-menu
                    offset-x
                    :close-on-content-click="false"
                    v-model="givenOptionsMenu"
                    :nudge-top="autocompleteOffset"
                  >
                    <template v-slot:activator="{ on, attrs }">
                      <v-btn
                        color="primary"
                        text
                        x-small
                        v-bind="attrs"
                        v-on="on"
                        class="pa-0 mt-n1"
                        @click="openNameFuzziness('given')"
                      >
                        <v-icon v-if="fuzziness.given.length === 1 && fuzziness.given[0] === 0" small class="mr-1"
                          >mdi-checkbox-blank-outline</v-icon
                        >
                        <v-icon v-else small class="mr-1">mdi-checkbox-marked</v-icon>
                        <span class="displayChosenOptions">{{ givenSpellingOptions }}</span>
                      </v-btn>
                    </template>
                    <div class="exactnessOptions">
                      <v-checkbox
                        v-for="(item, index) in givenFuzzinessLevels"
                        :key="index"
                        v-model="fuzziness.dlg"
                        :value="item.value"
                        :label="item.text"
                        @change="nameFuzzinessChecked(item.value)"
                        class="ma-0 pa-0"
                        dense
                      >
                      </v-checkbox>
                    </div>
                    <v-card-actions class="exactnessActions">
                      <v-btn text @click="givenOptionsMenu = false">Cancel</v-btn>
                      <v-spacer></v-spacer>
                      <v-btn class="primary" @click="nameFuzzinessChanged('given')">Apply</v-btn>
                    </v-card-actions>
                  </v-menu>
                </v-row>
              </v-col>
              <v-col cols="3" v-else>
                <h4 class="mt-5">Surname:</h4>
              </v-col>
              <v-col
                cols="12"
                :md="hasField('given') ? (searchPerformed ? 12 : 6) : 9"
                class="mt-3"
                v-if="hasField('surname')"
              >
                <h4 v-if="hasField('given')">Last Name</h4>
                <v-row no-gutters>
                  <v-text-field dense outlined v-model="query.surname" type="text" placeholder="Surname"></v-text-field>
                </v-row>
                <v-row no-gutters class="mt-n5 mb-1" v-if="query.surname">
                  <v-menu
                    offset-x
                    :close-on-content-click="false"
                    v-model="surnameOptionsMenu"
                    :nudge-top="autocompleteOffset"
                  >
                    <template v-slot:activator="{ on, attrs }">
                      <v-btn
                        color="primary"
                        text
                        x-small
                        v-bind="attrs"
                        v-on="on"
                        class="pa-0 mt-n1"
                        @click="openNameFuzziness('surname')"
                      >
                        <v-icon v-if="fuzziness.surname.length === 1 && fuzziness.surname[0] === 0" small class="mr-1"
                          >mdi-checkbox-blank-outline</v-icon
                        >
                        <v-icon v-else small class="mr-1">mdi-checkbox-marked</v-icon>
                        <span class="displayChosenOptions">{{ surnameSpellingOptions }}</span>
                      </v-btn>
                    </template>
                    <div class="exactnessOptions">
                      <v-checkbox
                        v-for="(item, index) in surnameFuzzinessLevels"
                        :key="index"
                        v-model="fuzziness.dlg"
                        :value="item.value"
                        :label="item.text"
                        @change="nameFuzzinessChecked(item.value)"
                        class="ma-0 pa-0"
                        dense
                      >
                      </v-checkbox>
                    </div>
                    <v-card-actions class="exactnessActions">
                      <v-btn text @click="surnameOptionsMenu = false">Cancel</v-btn>
                      <v-spacer></v-spacer>
                      <v-btn class="primary" @click="nameFuzzinessChanged('surname')">Apply</v-btn>
                    </v-card-actions>
                  </v-menu>
                </v-row>
              </v-col>
            </v-row>
            <!--any place and birth year -->
            <v-row no-gutters v-if="!searchPerformed && hasField('events')">
              <v-col cols="12" md="6" class="pr-3 mt-1">
                <h5>Place your ancestor might have lived</h5>
                <v-row no-gutters>
                  <v-autocomplete
                    outlined
                    dense
                    v-model="defaultPlace"
                    :loading="defaultPlaceLoading"
                    :items="defaultPlaceItems"
                    :search-input.sync="defaultPlaceSearch"
                    no-filter
                    auto-select-first
                    clearable
                    flat
                    hide-no-data
                    hide-details
                    solo
                    placeholder="Any place"
                    :menu-props="{ nudgeTop: autocompleteOffset }"
                    @change="defaultPlaceChanged()"
                  ></v-autocomplete>
                </v-row>
                <v-row no-gutters v-if="defaultPlace">
                  <v-col cols="12" class="exactCheck d-flex flex-row">
                    <v-checkbox
                      v-model="query.anyPlaceFuzziness"
                      :value="1"
                      class="mt-0 ml-n1 primary--text shrink smallCheckbox"
                    >
                    </v-checkbox>
                    <span class="mt-2 primary--text">Exact</span>
                  </v-col>
                </v-row>
              </v-col>
              <v-col cols="6">
                <h4>Birth year</h4>
                <v-row no-gutters>
                  <v-text-field
                    outlined
                    dense
                    v-model="query.birthDate"
                    type="text"
                    placeholder="Birth year"
                    @change="birthYearChanged()"
                  ></v-text-field>
                </v-row>
                <v-row no-gutters class="mt-n5" v-if="query.birthDate">
                  <v-menu
                    offset-x
                    :close-on-content-click="true"
                    v-model="birthOptionsMenu"
                    :nudge-top="autocompleteOffset"
                  >
                    <template v-slot:activator="{ on, attrs }">
                      <v-btn color="primary" text x-small v-bind="attrs" v-on="on" class="pa-0 mt-0">
                        <v-icon v-if="query.birthDateFuzziness === 0" small class="mr-1"
                          >mdi-checkbox-blank-outline</v-icon
                        >
                        <v-icon v-if="query.birthDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                        <span class="ml-1">{{
                          query.birthDateFuzziness === 0
                            ? "Date range"
                            : dateRanges.find(d => d.value === query.birthDateFuzziness).text
                        }}</span>
                      </v-btn>
                    </template>
                    <div class="exactnessOptions mt-2 pb-0">
                      <v-radio-group
                        v-for="(item, index) in dateRanges"
                        :key="index"
                        v-model="query.birthDateFuzziness"
                        class="ma-0 pa-0"
                        @change="birthOptionsMenu = false"
                      >
                        <v-radio :label="item.text" :value="item.value"></v-radio>
                      </v-radio-group>
                    </div>
                  </v-menu>
                </v-row>
              </v-col>
            </v-row>
            <!--Event buttons-->
            <v-row no-gutters :class="searchPerformed ? '' : 'mt-5'" v-if="hasField('events')">
              <v-col cols="12" :md="searchPerformed ? 12 : 3" :class="searchPerformed ? 'mt-3' : ''">
                <strong>Add event details:</strong>
              </v-col>
              <v-col cols="12" :md="searchPerformed ? 12 : 9">
                <v-btn
                  text
                  color="primary"
                  class="eventButton"
                  :disabled="showEvent.birth"
                  @click="showEvent.birth = true"
                  v-if="!showEvent.birth"
                  >Birth</v-btn
                >
                <v-btn
                  text
                  color="primary"
                  class="eventButton"
                  :disabled="showEvent.marriage"
                  @click="showEvent.marriage = true"
                  v-if="!showEvent.marriage"
                  >Marriage</v-btn
                >
                <v-btn
                  text
                  color="primary"
                  class="eventButton"
                  :disabled="showEvent.death"
                  @click="showEvent.death = true"
                  v-if="!showEvent.death"
                  >Death</v-btn
                >
                <v-btn
                  text
                  color="primary"
                  class="eventButton"
                  :disabled="showEvent.residence"
                  @click="showEvent.residence = true"
                  v-if="!showEvent.residence"
                  >Lived In</v-btn
                >
                <v-btn
                  text
                  color="primary"
                  class="eventButton"
                  :disabled="showEvent.any"
                  @click="showEvent.any = true"
                  v-if="!showEvent.any"
                  >Any Event</v-btn
                >
              </v-col>
            </v-row>
            <!--Birth-->
            <v-row no-gutters :class="searchPerformed ? 'ma-0 pa-0 d-flex' : 'my-3'" v-if="showEvent.birth">
              <v-col cols="2" class="order-1" :class="searchPerformed ? 'pt-2' : 'pl-5 pt-3'">
                <h4>Birth</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 3"
                :class="searchPerformed ? 'd-flex flex-row order-3' : 'order-2'"
              >
                <v-col :cols="searchPerformed ? 6 : 12" :class="searchPerformed ? '' : 'pr-3'" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.birthDate"
                    type="text"
                    placeholder="Birth date"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-col :class="!searchPerformed ? 'ma-0 pa-0 mt-1' : 'ma-0 pa-0 mt-2'" v-if="query.birthDate">
                  <v-menu
                    offset-x
                    :close-on-content-click="true"
                    v-model="birthOptionsMenu2"
                    :nudge-top="autocompleteOffset"
                  >
                    <template v-slot:activator="{ on, attrs }">
                      <v-btn
                        color="primary"
                        text
                        x-small
                        v-bind="attrs"
                        v-on="on"
                        :class="searchPerformed ? '' : 'pa-0 mt-n2'"
                      >
                        <v-icon v-if="query.birthDateFuzziness === 0" small class="mr-1"
                          >mdi-checkbox-blank-outline</v-icon
                        >
                        <v-icon v-if="query.birthDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                        <span class="ml-1" :class="searchPerformed ? 'mt-0' : ' mt-1'">{{
                          query.birthDateFuzziness === 0
                            ? "Date range"
                            : dateRanges.find(d => d.value === query.birthDateFuzziness).text
                        }}</span>
                      </v-btn>
                    </template>
                    <div class="exactnessOptions mt-2 pb-0">
                      <v-radio-group
                        v-for="(item, index) in dateRanges"
                        :key="index"
                        v-model="query.birthDateFuzziness"
                        class="ma-0 pa-0"
                        @change="birthOptionsMenu = false"
                      >
                        <v-radio :label="item.text" :value="item.value"></v-radio>
                      </v-radio-group>
                    </div>
                  </v-menu>
                </v-col>
              </v-col>
              <v-col :class="searchPerformed ? 'order-4 mt-2 d-flex flex-row' : 'order-3 mb-0'">
                <v-col :cols="searchPerformed ? 10 : 12" class="ma-0 pa-0">
                  <v-autocomplete
                    outlined
                    dense
                    v-model="query.birthPlace"
                    :loading="birthPlaceLoading"
                    :items="birthPlaceItems"
                    :search-input.sync="birthPlaceSearch"
                    no-filter
                    auto-select-first
                    clearable
                    flat
                    hide-no-data
                    hide-details
                    solo
                    placeholder="Birth place"
                    :menu-props="{ nudgeTop: autocompleteOffset }"
                  ></v-autocomplete>
                </v-col>
                <v-col
                  v-if="query.birthPlace"
                  :cols="searchPerformed ? 2 : 12"
                  class="exactCheck d-flex flex-row ml-0 pl-0"
                  :class="searchPerformed ? 'mt-n2' : 'mt-n3'"
                >
                  <v-checkbox
                    v-model="query.birthPlaceFuzziness"
                    :value="1"
                    class="shrink mt-0 smallCheckbox"
                    dense
                    primary
                    hide-details="true"
                  >
                  </v-checkbox
                  ><span class="mt-2 primary--text">Exact</span>
                </v-col>
              </v-col>
              <v-col
                :cols="searchPerformed ? 10 : 1"
                :class="searchPerformed ? 'order-2 pr-0 mr-0 text-right mb-n1' : 'ma-0 order-4'"
              >
                <v-btn text @click="clearEvent('birth')" class="grey--text" :class="searchPerformed ? 'mr-n5' : 'mt-0'"
                  ><v-icon class="pa-0 ma-0">mdi-close-circle-outline</v-icon></v-btn
                >
              </v-col>
            </v-row>
            <!--Marriage-->
            <v-row no-gutters :class="searchPerformed ? 'ma-0 pa-0 d-flex' : 'my-3'" v-if="showEvent.marriage">
              <v-col cols="2" class="order-1" :class="searchPerformed ? 'pt-2' : 'pl-5 pt-3'">
                <h4>Marriage</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 3"
                :class="searchPerformed ? 'd-flex flex-row order-3' : 'order-2'"
              >
                <v-col :cols="searchPerformed ? 6 : 12" :class="searchPerformed ? '' : 'pr-3'" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.marriageDate"
                    type="text"
                    placeholder="Marriage date"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-col :class="!searchPerformed ? 'ma-0 pa-0 mt-1' : 'ma-0 pa-0 mt-2'" v-if="query.marriageDate">
                  <v-menu
                    offset-x
                    :close-on-content-click="true"
                    v-model="marriageOptionsMenu2"
                    :nudge-top="autocompleteOffset"
                  >
                    <template v-slot:activator="{ on, attrs }">
                      <v-btn
                        color="primary"
                        text
                        x-small
                        v-bind="attrs"
                        v-on="on"
                        :class="searchPerformed ? '' : 'pa-0 mt-n2'"
                      >
                        <v-icon v-if="query.marriageDateFuzziness === 0" small class="mr-1"
                          >mdi-checkbox-blank-outline</v-icon
                        >
                        <v-icon v-if="query.marriageDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                        <span class="ml-1" :class="searchPerformed ? 'mt-0' : ' mt-1'">{{
                          query.marriageDateFuzziness === 0
                            ? "Date range"
                            : dateRanges.find(d => d.value === query.marriageDateFuzziness).text
                        }}</span>
                      </v-btn>
                    </template>
                    <div class="exactnessOptions mt-2 pb-0">
                      <v-radio-group
                        v-for="(item, index) in dateRanges"
                        :key="index"
                        v-model="query.marriageDateFuzziness"
                        class="ma-0 pa-0"
                        @change="marriageOptionsMenu = false"
                      >
                        <v-radio :label="item.text" :value="item.value"></v-radio>
                      </v-radio-group>
                    </div>
                  </v-menu>
                </v-col>
              </v-col>
              <v-col :class="searchPerformed ? 'order-4 mt-2 d-flex flex-row' : 'order-3 mb-0'">
                <v-col :cols="searchPerformed ? 10 : 12" class="ma-0 pa-0">
                  <v-autocomplete
                    outlined
                    dense
                    v-model="query.marriagePlace"
                    :loading="marriagePlaceLoading"
                    :items="marriagePlaceItems"
                    :search-input.sync="marriagePlaceSearch"
                    no-filter
                    auto-select-first
                    clearable
                    flat
                    hide-no-data
                    hide-details
                    solo
                    placeholder="Marriage place"
                    :menu-props="{ nudgeTop: autocompleteOffset }"
                  ></v-autocomplete>
                </v-col>
                <v-col
                  v-if="query.marriagePlace"
                  :cols="searchPerformed ? 2 : 12"
                  class="exactCheck d-flex flex-row ml-0 pl-0"
                  :class="searchPerformed ? 'mt-n2' : 'mt-n3'"
                >
                  <v-checkbox
                    v-model="query.marriagePlaceFuzziness"
                    :value="1"
                    class="shrink mt-0 smallCheckbox"
                    dense
                    primary
                    hide-details="true"
                  >
                  </v-checkbox
                  ><span class="mt-2 primary--text">Exact</span>
                </v-col>
              </v-col>
              <v-col
                :cols="searchPerformed ? 10 : 1"
                :class="searchPerformed ? 'order-2 pr-0 mr-0 text-right mb-n1' : 'ma-0 order-4'"
              >
                <v-btn
                  text
                  @click="clearEvent('marriage')"
                  class="grey--text"
                  :class="searchPerformed ? 'mr-n5' : 'mt-0'"
                  ><v-icon class="pa-0 ma-0">mdi-close-circle-outline</v-icon></v-btn
                >
              </v-col>
            </v-row>
            <!--Death-->
            <v-row no-gutters :class="searchPerformed ? 'ma-0 pa-0 d-flex' : 'my-3'" v-if="showEvent.death">
              <v-col cols="2" class="order-1" :class="searchPerformed ? 'pt-2' : 'pl-5 pt-3'">
                <h4>Death</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 3"
                :class="searchPerformed ? 'd-flex flex-row order-3' : 'order-2'"
              >
                <v-col :cols="searchPerformed ? 6 : 12" :class="searchPerformed ? '' : 'pr-3'" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.deathDate"
                    type="text"
                    placeholder="Death date"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-col :class="!searchPerformed ? 'ma-0 pa-0 mt-1' : 'ma-0 pa-0 mt-2'" v-if="query.deathDate">
                  <v-menu
                    offset-x
                    :close-on-content-click="true"
                    v-model="deathOptionsMenu2"
                    :nudge-top="autocompleteOffset"
                  >
                    <template v-slot:activator="{ on, attrs }">
                      <v-btn
                        color="primary"
                        text
                        x-small
                        v-bind="attrs"
                        v-on="on"
                        :class="searchPerformed ? '' : 'pa-0 mt-n2'"
                      >
                        <v-icon v-if="query.deathDateFuzziness === 0" small class="mr-1"
                          >mdi-checkbox-blank-outline</v-icon
                        >
                        <v-icon v-if="query.deathDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                        <span class="ml-1" :class="searchPerformed ? 'mt-0' : ' mt-1'">{{
                          query.deathDateFuzziness === 0
                            ? "Date range"
                            : dateRanges.find(d => d.value === query.deathDateFuzziness).text
                        }}</span>
                      </v-btn>
                    </template>
                    <div class="exactnessOptions mt-2 pb-0">
                      <v-radio-group
                        v-for="(item, index) in dateRanges"
                        :key="index"
                        v-model="query.deathDateFuzziness"
                        class="ma-0 pa-0"
                        @change="deathOptionsMenu = false"
                      >
                        <v-radio :label="item.text" :value="item.value"></v-radio>
                      </v-radio-group>
                    </div>
                  </v-menu>
                </v-col>
              </v-col>
              <v-col :class="searchPerformed ? 'order-4 mt-2 d-flex flex-row' : 'order-3 mb-0'">
                <v-col :cols="searchPerformed ? 10 : 12" class="ma-0 pa-0">
                  <v-autocomplete
                    outlined
                    dense
                    v-model="query.deathPlace"
                    :loading="deathPlaceLoading"
                    :items="deathPlaceItems"
                    :search-input.sync="deathPlaceSearch"
                    no-filter
                    auto-select-first
                    clearable
                    flat
                    hide-no-data
                    hide-details
                    solo
                    placeholder="Death place"
                    :menu-props="{ nudgeTop: autocompleteOffset }"
                  ></v-autocomplete>
                </v-col>
                <v-col
                  v-if="query.deathPlace"
                  :cols="searchPerformed ? 2 : 12"
                  class="exactCheck d-flex flex-row ml-0 pl-0"
                  :class="searchPerformed ? 'mt-n2' : 'mt-n3'"
                >
                  <v-checkbox
                    v-model="query.deathPlaceFuzziness"
                    :value="1"
                    class="shrink mt-0 smallCheckbox"
                    dense
                    primary
                    hide-details="true"
                  >
                  </v-checkbox
                  ><span class="mt-2 primary--text">Exact</span>
                </v-col>
              </v-col>
              <v-col
                :cols="searchPerformed ? 10 : 1"
                :class="searchPerformed ? 'order-2 pr-0 mr-0 text-right mb-n1' : 'ma-0 order-4'"
              >
                <v-btn text @click="clearEvent('death')" class="grey--text" :class="searchPerformed ? 'mr-n5' : 'mt-0'"
                  ><v-icon class="pa-0 ma-0">mdi-close-circle-outline</v-icon></v-btn
                >
              </v-col>
            </v-row>
            <!--Residence-->
            <v-row no-gutters :class="searchPerformed ? 'ma-0 pa-0 d-flex' : 'my-3'" v-if="showEvent.residence">
              <v-col cols="2" class="order-1" :class="searchPerformed ? 'pt-2' : 'pl-5 pt-3'">
                <h4>Residence</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 3"
                :class="searchPerformed ? 'd-flex flex-row order-3' : 'order-2'"
              >
                <v-col :cols="searchPerformed ? 6 : 12" :class="searchPerformed ? '' : 'pr-3'" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.residenceDate"
                    type="text"
                    placeholder="Residence date"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-col :class="!searchPerformed ? 'ma-0 pa-0 mt-1' : 'ma-0 pa-0 mt-2'" v-if="query.residenceDate">
                  <v-menu
                    offset-x
                    :close-on-content-click="true"
                    v-model="residenceOptionsMenu2"
                    :nudge-top="autocompleteOffset"
                  >
                    <template v-slot:activator="{ on, attrs }">
                      <v-btn
                        color="primary"
                        text
                        x-small
                        v-bind="attrs"
                        v-on="on"
                        :class="searchPerformed ? '' : 'pa-0 mt-n2'"
                      >
                        <v-icon v-if="query.residenceDateFuzziness === 0" small class="mr-1"
                          >mdi-checkbox-blank-outline</v-icon
                        >
                        <v-icon v-if="query.residenceDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                        <span class="ml-1" :class="searchPerformed ? 'mt-0' : ' mt-1'">{{
                          query.residenceDateFuzziness === 0
                            ? "Date range"
                            : dateRanges.find(d => d.value === query.residenceDateFuzziness).text
                        }}</span>
                      </v-btn>
                    </template>
                    <div class="exactnessOptions mt-2 pb-0">
                      <v-radio-group
                        v-for="(item, index) in dateRanges"
                        :key="index"
                        v-model="query.residenceDateFuzziness"
                        class="ma-0 pa-0"
                        @change="residenceOptionsMenu = false"
                      >
                        <v-radio :label="item.text" :value="item.value"></v-radio>
                      </v-radio-group>
                    </div>
                  </v-menu>
                </v-col>
              </v-col>
              <v-col :class="searchPerformed ? 'order-4 mt-2 d-flex flex-row' : 'order-3 mb-0'">
                <v-col :cols="searchPerformed ? 10 : 12" class="ma-0 pa-0">
                  <v-autocomplete
                    outlined
                    dense
                    v-model="query.residencePlace"
                    :loading="residencePlaceLoading"
                    :items="residencePlaceItems"
                    :search-input.sync="residencePlaceSearch"
                    no-filter
                    auto-select-first
                    clearable
                    flat
                    hide-no-data
                    hide-details
                    solo
                    placeholder="Residence place"
                    :menu-props="{ nudgeTop: autocompleteOffset }"
                  ></v-autocomplete>
                </v-col>
                <v-col
                  v-if="query.residencePlace"
                  :cols="searchPerformed ? 2 : 12"
                  class="exactCheck d-flex flex-row ml-0 pl-0"
                  :class="searchPerformed ? 'mt-n2' : 'mt-n3'"
                >
                  <v-checkbox
                    v-model="query.residencePlaceFuzziness"
                    :value="1"
                    class="shrink mt-0 smallCheckbox"
                    dense
                    primary
                    hide-details="true"
                  >
                  </v-checkbox
                  ><span class="mt-2 primary--text">Exact</span>
                </v-col>
              </v-col>
              <v-col
                :cols="searchPerformed ? 10 : 1"
                :class="searchPerformed ? 'order-2 pr-0 mr-0 text-right mb-n1' : 'ma-0 order-4'"
              >
                <v-btn
                  text
                  @click="clearEvent('residence')"
                  class="grey--text"
                  :class="searchPerformed ? 'mr-n5' : 'mt-0'"
                  ><v-icon class="pa-0 ma-0">mdi-close-circle-outline</v-icon></v-btn
                >
              </v-col>
            </v-row>
            <!--Any-->
            <v-row
              no-gutters
              :class="searchPerformed ? 'ma-0 pa-0 d-flex' : 'my-3'"
              v-if="showEvent.any && hasField('events')"
            >
              <v-col :cols="searchPerformed ? 3 : 2" class="order-1" :class="searchPerformed ? 'pt-2' : 'pl-5 pt-3'">
                <h4>Any event</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 3"
                :class="searchPerformed ? 'd-flex flex-row order-3' : 'order-2'"
              >
                <v-col :cols="searchPerformed ? 6 : 12" :class="searchPerformed ? '' : 'pr-3'" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.anyDate"
                    type="text"
                    placeholder="Any date"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-col :class="!searchPerformed ? 'ma-0 pa-0 mt-1' : 'ma-0 pa-0 mt-2'" v-if="query.anyDate">
                  <v-menu
                    offset-x
                    :close-on-content-click="true"
                    v-model="anyOptionsMenu2"
                    :nudge-top="autocompleteOffset"
                  >
                    <template v-slot:activator="{ on, attrs }">
                      <v-btn
                        color="primary"
                        text
                        x-small
                        v-bind="attrs"
                        v-on="on"
                        :class="searchPerformed ? '' : 'pa-0 mt-n2'"
                      >
                        <v-icon v-if="query.anyDateFuzziness === 0" small class="mr-1"
                          >mdi-checkbox-blank-outline</v-icon
                        >
                        <v-icon v-if="query.anyDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                        <span class="ml-1" :class="searchPerformed ? 'mt-0' : ' mt-1'">{{
                          query.anyDateFuzziness === 0
                            ? "Date range"
                            : dateRanges.find(d => d.value === query.anyDateFuzziness).text
                        }}</span>
                      </v-btn>
                    </template>
                    <div class="exactnessOptions mt-2 pb-0">
                      <v-radio-group
                        v-for="(item, index) in dateRanges"
                        :key="index"
                        v-model="query.anyDateFuzziness"
                        class="ma-0 pa-0"
                        @change="anyOptionsMenu = false"
                      >
                        <v-radio :label="item.text" :value="item.value"></v-radio>
                      </v-radio-group>
                    </div>
                  </v-menu>
                </v-col>
              </v-col>
              <v-col :class="searchPerformed ? 'order-4 mt-2 d-flex flex-row' : 'order-3 mb-0'">
                <v-col :cols="searchPerformed ? 10 : 12" class="ma-0 pa-0">
                  <v-autocomplete
                    outlined
                    dense
                    v-model="query.anyPlace"
                    :loading="anyPlaceLoading"
                    :items="anyPlaceItems"
                    :search-input.sync="anyPlaceSearch"
                    no-filter
                    auto-select-first
                    clearable
                    flat
                    hide-no-data
                    hide-details
                    solo
                    placeholder="Any place"
                    :menu-props="{ nudgeTop: autocompleteOffset }"
                    @change="anyPlaceChanged()"
                  ></v-autocomplete>
                </v-col>
                <v-col
                  v-if="query.anyPlace"
                  :cols="searchPerformed ? 2 : 12"
                  class="exactCheck d-flex flex-row ml-0 pl-0"
                  :class="searchPerformed ? 'mt-n2' : 'mt-n3'"
                >
                  <v-checkbox
                    v-model="query.anyPlaceFuzziness"
                    :value="1"
                    class="shrink mt-0 smallCheckbox"
                    dense
                    primary
                    hide-details="true"
                  >
                  </v-checkbox
                  ><span class="mt-2 primary--text">Exact</span>
                </v-col>
              </v-col>
              <v-col
                :cols="searchPerformed ? 9 : 1"
                :class="searchPerformed ? 'order-2 pr-0 mr-0 text-right mb-n1' : 'ma-0 order-4'"
              >
                <v-btn text @click="clearEvent('any')" class="grey--text" :class="searchPerformed ? 'mr-n5' : 'mt-0'"
                  ><v-icon class="pa-0 ma-0">mdi-close-circle-outline</v-icon></v-btn
                >
              </v-col>
            </v-row>
            <!--Relative buttons-->
            <v-row no-gutters class="mt-5" v-if="hasField('relationships')">
              <v-col cols="12" :md="searchPerformed ? 12 : 3">
                <strong>Add family member:</strong>
              </v-col>
              <v-col cols="12" :md="searchPerformed ? 12 : 9">
                <v-btn
                  text
                  color="primary"
                  class="eventButton"
                  :disabled="showRelative.father"
                  @click="showRelative.father = true"
                  v-if="!showRelative.father"
                  >Father</v-btn
                >
                <v-btn
                  text
                  color="primary"
                  class="eventButton"
                  :disabled="showRelative.mother"
                  @click="showRelative.mother = true"
                  v-if="!showRelative.mother"
                  >Mother</v-btn
                >
                <v-btn
                  text
                  color="primary"
                  class="eventButton"
                  :disabled="showRelative.spouse"
                  @click="showRelative.spouse = true"
                  v-if="!showRelative.spouse"
                  >Spouse</v-btn
                >
                <v-btn
                  text
                  color="primary"
                  class="eventButton"
                  :disabled="showRelative.other"
                  @click="showRelative.other = true"
                  v-if="!showRelative.other"
                  >Other</v-btn
                >
              </v-col>
            </v-row>
            <!--Father-->
            <v-row no-gutters :class="searchPerformed ? 'ma-0 pa-0 d-flex' : 'my-3'" v-if="showRelative.father">
              <v-col cols="2" class="order-1" :class="searchPerformed ? 'pt-2' : 'pl-5 pt-3'">
                <h4 class="mb-1">Father</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 3"
                :class="searchPerformed ? 'd-flex flex-row order-3' : 'order-2'"
              >
                <v-col :cols="searchPerformed ? 9 : 12" :class="searchPerformed ? 'mb-2' : 'pr-3'" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.fatherGiven"
                    type="text"
                    placeholder="Father's given name"
                    class="ma-0"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-row no-gutters v-if="query.fatherGiven" :class="searchPerformed ? 'mt-1' : 'mt-n1 ma-0 pa-0'">
                  <v-col cols="12" class="exactCheck d-flex flex-row">
                    <v-checkbox
                      v-model="fuzziness.fatherGiven"
                      :value="1"
                      class="shrink mt-0 smallCheckbox"
                      dense
                      primary
                      hide-details="true"
                      ripple="false"
                    >
                    </v-checkbox
                    ><span class="mt-2 primary--text">Exact</span>
                  </v-col>
                </v-row>
              </v-col>
              <v-col :class="searchPerformed ? 'order-4 d-flex flex-row' : 'order-3'" class="ma-0 pa-0">
                <v-col :cols="searchPerformed ? 9 : 12" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.fatherSurname"
                    type="text"
                    placeholder="Father's surname"
                    class="ma-0"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-col
                  v-if="query.fatherSurname"
                  :cols="searchPerformed ? 3 : 12"
                  class="exactCheck d-flex flex-row mt-1"
                  :class="searchPerformed ? 'mt-n2 ml-n3' : 'mt-n4 ml-n3'"
                >
                  <v-checkbox
                    v-model="fuzziness.fatherSurname"
                    :value="1"
                    class="shrink mt-0 smallCheckbox"
                    dense
                    primary
                    hide-details="true"
                  >
                  </v-checkbox
                  ><span v-show="query.fatherSurname" class="mt-2 primary--text">Exact</span>
                </v-col>
              </v-col>
              <v-col
                :cols="searchPerformed ? 10 : 1"
                :class="searchPerformed ? 'order-2 pr-0 mr-0 text-right mb-n1' : 'ma-0 order-4'"
              >
                <v-btn
                  text
                  @click="clearRelative('father')"
                  class="grey--text mt-0"
                  :class="searchPerformed ? 'mr-n5' : 'mt-0'"
                  ><v-icon>mdi-close-circle-outline</v-icon></v-btn
                >
              </v-col>
            </v-row>
            <!--Mother-->
            <v-row no-gutters :class="searchPerformed ? 'ma-0 pa-0 d-flex' : 'my-3'" v-if="showRelative.mother">
              <v-col cols="2" class="order-1" :class="searchPerformed ? 'pt-2' : 'pl-5 pt-3'">
                <h4 class="mb-1">Mother</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 3"
                :class="searchPerformed ? 'd-flex flex-row order-3' : 'order-2'"
              >
                <v-col :cols="searchPerformed ? 9 : 12" :class="searchPerformed ? 'mb-2' : 'pr-3'" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.motherGiven"
                    type="text"
                    placeholder="Mother's given name"
                    class="ma-0"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-row no-gutters v-if="query.motherGiven" :class="searchPerformed ? 'mt-1' : 'mt-n1 ma-0 pa-0'">
                  <v-col cols="12" class="exactCheck d-flex flex-row">
                    <v-checkbox
                      v-model="fuzziness.motherGiven"
                      :value="1"
                      class="shrink mt-0 smallCheckbox"
                      dense
                      primary
                      hide-details="true"
                    >
                    </v-checkbox
                    ><span class="mt-2 primary--text">Exact</span>
                  </v-col>
                </v-row>
              </v-col>
              <v-col :class="searchPerformed ? 'order-4 d-flex flex-row' : 'order-3'" class="ma-0 pa-0">
                <v-col :cols="searchPerformed ? 9 : 12" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.motherSurname"
                    type="text"
                    placeholder="Mother's surname"
                    class="ma-0"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-col
                  v-if="query.motherSurname"
                  :cols="searchPerformed ? 3 : 12"
                  class="exactCheck d-flex flex-row mt-1"
                  :class="searchPerformed ? 'mt-n2 ml-n3' : 'mt-n4 ml-n3'"
                >
                  <v-checkbox
                    v-model="fuzziness.motherSurname"
                    :value="1"
                    class="shrink mt-0 smallCheckbox"
                    dense
                    primary
                    hide-details="true"
                  >
                  </v-checkbox
                  ><span class="mt-2 primary--text">Exact</span>
                </v-col>
              </v-col>
              <v-col
                :cols="searchPerformed ? 10 : 1"
                :class="searchPerformed ? 'order-2 pr-0 mr-0 text-right mb-n1' : 'ma-0 order-4'"
              >
                <v-btn
                  text
                  @click="clearRelative('mother')"
                  class="grey--text mt-0"
                  :class="searchPerformed ? 'mr-n5' : 'mt-0'"
                  ><v-icon>mdi-close-circle-outline</v-icon></v-btn
                >
              </v-col>
            </v-row>
            <!--Spouse-->
            <v-row no-gutters :class="searchPerformed ? 'ma-0 pa-0 d-flex' : 'my-3'" v-if="showRelative.spouse">
              <v-col cols="2" class="order-1" :class="searchPerformed ? 'pt-2' : 'pl-5 pt-3'">
                <h4 class="mb-1">Spouse</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 3"
                :class="searchPerformed ? 'd-flex flex-row order-3' : 'order-2'"
              >
                <v-col :cols="searchPerformed ? 9 : 12" :class="searchPerformed ? 'mb-2' : 'pr-3'" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.spouseGiven"
                    type="text"
                    placeholder="Spouse's given name"
                    class="ma-0"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-row no-gutters v-if="query.spouseGiven" :class="searchPerformed ? 'mt-1' : 'mt-n1 ma-0 pa-0'">
                  <v-col cols="12" class="exactCheck d-flex flex-row">
                    <v-checkbox
                      v-model="fuzziness.spouseGiven"
                      :value="1"
                      class="shrink mt-0 smallCheckbox"
                      dense
                      primary
                      hide-details="true"
                    >
                    </v-checkbox
                    ><span class="mt-2 primary--text">Exact</span>
                  </v-col>
                </v-row>
              </v-col>
              <v-col :class="searchPerformed ? 'order-4 d-flex flex-row' : 'order-3'" class="ma-0 pa-0">
                <v-col :cols="searchPerformed ? 9 : 12" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.spouseSurname"
                    type="text"
                    placeholder="Spouse's surname"
                    class="ma-0"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-col
                  v-if="query.spouseSurname"
                  :cols="searchPerformed ? 3 : 12"
                  class="exactCheck d-flex flex-row mt-1"
                  :class="searchPerformed ? 'mt-n2 ml-n3' : 'mt-n4 ml-n3'"
                >
                  <v-checkbox
                    v-model="fuzziness.spouseSurname"
                    :value="1"
                    class="shrink mt-0 smallCheckbox"
                    dense
                    primary
                    hide-details="true"
                  >
                  </v-checkbox
                  ><span class="mt-2 primary--text">Exact</span>
                </v-col>
              </v-col>
              <v-col
                :cols="searchPerformed ? 10 : 1"
                :class="searchPerformed ? 'order-2 pr-0 mr-0 text-right mb-n1' : 'ma-0 order-4'"
              >
                <v-btn
                  text
                  @click="clearRelative('spouse')"
                  class="grey--text mt-0"
                  :class="searchPerformed ? 'mr-n5' : 'mt-0'"
                  ><v-icon>mdi-close-circle-outline</v-icon></v-btn
                >
              </v-col>
            </v-row>
            <!--Other-->
            <v-row no-gutters :class="searchPerformed ? 'ma-0 pa-0 d-flex' : 'my-3'" v-if="showRelative.other">
              <v-col :cols="searchPerformed ? 4 : 2" class="order-1" :class="searchPerformed ? 'pt-2' : 'pl-5 pt-3'">
                <h4 class="mb-1">Other person</h4>
              </v-col>
              <v-col
                cols="12"
                :md="searchPerformed ? 12 : 3"
                :class="searchPerformed ? 'd-flex flex-row order-3' : 'order-2'"
              >
                <v-col :cols="searchPerformed ? 9 : 12" :class="searchPerformed ? 'mb-2' : 'pr-3'" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.otherGiven"
                    type="text"
                    placeholder="Other's given name"
                    class="ma-0"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-row no-gutters v-if="query.otherGiven" :class="searchPerformed ? 'mt-1' : 'mt-n1 ma-0 pa-0'">
                  <v-col cols="12" class="exactCheck d-flex flex-row">
                    <v-checkbox
                      v-model="fuzziness.otherGiven"
                      :value="1"
                      class="shrink mt-0 smallCheckbox"
                      dense
                      primary
                      hide-details="true"
                    >
                    </v-checkbox
                    ><span class="mt-2 primary--text">Exact</span>
                  </v-col>
                </v-row>
              </v-col>
              <v-col :class="searchPerformed ? 'order-4 d-flex flex-row' : 'order-3'" class="ma-0 pa-0">
                <v-col :cols="searchPerformed ? 9 : 12" class="ma-0 pa-0">
                  <v-text-field
                    outlined
                    dense
                    v-model="query.otherSurname"
                    type="text"
                    placeholder="Other's surname"
                    class="ma-0"
                    hide-details="true"
                  ></v-text-field>
                </v-col>
                <v-col
                  v-if="query.otherSurname"
                  :cols="searchPerformed ? 3 : 12"
                  class="exactCheck d-flex flex-row mt-1"
                  :class="searchPerformed ? 'mt-n2 ml-n3' : 'mt-n4 ml-n3'"
                >
                  <v-checkbox
                    v-model="fuzziness.otherSurname"
                    :value="1"
                    class="shrink mt-0 smallCheckbox"
                    dense
                    primary
                    hide-details="true"
                  >
                  </v-checkbox
                  ><span class="mt-2 primary--text">Exact</span>
                </v-col>
              </v-col>
              <v-col
                :cols="searchPerformed ? 8 : 1"
                :class="searchPerformed ? 'order-2 pr-0 mr-0 text-right mb-n1' : 'ma-0 order-4'"
              >
                <v-btn
                  text
                  @click="clearRelative('other')"
                  class="grey--text mt-0"
                  :class="searchPerformed ? 'mr-n5' : 'mt-0'"
                  ><v-icon>mdi-close-circle-outline</v-icon></v-btn
                >
              </v-col>
            </v-row>
            <!--Place-->
            <v-row no-gutters class="mt-2 mb-6" v-if="hasField('place')">
              <v-col cols="3">
                <h4 class="mt-2">Place:</h4>
              </v-col>
              <v-col :cols="searchPerformed ? 7 : 8" class="ma-0 pa-0">
                <v-autocomplete
                  outlined
                  dense
                  v-model="query.anyPlace"
                  :loading="anyPlaceLoading"
                  :items="anyPlaceItems"
                  :search-input.sync="anyPlaceSearch"
                  no-filter
                  auto-select-first
                  clearable
                  flat
                  hide-no-data
                  hide-details
                  solo
                  placeholder="Any place"
                  :menu-props="{ nudgeTop: autocompleteOffset }"
                  @change="anyPlaceChanged()"
                ></v-autocomplete>
              </v-col>
              <v-col
                v-if="query.anyPlace"
                cols="1"
                class="exactCheck d-flex flex-row ml-0 pl-0"
                :class="searchPerformed ? 'mt-1' : 'mt-1'"
              >
                <v-checkbox
                  v-model="query.anyPlaceFuzziness"
                  :value="1"
                  class="shrink mt-0 smallCheckbox"
                  dense
                  primary
                  hide-details="true"
                >
                </v-checkbox
                ><span class="mt-2 primary--text">Exact</span>
              </v-col>
            </v-row>
            <!--Keywords-->
            <v-row no-gutters class="mt-4" v-if="hasField('keywords')">
              <v-col cols="3">
                <h4 class="mt-2">Keywords:</h4>
              </v-col>
              <v-col cols="9">
                <v-text-field
                  outlined
                  dense
                  v-model="query.keywords"
                  type="text"
                  placeholder="Other text"
                  class="ma-0 mb-n2"
                ></v-text-field>
              </v-col>
            </v-row>
            <!--Buttons-->
            <v-row class="d-flex flex-row">
              <v-btn class="mt-2 mb-4 ml-3" type="submit" color="primary"
                ><span v-if="!searchPerformed">Search</span><span v-if="searchPerformed">Update</span></v-btn
              >
              <v-spacer v-if="searchPerformed"></v-spacer>
              <v-btn v-if="searchPerformed" text to="/" class="mt-2"
                ><v-icon left>mdi-close-circle-outline</v-icon>Clear all</v-btn
              >
            </v-row>
          </v-form>
        </v-col>
      </v-row>
    </v-col>
  </v-row>
</template>

<script>
import SearchResult from "../components/SearchResult.vue";
import Server from "@/services/Server.js";
import { mapState } from "vuex";
import store from "@/store";

function decodeFuzziness(f) {
  let result = [];
  for (let i = 32; i > 0; i = i / 2) {
    if (f >= i) {
      result.push(i);
      f -= i;
    }
  }
  if (result.length === 0) {
    result.push(0);
  }
  return result;
}

const defaultCategory =
  typeof window.ourroots.category === "string" && window.ourroots.category.length > 0 ? window.ourroots.category : "";
const defaultCollection =
  typeof window.ourroots.collection === "string" && window.ourroots.collection.length > 0
    ? window.ourroots.collection
    : "";
const surnameFirst = typeof window.ourroots.surnameFirst === "string" && window.ourroots.surnameFirst.length > 0;

const defaultQuery = {
  size: 0,
  collectionPlace1Facet: true,
  category: defaultCategory,
  collection: defaultCollection,
  categoryFacet: !defaultCategory,
  collectionFacet: !!defaultCategory && !defaultCollection,
  surnameFirst: surnameFirst
};

export default {
  components: {
    SearchResult
  },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    let query = Object.keys(routeTo.query).length > 0 ? routeTo.query : defaultQuery;
    store
      .dispatch("search", query)
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  beforeRouteUpdate(routeTo, routeFrom, next) {
    let query = Object.keys(routeTo.query).length > 0 ? routeTo.query : defaultQuery;
    store
      .dispatch("search", query)
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  mounted() {
    let app = document.getElementById("app");
    let appClientRect = null;
    if (app) {
      appClientRect = app.getBoundingClientRect();
      this.autocompleteOffset = Math.round(appClientRect.top + window.scrollY);
    }
    console.log(
      "appClientRect",
      appClientRect,
      "window.scrollY",
      window.scrollY,
      "autocompleteOffset",
      this.autocompleteOffset
    );
  },
  created() {
    console.log("created", window.ourroots);
    if (this.$route.query && Object.keys(this.$route.query).length > 0) {
      this.searchPerformed = true;
      this.query = Object.assign(this.query, this.$route.query);
      // convert fuzziness to integer
      for (let e of ["birth", "marriage", "death", "residence", "any"]) {
        for (let f of ["Date", "Place"]) {
          let item = e + f + "Fuzziness";
          this.query[item] = +this.query[item];
        }
      }
      for (let f in this.fuzziness) {
        this.fuzziness[f] = decodeFuzziness(this.query[f + "Fuzziness"]);
      }
      for (let e of ["birth", "marriage", "death", "residence", "any"]) {
        for (let f of ["Date", "Place"]) {
          if (this.query[e + f]) {
            this.showEvent[e] = true;
            if (f === "Place") {
              this[e + f + "Items"] = [this.query[e + f]];
            }
          }
        }
      }
      for (let r of ["father", "mother", "spouse", "other"]) {
        for (let f of ["Given", "Surname"]) {
          if (this.query[r + f]) {
            this.showRelative[r] = true;
          }
        }
      }
      this.page = Math.floor(this.query.from / this.pageSize) + 1;
    }
  },
  data() {
    return {
      defaultCategory: defaultCategory,
      defaultCollection: defaultCollection,
      autocompleteOffset: 0,
      editSearch: false,
      page: 1,
      pageSize: 10,
      surnameFirst: surnameFirst,
      showRelative: {
        father: false,
        mother: false,
        spouse: false,
        other: false
      },
      showEvent: {
        birth: false,
        marriage: false,
        residence: false,
        death: false,
        any: false
      },
      searchPerformed: false,
      query: {
        category: defaultCategory,
        collection: defaultCollection,
        collectionPlace1: "",
        collectionPlace2: "",
        collectionPlace3: "",
        given: "",
        surname: "",
        birthDate: "",
        birthPlace: "",
        deathDate: "",
        deathPlace: "",
        marriageDate: "",
        marriagePlace: "",
        residenceDate: "",
        residencePlace: "",
        anyDate: "",
        anyPlace: "",
        fatherGiven: "",
        fatherSurname: "",
        motherGiven: "",
        motherSurname: "",
        spouseGiven: "",
        spouseSurname: "",
        otherGiven: "",
        otherSurname: "",
        keywords: "",
        title: "",
        author: "",
        from: 0,
        size: 0,
        surnameFirst: surnameFirst,
        birthDateFuzziness: 0,
        marriageDateFuzziness: 0,
        residenceDateFuzziness: 0,
        deathDateFuzziness: 0,
        anyDateFuzziness: 0,
        birthPlaceFuzziness: 0,
        marriagePlaceFuzziness: 0,
        residencePlaceFuzziness: 0,
        deathPlaceFuzziness: 0,
        anyPlaceFuzziness: 0
      },
      fuzziness: {
        dlg: [],
        given: [1, 4, 32],
        surname: [1, 4],
        fatherGiven: [0],
        fatherSurname: [0],
        motherGiven: [0],
        motherSurname: [0],
        spouseGiven: [0],
        spouseSurname: [0],
        otherGiven: [0],
        otherSurname: [0]
      },
      dateRanges: [
        { value: 0, text: "Optional" },
        { value: 1, text: "Exact to this year" },
        { value: 2, text: "+/- 1 year" },
        { value: 3, text: "+/- 2 years" },
        { value: 4, text: "+/- 5 years" },
        { value: 5, text: "+/- 10 years" }
      ],
      givenFuzzinessLevels: [
        { value: 0, text: "Optional" },
        { value: 1, text: "Exact spelling" },
        { value: 2, text: "Alternate spellings" },
        { value: 4, text: "Sounds like (nysiis)" },
        { value: 8, text: "Sounds like (soundex)" },
        { value: 16, text: "Fuzzy" },
        { value: 32, text: "Initials" }
      ],
      surnameFuzzinessLevels: [
        { value: 0, text: "Optional" },
        { value: 1, text: "Exact" },
        { value: 2, text: "Alternate spellings" },
        { value: 4, text: "Sounds like (nysiis)" },
        { value: 8, text: "Sounds like (soundex)" },
        { value: 16, text: "Fuzzy" }
      ],
      placeFuzzinessLevels: [
        { value: 0, text: "Optional" },
        { value: 1, text: "Exact" },
        { value: 3, text: "Exact and higher-level places" }
      ],
      wildcardRegex: /[~*?]/,
      defaultPlace: "",
      placeTimeout: null,
      birthPlaceSearch: "",
      marriagePlaceSearch: "",
      residencePlaceSearch: "",
      deathPlaceSearch: "",
      defaultPlaceSearch: "",
      anyPlaceSearch: "",
      birthPlaceItems: [],
      marriagePlaceItems: [],
      residencePlaceItems: [],
      deathPlaceItems: [],
      defaultPlaceItems: [],
      anyPlaceItems: [],
      birthPlaceLoading: false,
      marriagePlaceLoading: false,
      residencePlaceLoading: false,
      deathPlaceLoading: false,
      defaultPlaceLoading: false,
      anyPlaceLoading: false,
      //option menus
      givenOptionsMenu: false,
      surnameOptionsMenu: false,
      placeOptionsMenu: false,
      birthOptionsMenu: false,
      birthOptionsMenu2: false,
      fields: (typeof window.ourroots.fields === "string" && window.ourroots.fields.length > 0
        ? window.ourroots.fields
        : "given,surname,events,relationships,keywords"
      ).split(/\s*,\s*/)
    };
  },
  computed: {
    formattedFrom() {
      return parseInt(this.query.from, 10);
    },
    formattedLength() {
      return parseInt(this.search.searchList.length, 10);
    },
    formattedSearchTotal() {
      return parseInt(this.search.searchTotal, 10);
    },
    numPages() {
      return Math.ceil(this.search.searchTotal / this.pageSize);
    },
    placeFacet() {
      let key = null;
      if (this.search.searchFacets.collectionPlace1) {
        key = "collectionPlace1";
      } else if (this.search.searchFacets.collectionPlace2) {
        key = "collectionPlace2";
      } else if (this.search.searchFacets.collectionPlace3) {
        key = "collectionPlace3";
      }
      return key ? { key, buckets: this.search.searchFacets[key].buckets } : null;
    },
    categoryFacet() {
      let key = null;
      if (this.search.searchFacets.category) {
        key = "category";
      } else if (this.search.searchFacets.collection) {
        key = "collection";
      }
      if (key === null) {
        return null;
      }
      let buckets = !this.search.searchFacets[key].buckets ? [] : [...this.search.searchFacets[key].buckets];
      let sortFn = (a, b) => {
        return a.label < b.label ? -1 : a.label > b.label ? 1 : 0;
      };
      buckets.sort(sortFn);
      return { key, buckets };
    },
    givenSpellingOptions() {
      if (this.fuzziness.given.length === 1 && this.fuzziness.given[0] === 0) {
        return "Spelling Options";
      }
      return this.fuzziness.given.map(f => this.givenFuzzinessLevels.find(l => l.value === f).text).join(" & ");
    },
    surnameSpellingOptions() {
      if (this.fuzziness.surname.length === 1 && this.fuzziness.surname[0] === 0) {
        return "Spelling Options";
      }
      return this.fuzziness.surname.map(f => this.surnameFuzzinessLevels.find(l => l.value === f).text).join(" & ");
    },
    birthDateFuzzinessText() {
      return this.query.birthDateFuzziness === 0
        ? "Exactness"
        : this.dateRanges.find(d => d.value === this.query.birthDateFuzziness).text;
    },
    eventsLabel() {
      return this.hasField("title") || this.hasField("author") ? "Places" : "Events";
    },
    relationshipsLabel() {
      return this.hasField("title") || this.hasField("author") ? "Surnames" : "Relationships";
    },
    ...mapState(["search"])
  },
  watch: {
    birthPlaceSearch(val) {
      val && val !== this.query.birthPlace && this.placeSearch(val, "birthPlace");
    },
    marriagePlaceSearch(val) {
      val && val !== this.query.marriagePlace && this.placeSearch(val, "marriagePlace");
    },
    residencePlaceSearch(val) {
      val && val !== this.query.residencePlace && this.placeSearch(val, "residencePlace");
    },
    deathPlaceSearch(val) {
      val && val !== this.query.deathPlace && this.placeSearch(val, "deathPlace");
    },
    defaultPlaceSearch(val) {
      val && val !== this.defaultPlace && this.placeSearch(val, "defaultPlace");
    },
    anyPlaceSearch(val) {
      val && val !== this.query.anyPlace && this.placeSearch(val, "anyPlace");
    }
  },
  methods: {
    hasField(fld) {
      return this.fields.includes(fld);
    },
    clearRelative(relative) {
      this.showRelative[relative] = false;
      this.query[relative + "Given"] = null;
      this.query[relative + "Surname"] = null;
      this.fuzziness[relative + "Given"] = [0];
      this.fuzziness[relative + "Surname"] = [0];
    },
    clearEvent(event) {
      this.showEvent[event] = false;
      this.query[event + "Date"] = null;
      this.query[event + "Place"] = null;
      this.query[event + "DateFuzziness"] = 0;
      this.query[event + "PlaceFuzziness"] = 0;
      if (event === "any") {
        this.defaultPlace = this.defaultPlaceSearch = "";
      }
    },
    defaultPlaceChanged() {
      this.query.anyPlace = this.defaultPlace;
      this.anyPlaceSearch = this.defaultPlaceSearch;
      this.anyPlaceItems = this.defaultPlaceItems;
      if (this.query.anyPlace && !this.showEvent.any) {
        this.showEvent.any = true;
      } else if (!this.query.anyPlace && !this.query.anyDate && this.showEvent.any) {
        this.showEvent.any = false;
      }
      this.placeOptionsMenu = false;
    },
    anyPlaceChanged() {
      this.defaultPlace = this.query.anyPlace;
      this.defaultPlaceSearch = this.anyPlaceSearch;
      this.defaultPlaceItems = this.anyPlaceItems;
    },
    birthYearChanged() {
      if (this.query.birthDate && !this.showEvent.birth) {
        this.showEvent.birth = true;
      } else if (!this.query.birthPlace && !this.query.birthDate && this.showEvent.birth) {
        this.showEvent.birth = false;
      }
      this.birthOptionsMenu = false;
      this.birthOptionsMenu2 = false;
    },
    placeSearch(text, prefix) {
      if (this.placeTimeout) {
        clearTimeout(this.placeTimeout);
      }
      this[prefix + "Loading"] = true;
      this.placeTimeout = setTimeout(() => {
        this.placeTimeout = null;
        if (this.wildcardRegex.test(text)) {
          this[prefix + "Items"] = [text];
          this[prefix + "Loading"] = false;
        } else {
          Server.placeSearch(text)
            .then(res => {
              this[prefix + "Items"] = res.data.map(p => p.fullName);
            })
            .finally(() => {
              this[prefix + "Loading"] = false;
            });
        }
      }, 400);
    },
    nameFuzzinessChecked(value) {
      if (this.fuzziness.dlg.length === 0) {
        this.fuzziness.dlg = [0];
      } else if (this.fuzziness.dlg.length > 1 && this.fuzziness.dlg.indexOf(0) >= 0 && value === 0) {
        this.fuzziness.dlg = [0];
      } else if (this.fuzziness.dlg.length > 1 && this.fuzziness.dlg.indexOf(0) >= 0) {
        this.fuzziness.dlg.splice(this.fuzziness.dlg.indexOf(0), 1);
      }
    },
    nameFuzzinessChanged(nameType) {
      this.fuzziness[nameType] = this.fuzziness.dlg.slice(0);
      this.givenOptionsMenu = false;
      this.surnameOptionsMenu = false;
    },
    openNameFuzziness(nameType) {
      this.fuzziness.dlg = this.fuzziness[nameType].slice(0);
    },
    getQuery(facetKey, facetValue) {
      let query = Object.assign({}, this.query);
      // set fuzziness
      for (let f of [
        "given",
        "surname",
        "fatherGiven",
        "fatherSurname",
        "motherGiven",
        "motherSurname",
        "spouseGiven",
        "spouseSurname",
        "otherGiven",
        "otherSurname",
        "birthDate",
        "marriageDate",
        "residenceDate",
        "deathDate",
        "anyDate",
        "birthPlace",
        "marriagePlace",
        "residencePlace",
        "deathPlace",
        "anyPlace"
      ]) {
        if (!f.endsWith("Date") && !f.endsWith("Place")) {
          query[f + "Fuzziness"] = this.fuzziness[f].reduce((acc, val) => acc + +val, 0);
        }
        if (query[f + "Fuzziness"] === 0) {
          delete query[f + "Fuzziness"];
        }
      }

      if (facetKey) {
        if (facetValue) {
          query[facetKey] = facetValue;
        } else {
          switch (facetKey) {
            case "collectionPlace1":
              delete query["collectionPlace1"];
            // eslint-disable-next-line no-fallthrough
            case "collectionPlace2":
              delete query["collectionPlace2"];
            // eslint-disable-next-line no-fallthrough
            case "collectionPlace3":
              delete query["collectionPlace3"];
              break;
            case "category":
              delete query["category"];
            // eslint-disable-next-line no-fallthrough
            case "collection":
              delete query["collection"];
          }
        }
      }

      delete query["collectionPlace1Facet"];
      delete query["collectionPlace2Facet"];
      delete query["collectionPlace3Facet"];
      delete query["categoryFacet"];
      delete query["collectionFacet"];

      if (!query["collectionPlace1"]) {
        query["collectionPlace1Facet"] = true;
      } else if (!query["collectionPlace2"]) {
        query["collectionPlace2Facet"] = true;
      } else if (!query["collectionPlace3"]) {
        query["collectionPlace3Facet"] = true;
      }
      if (!query["category"]) {
        query["categoryFacet"] = true;
      } else if (!query["collection"]) {
        query["collectionFacet"] = true;
      }

      query.from = (this.page - 1) * this.pageSize;
      query.size = this.pageSize;
      query.surnameFirst = surnameFirst;

      return query;
    },
    pageChanged() {
      if ((this.page - 1) * this.pageSize !== this.query.from) {
        this.issueQuery();
      }
    },
    go() {
      this.page = 1;
      this.issueQuery();
    },
    issueQuery() {
      let query = this.getQuery();

      // issue query
      this.$router.push({
        name: "search",
        query: query
      });
    }
  }
};
</script>

<style scoped>
/* .v-checkbox {
  padding: 0px;
  font-size: 50%;
}
.v-checkbox label {
  font-size: 50%;
} */
.search {
  width: 100%;
}
.edit-search {
  margin-left: -32px;
  margin-bottom: -64px;
}
.displayChosenOptions {
  max-width: 370px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.exactCheck {
  font-size: 0.625rem;
  text-transform: uppercase;
  font-weight: 500;
  letter-spacing: 0.0892857143em;
}
.exactnessOptions {
  background: #ffffff;
  padding: 16px;
}
.exactnessActions {
  background: #fff;
  margin-top: -16px;
}
.eventButton {
  margin-top: -6px;
}
.resultsHeader {
  padding: 8px 0;
  background: #f1f1f1;
}
.result {
  width: 100%;
  padding: 12px;
  border-top: solid 1px #e6e6e6;
}
.result a {
  text-decoration: none;
}

/* .result:nth-child(odd) {
  background-color: #f7f7f7;
}
.result:nth-child(even) {
  background-color: #ffffff;
} */
</style>
