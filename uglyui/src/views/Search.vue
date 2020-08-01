<template>
  <div class="search">
    <h1>Search</h1>
    <v-form @submit.prevent="go">
      <v-row>
        <v-col cols="2">
          <h4>Primary</h4>
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
                v-model="fuzziness.principalGiven"
                :change="nameFuzzinessChanged('principalGiven')"
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
                v-model="fuzziness.principalSurname"
                :change="nameFuzzinessChanged('principalSurname')"
                label="Exactness"
              ></v-select>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
      <!--Father-->
      <v-row>
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
      </v-row>
      <!--Mother-->
      <v-row>
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
      </v-row>
      <!--Spouse-->
      <v-row>
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
      </v-row>
      <!--Other-->
      <v-row>
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
      </v-row>

      <!--Birth-->
      <v-row>
        <v-col cols="2">
          <h4>Birth</h4>
        </v-col>
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
      <!--Residence-->
      <v-row>
        <v-col cols="2">
          <h4>Residence</h4>
        </v-col>
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
          <v-select outlined :items="dateRanges" v-model="query.residenceDateFuzziness" label="Exactness"></v-select>
        </v-col>
      </v-row>
      <!--Marriage-->
      <v-row>
        <v-col cols="2">
          <h4>Marriage</h4>
        </v-col>
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
      <!--Death-->
      <v-row>
        <v-col cols="2">
          <h4>Death</h4>
        </v-col>
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
      <!--Any-->
      <v-row>
        <v-col cols="2">
          <h4>Any</h4>
        </v-col>
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

      <v-btn class="mt-2 mb-4" type="submit" color="primary">Go</v-btn>
    </v-form>

    <v-row class="pa-3" v-if="searchPerformed && search.searchTotal === 0">
      <p>No results found</p>
    </v-row>
    <v-row v-if="searchPerformed && search.searchTotal > 0">
      <p>Showing 1 - {{ search.searchList.length }} of {{ search.searchTotal }}</p>
    </v-row>
    <v-row v-for="(result, $ix) in search.searchList" :key="$ix">
      <SearchResult :result="result" />
    </v-row>
  </div>
</template>

<script>
import SearchResult from "../components/SearchResult.vue";
import NProgress from "nprogress";
import { mapState } from "vuex";

export default {
  components: {
    SearchResult
  },
  data() {
    return {
      searchPerformed: false,
      query: {
        birthDateFuzziness: 0,
        marriageDateFuzziness: 0,
        residenceDateFuzziness: 0,
        deathDateFuzziness: 0,
        anyDateFuzziness: 0
      },
      fuzziness: {
        principalGiven: [0],
        principalSurname: [0],
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
      ]
    };
  },
  computed: mapState(["search"]),
  methods: {
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
    go() {
      console.log("this.fuzziness", this.fuzziness);
      this.query.givenFuzziness = this.fuzziness.principalGiven.reduce((acc, val) => acc + +val, 0);
      this.query.surnameFuzziness = this.fuzziness.principalSurname.reduce((acc, val) => acc + +val, 0);
      this.query.fatherGivenFuzziness = this.fuzziness.fatherGiven.reduce((acc, val) => acc + +val, 0);
      this.query.fatherSurnameFuzziness = this.fuzziness.fatherSurname.reduce((acc, val) => acc + +val, 0);
      this.query.motherGivenFuzziness = this.fuzziness.motherGiven.reduce((acc, val) => acc + +val, 0);
      this.query.motherSurnameFuzziness = this.fuzziness.motherSurname.reduce((acc, val) => acc + +val, 0);
      this.query.spouseGivenFuzziness = this.fuzziness.spouseGiven.reduce((acc, val) => acc + +val, 0);
      this.query.spouseSurnameFuzziness = this.fuzziness.spouseSurname.reduce((acc, val) => acc + +val, 0);
      this.query.otherGivenFuzziness = this.fuzziness.otherGiven.reduce((acc, val) => acc + +val, 0);
      this.query.otherSurnameFuzziness = this.fuzziness.otherSurname.reduce((acc, val) => acc + +val, 0);
      NProgress.start();
      this.$store
        .dispatch("search", this.query)
        .then(() => {
          this.searchPerformed = true;
          NProgress.done();
        })
        .catch(() => {
          NProgress.done();
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
