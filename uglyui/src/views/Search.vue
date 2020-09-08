<template>
  <div class="search">
    <h1>Search</h1>
    <v-form @submit.prevent="go">
      <v-row>
        <v-col cols="2">
          <h4>Name</h4>
        </v-col>
        <v-col>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.given"
                type="text"
                placeholder="Given name"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="givenFuzzinessLevels"
                v-model="fuzziness.given"
                :change="nameFuzzinessChanged('given')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                dense
                outlined
                v-model="query.surname"
                type="text"
                placeholder="Surname"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="surnameFuzzinessLevels"
                v-model="fuzziness.surname"
                :change="nameFuzzinessChanged('surname')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
      </v-row>

      <!-- any place and birth year -->
      <v-row>
        <v-col cols="2">
          <h4>Place your ancestor might have lived</h4>
        </v-col>
        <v-col>
          <v-row>
            <v-col>
              <v-autocomplete
                outlined
                dense
                v-model="query.anyPlace"
                :loading="anyPlaceLoading"
                :items="anyPlaceItems"
                :search-input.sync="anyPlaceSearch"
                no-filter
                auto-select-first
                flat
                hide-no-data
                hide-details
                solo
                placeholder="Any place"
                class="ma-0 mb-n2"
                @change="anyPlaceChanged()"
              ></v-autocomplete>
            </v-col>
            <v-col>
              <v-select
                outlined
                :items="placeFuzzinessLevels"
                v-model="query.anyPlaceFuzziness"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="2">
          <h4>Birth year</h4>
        </v-col>
        <v-col>
          <v-row>
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.birthDate"
                type="text"
                placeholder="Birth year"
                class="ma-0 mb-n2"
                @change="birthYearChanged()"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select outlined :items="dateRanges" v-model="query.birthDateFuzziness" label="Exactness"></v-select>
            </v-col>
          </v-row>
        </v-col>
      </v-row>

      <!--Events-->
      <v-row>
        <v-col>
          <strong>Add event:</strong>
          <v-btn text color="primary" :disabled="showEvent.birth" @click="showEvent.birth = true">Birth</v-btn>
          <v-btn text color="primary" :disabled="showEvent.marriage" @click="showEvent.marriage = true">Marriage</v-btn>
          <v-btn text color="primary" :disabled="showEvent.death" @click="showEvent.death = true">Death</v-btn>
          <v-btn text color="primary" :disabled="showEvent.residence" @click="showEvent.residence = true"
            >Lived In</v-btn
          >
          <v-btn text color="primary" :disabled="showEvent.any" @click="showEvent.any = true">Any Event</v-btn>
        </v-col>
      </v-row>
      <!--Birth-->
      <v-row v-if="showEvent.birth">
        <v-col cols="2">
          <h4>Birth</h4>
        </v-col>
        <v-col>
          <v-row>
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.birthDate"
                type="text"
                placeholder="Birth year"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select outlined :items="dateRanges" v-model="query.birthDateFuzziness" label="Exactness"></v-select>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-autocomplete
                outlined
                dense
                v-model="query.birthPlace"
                :loading="birthPlaceLoading"
                :items="birthPlaceItems"
                :search-input.sync="birthPlaceSearch"
                no-filter
                auto-select-first
                flat
                hide-no-data
                hide-details
                solo
                placeholder="Birth place"
                class="ma-0 mb-n2"
              ></v-autocomplete>
            </v-col>
            <v-col>
              <v-select
                outlined
                :items="placeFuzzinessLevels"
                v-model="query.birthPlaceFuzziness"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="1">
          <v-btn text @click="clearEvent('birth')">X</v-btn>
        </v-col>
      </v-row>
      <!--Marriage-->
      <v-row v-if="showEvent.marriage">
        <v-col cols="2">
          <h4>Marriage</h4>
        </v-col>
        <v-col>
          <v-row>
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.marriageDate"
                type="text"
                placeholder="Marriage year"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select outlined :items="dateRanges" v-model="query.marriageDateFuzziness" label="Exactness"></v-select>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-autocomplete
                outlined
                dense
                v-model="query.marriagePlace"
                :loading="marriagePlaceLoading"
                :items="marriagePlaceItems"
                :search-input.sync="marriagePlaceSearch"
                no-filter
                auto-select-first
                flat
                hide-no-data
                hide-details
                solo
                placeholder="Marriage place"
                class="ma-0 mb-n2"
              ></v-autocomplete>
            </v-col>
            <v-col>
              <v-select
                outlined
                :items="placeFuzzinessLevels"
                v-model="query.marriagePlaceFuzziness"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="1">
          <v-btn text @click="clearEvent('marriage')">X</v-btn>
        </v-col>
      </v-row>
      <!--Death-->
      <v-row v-if="showEvent.death">
        <v-col cols="2">
          <h4>Death</h4>
        </v-col>
        <v-col>
          <v-row>
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.deathDate"
                type="text"
                placeholder="Death year"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select outlined :items="dateRanges" v-model="query.deathDateFuzziness" label="Exactness"></v-select>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-autocomplete
                outlined
                dense
                v-model="query.deathPlace"
                :loading="deathPlaceLoading"
                :items="deathPlaceItems"
                :search-input.sync="deathPlaceSearch"
                no-filter
                auto-select-first
                flat
                hide-no-data
                hide-details
                solo
                placeholder="Death place"
                class="ma-0 mb-n2"
              ></v-autocomplete>
            </v-col>
            <v-col>
              <v-select
                outlined
                :items="placeFuzzinessLevels"
                v-model="query.deathPlaceFuzziness"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="1">
          <v-btn text @click="clearEvent('death')">X</v-btn>
        </v-col>
      </v-row>
      <!--Residence-->
      <v-row v-if="showEvent.residence">
        <v-col cols="2">
          <h4>Lived In</h4>
        </v-col>
        <v-col>
          <v-row>
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.residenceDate"
                type="text"
                placeholder="Residence year"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :items="dateRanges"
                v-model="query.residenceDateFuzziness"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-autocomplete
                outlined
                dense
                v-model="query.residencePlace"
                :loading="residencePlaceLoading"
                :items="residencePlaceItems"
                :search-input.sync="residencePlaceSearch"
                no-filter
                auto-select-first
                flat
                hide-no-data
                hide-details
                solo
                placeholder="Residence place"
                class="ma-0 mb-n2"
              ></v-autocomplete>
            </v-col>
            <v-col>
              <v-select
                outlined
                :items="placeFuzzinessLevels"
                v-model="query.residencePlaceFuzziness"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="1">
          <v-btn text @click="clearEvent('residence')">X</v-btn>
        </v-col>
      </v-row>
      <!--Any-->
      <v-row v-if="showEvent.any">
        <v-col cols="2">
          <h4>Any Event</h4>
        </v-col>
        <v-col>
          <v-row>
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.anyDate"
                type="text"
                placeholder="Any year"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select outlined :items="dateRanges" v-model="query.anyDateFuzziness" label="Exactness"></v-select>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-autocomplete
                outlined
                dense
                v-model="query.anyPlace"
                :loading="anyPlaceLoading"
                :items="anyPlaceItems"
                :search-input.sync="anyPlaceSearch"
                no-filter
                auto-select-first
                flat
                hide-no-data
                hide-details
                solo
                placeholder="Any place"
                class="ma-0 mb-n2"
              ></v-autocomplete>
            </v-col>
            <v-col>
              <v-select
                outlined
                :items="placeFuzzinessLevels"
                v-model="query.anyPlaceFuzziness"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="1">
          <v-btn text @click="clearEvent('any')">X</v-btn>
        </v-col>
      </v-row>

      <!--Relatives-->
      <v-row>
        <v-col>
          <strong>Add family member:</strong>
          <v-btn text color="primary" :disabled="showRelative.father" @click="showRelative.father = true">Father</v-btn>
          <v-btn text color="primary" :disabled="showRelative.mother" @click="showRelative.mother = true">Mother</v-btn>
          <v-btn text color="primary" :disabled="showRelative.spouse" @click="showRelative.spouse = true">Spouse</v-btn>
          <v-btn text color="primary" :disabled="showRelative.other" @click="showRelative.other = true">Other</v-btn>
        </v-col>
      </v-row>
      <!--Father-->
      <v-row v-if="showRelative.father">
        <v-col cols="2">
          <h4 class="mb-1">Father</h4>
        </v-col>
        <v-col>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.fatherGiven"
                type="text"
                placeholder="Father's given name"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="givenFuzzinessLevels"
                v-model="fuzziness.fatherGiven"
                :change="nameFuzzinessChanged('fatherGiven')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.fatherSurname"
                type="text"
                placeholder="Father's surname"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="surnameFuzzinessLevels"
                v-model="fuzziness.fatherSurname"
                :change="nameFuzzinessChanged('fatherSurname')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="1">
          <v-btn text @click="clearRelative('father')">X</v-btn>
        </v-col>
      </v-row>
      <!--Mother-->
      <v-row v-if="showRelative.mother">
        <v-col cols="2">
          <h4 class="mb-1">Mother</h4>
        </v-col>
        <v-col>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.motherGiven"
                type="text"
                placeholder="Mother's given name"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="givenFuzzinessLevels"
                v-model="fuzziness.motherGiven"
                :change="nameFuzzinessChanged('motherGiven')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.motherSurname"
                type="text"
                placeholder="Mother's surname"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="surnameFuzzinessLevels"
                v-model="fuzziness.motherSurname"
                :change="nameFuzzinessChanged('motherSurname')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="1">
          <v-btn text @click="clearRelative('mother')">X</v-btn>
        </v-col>
      </v-row>
      <!--Spouse-->
      <v-row v-if="showRelative.spouse">
        <v-col cols="2">
          <h4 class="mb-1">Spouse</h4>
        </v-col>
        <v-col>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.spouseGiven"
                type="text"
                placeholder="Spouse's given name"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="givenFuzzinessLevels"
                v-model="fuzziness.spouseGiven"
                :change="nameFuzzinessChanged('spouseGiven')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.spouseSurname"
                type="text"
                placeholder="Spouse's surname"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="surnameFuzzinessLevels"
                v-model="fuzziness.spouseSurname"
                :change="nameFuzzinessChanged('spouseSurname')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="1">
          <v-btn text @click="clearRelative('spouse')">X</v-btn>
        </v-col>
      </v-row>
      <!--Other-->
      <v-row v-if="showRelative.other">
        <v-col cols="2">
          <h4 class="mb-1">Other person</h4>
        </v-col>
        <v-col>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.otherGiven"
                type="text"
                placeholder="Other person's given name"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="givenFuzzinessLevels"
                v-model="fuzziness.otherGiven"
                :change="nameFuzzinessChanged('otherGiven')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
          <v-row class="pa-0 ma-0">
            <v-col>
              <v-text-field
                outlined
                dense
                v-model="query.otherSurname"
                type="text"
                placeholder="Other person's surname"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                outlined
                :multiple="true"
                :items="surnameFuzzinessLevels"
                v-model="fuzziness.otherSurname"
                :change="nameFuzzinessChanged('otherSurname')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
        <v-col cols="1">
          <v-btn text @click="clearRelative('other')">X</v-btn>
        </v-col>
      </v-row>

      <!--Keywords-->
      <v-row>
        <v-col cols="2">
          <h4 class="mb-1">Keyword</h4>
        </v-col>
        <v-col>
          <v-text-field
            outlined
            dense
            v-model="query.keywords"
            type="text"
            placeholder="Occupation, etc."
            class="ma-0 mb-n2"
          ></v-text-field>
        </v-col>
      </v-row>

      <v-btn class="mt-2 mb-4" type="submit" color="primary">Go</v-btn>
    </v-form>

    <v-row class="pa-3" v-if="searchPerformed && search.searchTotal === 0">
      <v-col>
        <p>No results found</p>
      </v-col>
    </v-row>

    <v-row v-if="searchPerformed && search.searchTotal > 0">
      <v-col>
        <h4>All Categories</h4>
        <div v-if="query.category">
          <span>{{ query.category }}</span> &nbsp;
          <router-link :to="{ name: 'search', query: getQuery('category', null) }">remove</router-link>
        </div>
        <div v-if="query.collection">
          <span>{{ query.collection }}</span> &nbsp;
          <router-link :to="{ name: 'search', query: getQuery('collection', null) }">remove</router-link>
        </div>
        <ul v-if="categoryFacet">
          <li v-for="(bucket, $ix) in categoryFacet.buckets" :key="$ix">
            <router-link :to="{ name: 'search', query: getQuery(categoryFacet.key, bucket.label) }">{{
              bucket.label
            }}</router-link>
            {{ bucket.count }}
          </li>
        </ul>
      </v-col>
    </v-row>

    <v-row v-if="searchPerformed && search.searchTotal > 0">
      <v-col>
        <h4>Collection Location</h4>
        <div v-if="query.collectionPlace1">
          <span>{{ query.collectionPlace1 }}</span> &nbsp;
          <router-link :to="{ name: 'search', query: getQuery('collectionPlace1', null) }">remove</router-link>
        </div>
        <div v-if="query.collectionPlace2">
          <span>{{ query.collectionPlace2 }}</span> &nbsp;
          <router-link :to="{ name: 'search', query: getQuery('collectionPlace2', null) }">remove</router-link>
        </div>
        <div v-if="query.collectionPlace3">
          <span>{{ query.collectionPlace3 }}</span> &nbsp;
          <router-link :to="{ name: 'search', query: getQuery('collectionPlace3', null) }">remove</router-link>
        </div>
        <ul v-if="placeFacet">
          <li v-for="(bucket, $ix) in placeFacet.buckets" :key="$ix">
            <router-link :to="{ name: 'search', query: getQuery(placeFacet.key, bucket.label) }">{{
              bucket.label
            }}</router-link>
            {{ bucket.count }}
          </li>
        </ul>
      </v-col>
    </v-row>

    <v-row v-if="searchPerformed && search.searchTotal > 0">
      <v-col>
        <p>Showing {{ query.from + 1 }} - {{ query.from + search.searchList.length }} of {{ search.searchTotal }}</p>
      </v-col>
    </v-row>

    <v-row v-for="(result, $ix) in search.searchList" :key="$ix">
      <v-col>
        <SearchResult :result="result" />
      </v-col>
    </v-row>

    <v-row v-if="searchPerformed && search.searchTotal > 0">
      <v-col>
        <v-pagination v-model="page" :length="numPages" :total-visible="7" @input="pageChanged()"></v-pagination>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import SearchResult from "../components/SearchResult.vue";
import Server from "@/services/Server.js";
import { mapState } from "vuex";
import store from "@/store";

function decodeFuzziness(f) {
  let result = [0];
  for (let i = 32; i > 0; i = i / 2) {
    if (f >= i) {
      result.push(i);
      f -= i;
    }
  }
  return result;
}

export default {
  components: {
    SearchResult
  },
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    store
      .dispatch("search", routeTo.query)
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  beforeRouteUpdate(routeTo, routeFrom, next) {
    store
      .dispatch("search", routeTo.query)
      .then(() => {
        next();
      })
      .catch(() => {
        next("/");
      });
  },
  created() {
    if (this.$route.query && Object.keys(this.$route.query).length > 0) {
      this.searchPerformed = true;
      this.query = Object.assign(this.query, this.$route.query);
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
      page: 1,
      pageSize: 10,
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
        given: [0],
        surname: [0],
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
        { value: 0, text: "Default" },
        { value: 1, text: "Exact to this year" },
        { value: 2, text: "+/- 1 years" },
        { value: 3, text: "+/- 2 years" },
        { value: 4, text: "+/- 5 years" },
        { value: 5, text: "+/- 10 years" }
      ],
      givenFuzzinessLevels: [
        { value: 0, text: "Default" },
        { value: 1, text: "Exact" },
        { value: 2, text: "Alternate spellings" },
        { value: 4, text: "Sounds like (narrow)" },
        { value: 8, text: "Sounds like (broad)" },
        { value: 16, text: "Fuzzy" },
        { value: 32, text: "Initials" }
      ],
      surnameFuzzinessLevels: [
        { value: 0, text: "Default" },
        { value: 1, text: "Exact" },
        { value: 2, text: "Alternate spellings" },
        { value: 4, text: "Sounds like (narrow)" },
        { value: 8, text: "Sounds like (broad)" },
        { value: 16, text: "Fuzzy" }
      ],
      placeFuzzinessLevels: [
        { value: 0, text: "Default" },
        { value: 1, text: "Exact" },
        { value: 3, text: "Exact and higher-level places" }
      ],
      wildcardRegex: /[~*?]/,
      placeTimeout: null,
      birthPlaceSearch: "",
      marriagePlaceSearch: "",
      residencePlaceSearch: "",
      deathPlaceSearch: "",
      anyPlaceSearch: "",
      birthPlaceItems: [],
      marriagePlaceItems: [],
      residencePlaceItems: [],
      deathPlaceItems: [],
      anyPlaceItems: [],
      birthPlaceLoading: false,
      marriagePlaceLoading: false,
      residencePlaceLoading: false,
      deathPlaceLoading: false,
      anyPlaceLoading: false
    };
  },
  computed: {
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
      return key ? { key, buckets: this.search.searchFacets[key].buckets } : null;
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
    anyPlaceSearch(val) {
      val && val !== this.query.anyPlace && this.placeSearch(val, "anyPlace");
    }
  },
  methods: {
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
    },
    anyPlaceChanged() {
      if (this.query.anyPlace && !this.showEvent.any) {
        this.showEvent.any = true;
      } else if (!this.query.anyPlace && !this.query.anyDate && this.showEvent.any) {
        this.showEvent.any = false;
      }
    },
    birthYearChanged() {
      if (this.query.birthDate && !this.showEvent.birth) {
        this.showEvent.birth = true;
      } else if (!this.query.birthPlace && !this.query.birthDate && this.showEvent.birth) {
        this.showEvent.birth = false;
      }
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
    nameFuzzinessChanged(fuzziness) {
      if (this.fuzziness[fuzziness].length === 0) {
        this.fuzziness[fuzziness] = [0];
      } else if (
        this.fuzziness[fuzziness].length > 1 &&
        this.fuzziness[fuzziness].indexOf(0) === this.fuzziness[fuzziness].length - 1
      ) {
        this.fuzziness[fuzziness] = [0];
      } else if (this.fuzziness[fuzziness].length > 1 && this.fuzziness[fuzziness].indexOf(0) >= 0) {
        this.fuzziness[fuzziness].splice(this.fuzziness[fuzziness].indexOf(0), 1);
      }
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
.v-checkbox {
  padding: 0px;
  font-size: 50%;
}
.v-checkbox label {
  font-size: 50%;
}
</style>
