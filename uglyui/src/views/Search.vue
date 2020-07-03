<template>
  <div class="search">
    <h1>Search</h1>
    <v-form @submit.prevent="go">
      <v-row>
        <v-col cols="1">
          <h4>Primary</h4>
        </v-col>
        <v-col>
          <v-row class="pa-0 ma-0">
          <v-text-field outlined dense v-model="query.given" type="text" placeholder="Given name" class="ma-0 mb-n2"
          :value="value" @input="updateValue" v-bind="$attrs" v-on="listeners"
          ></v-text-field>
          </v-row>    
          <v-row class="pa-0 ma-0 pl-1 mt-n5">
              <v-checkbox
              class="mt-0" 
              type="checkbox" 
              id="principalGivenAlt" 
              v-model="fuzziness.principalGiven" 
              value="1"
              label="Alternate spellings"
            ></v-checkbox>
            <!--input & label-->
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="principalGivenSoundsLikeNarrow"
              v-model="fuzziness.principalGiven"
              value="2"
              label="Sounds like (narrow)"
            ></v-checkbox>
            <!--input & label-->
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="principalGivenSoundsLikeBroad"
              v-model="fuzziness.principalGiven"
              value="4"
              label="Sounds like (broad)"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="principalGivenFuzzy" 
              v-model="fuzziness.principalGiven" 
              value="8"
              label="Fuzzy"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="principalGivenInitials"
              v-model="fuzziness.principalGiven"
              value="16"
              label="Initials"
            ></v-checkbox>
          </v-row>
          <v-row class="pa-0 ma-0">
            <v-text-field dense outlined v-model="query.surname" type="text" placeholder="Surname" class="ma-0 mb-n2"
            :value="value" @input="updateValue" v-bind="$attrs" v-on="listeners"
            ></v-text-field>
          </v-row>
          <v-row class="pa-0 ma-0 pl-1 mt-n5">
            <v-checkbox 
              class="mt-0" 
              type="checkbox" 
              id="principalSurnameAlt" 
              v-model="fuzziness.principalSurname" 
              value="1"
              label="Alternate spellings"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="principalSurnameSoundsLikeNarrow"
              v-model="fuzziness.principalSurname"
              value="2"
              label="Sounds like (narrow)"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="principalSurnameSoundsLikeBroad"
              v-model="fuzziness.principalSurname"
              value="4"
              label="Sounds like (broad)"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="principalSurnameFuzzy"
              v-model="fuzziness.principalSurname"
              value="8"
              label="Fuzzy"
            ></v-checkbox>
          </v-row>  
        </v-col>
      </v-row>
      <!--Father-->
      <v-row>
        <v-col cols="1">
          <h4 class="mb-1">Father</h4>
        </v-col>
        <v-col>
          <v-row class="pa-0 ma-0">
          <v-text-field outlined dense v-model="query.fatherGiven" type="text" placeholder="Father's given name" class="ma-0 mb-n2"
          :value="value" @input="updateValue" v-bind="$attrs" v-on="listeners"
          ></v-text-field>
        </v-row>
        <v-row class="pa-0 ma-0 pl-1 mt-n5">
          <v-checkbox 
            class="mt-0" 
            type="checkbox" 
            id="fatherGivenAlt" 
            v-model="fuzziness.fatherGiven" 
            value="1" 
            label="Alternate spellings"
          ></v-checkbox>
          <v-checkbox
            class="mt-0 ml-3"
            type="checkbox"
            id="fatherGivenSoundsLikeNarrow"
            v-model="fuzziness.fatherGiven"
            value="2"
            label="Sounds like (narrow)"
          ></v-checkbox>
          <v-checkbox
            class="mt-0 ml-3"
            type="checkbox"
            id="fatherGivenSoundsLikeBroad"
            v-model="fuzziness.fatherGiven"
            value="4"
            label="Sounds like (broad)"
          ></v-checkbox>
          <v-checkbox 
            class="mt-0 ml-3" 
            type="checkbox" 
            id="fatherGivenFuzzy" 
            v-model="fuzziness.fatherGiven" 
            value="8" 
            label="Fuzzy"
          ></v-checkbox>
          <v-checkbox 
            class="mt-0 ml-3" 
            type="checkbox" 
            id="fatherGivenInitials" 
            v-model="fuzziness.fatherGiven" 
            value="16" 
            label="Initials"
            ></v-checkbox>
        </v-row>
        <v-row class="pa-0 ma-0">
          <v-text-field outlined dense v-model="query.fatherSurname" type="text" placeholder="Father's surname" class="ma-0 mb-n2"></v-text-field>
        </v-row>
        <v-row class="pa-0 ma-0 pl-1 mt-n5">  
          <v-checkbox 
            class="mt-0" 
            type="checkbox" 
            id="fatherSurnameAlt" 
            v-model="fuzziness.fatherSurname" 
            value="1"
            label="Alternate spellings"
          ></v-checkbox>
          <v-checkbox
            class="mt-0 ml-3"
            type="checkbox"
            id="fatherSurnameSoundsLikeNarrow"
            v-model="fuzziness.fatherSurname"
            value="2"
            label="Sounds like (narrow)"
          ></v-checkbox>
          <v-checkbox
            class="mt-0 ml-3"
            type="checkbox"
            id="fatherSurnameSoundsLikeBroad"
            v-model="fuzziness.fatherSurname"
            value="4"
            label="Sounds like (broad)"
          ></v-checkbox>
          <v-checkbox 
            class="mt-0 ml-3" 
            type="checkbox" 
            id="fatherSurnameFuzzy" 
            v-model="fuzziness.fatherSurname" 
            value="8" 
            label="Fuzzy"/>
        </v-row>
        </v-col>
      </v-row>
      <!--Mother-->
      <v-row>
        <v-col cols="1">
          <h4>Mother</h4>
        </v-col>  
        <v-col>
          <v-row class="pa-0 ma-0">
            <v-text-field outlined dense v-model="query.motherGiven" type="text" placeholder="Mother's given name" class="ma-0 mb-n2"
            :value="value" @input="updateValue" v-bind="$attrs" v-on="listeners"
            ></v-text-field>
          </v-row>
          <v-row class="pa-0 ma-0 pl-1 mt-n5">
            <v-checkbox 
              class="mt-0" 
              type="checkbox" 
              id="motherGivenAlt" 
              v-model="fuzziness.motherGiven" 
              value="1" 
              label="Alternate spellings"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="motherGivenSoundsLikeNarrow"
              v-model="fuzziness.motherGiven"
              value="2"
              label="Sounds like (narrow)"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="motherGivenSoundsLikeBroad"
              v-model="fuzziness.motherGiven"
              value="4"
              label="Sounds like (broad)"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="motherGivenFuzzy" 
              v-model="fuzziness.motherGiven" 
              value="8" 
              label="Fuzzy"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="motherGivenInitials" 
              v-model="fuzziness.motherGiven" 
              value="16" 
              label="Initials"
            ></v-checkbox>
          </v-row>
          <v-row class="pa-0 ma-0">
            <v-text-field outlined dense v-model="query.motherSurname" type="text" placeholder="Mother's surname" class="ma-0 mb-n2"
            :value="value" @input="updateValue" v-bind="$attrs" v-on="listeners"
            ></v-text-field>
          </v-row>
          <v-row class="pa-0 ma-0 pl-1 mt-n5">  
            <v-checkbox 
              class="mt-0" 
              type="checkbox" 
              id="motherSurnameAlt" 
              v-model="fuzziness.motherSurname" 
              value="1"
              label="Alternate spellings"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="motherSurnameSoundsLikeNarrow"
              v-model="fuzziness.motherSurname"
              value="2"
              label="Sounds like (narrow)"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="motherSurnameSoundsLikeBroad"
              v-model="fuzziness.motherSurname"
              value="4"
              label="Sounds like (broad)"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="motherSurnameFuzzy" 
              v-model="fuzziness.motherSurname" 
              value="8" 
              label="Fuzzy"
            ></v-checkbox>
          </v-row>
        </v-col>  
      </v-row>
      <!--Spouse-->
      <v-row>
        <v-col cols="1">
          <h4>Spouse</h4>
        </v-col>
        <v-col>
          <v-row class="pa-0 ma-0">
            <v-text-field outlined dense v-model="query.spouseGiven" type="text" placeholder="Spouse's given name" class="ma-0 mb-n2"
            :value="value" @input="updateValue" v-bind="$attrs" v-on="listeners"
            ></v-text-field>
          </v-row>
          <v-row class="pa-0 ma-0 pl-1 mt-n5">
            <v-checkbox 
              class="mt-0" 
              type="checkbox" 
              id="spouseGivenAlt" 
              v-model="fuzziness.spouseGiven" 
              value="1" 
              label="Alternate spellings"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="spouseGivenSoundsLikeNarrow"
              v-model="fuzziness.spouseGiven"
              value="2"
              label="Sounds like (narrow)"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="spouseGivenSoundsLikeBroad"
              v-model="fuzziness.spouseGiven"
              value="4"
              label="Sounds like (broad)"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="spouseGivenFuzzy" 
              v-model="fuzziness.spouseGiven" 
              value="8" 
              label="Fuzzy"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="spouseGivenInitials" 
              v-model="fuzziness.spouseGiven" 
              value="16"
              label="Initials" 
            ></v-checkbox>
            </v-row>
          <v-row class="pa-0 ma-0">
            <v-text-field outlined dense v-model="query.spouseSurname" type="text" placeholder="Spouse's surname" class="ma-0 mb-n2"
            :value="value" @input="updateValue" v-bind="$attrs" v-on="listeners"
            ></v-text-field>
          </v-row>
          <v-row class="pa-0 ma-0 pl-1 mt-n5">
            <v-checkbox 
              class="mt-0" 
              type="checkbox" 
              id="spouseSurnameAlt" 
              v-model="fuzziness.spouseSurname" 
              value="1" 
              label="Alternate spellings"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="spouseSurnameSoundsLikeNarrow"
              v-model="fuzziness.spouseSurname"
              value="2"
              label="Sounds like (narrow)"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="spouseSurnameSoundsLikeBroad"
              v-model="fuzziness.spouseSurname"
              value="4"
              label="Sounds like (broad)"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="spouseSurnameFuzzy" 
              v-model="fuzziness.spouseSurname" 
              value="8" 
              label="Fuzzy"
            ></v-checkbox>
          </v-row>
        </v-col>  
      </v-row>  
      <!--Other-->
      <v-row>
        <v-col cols="1">
          <h4>Other person</h4>
        </v-col>
        <v-col>
          <v-row class="pa-0 ma-0">
            <v-text-field outlined dense v-model="query.otherGiven" type="text" placeholder="Other person's given name" class="ma-0 mb-n2"
            :value="value" @input="updateValue" v-bind="$attrs" v-on="listeners"
            ></v-text-field>
          </v-row>
          <v-row class="pa-0 ma-0 pl-1 mt-n5">
            <v-checkbox 
              class="mt-0" 
              type="checkbox" 
              id="otherGivenAlt" 
              v-model="fuzziness.otherGiven" 
              value="1" 
              label="Alternate spellings"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="otherGivenSoundsLikeNarrow" 
              v-model="fuzziness.otherGiven" 
              value="2" 
              label="Sounds like (narrow)"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="otherGivenSoundsLikeBroad" 
              v-model="fuzziness.otherGiven" 
              value="4" 
              label="Sounds like (broad)"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="otherGivenFuzzy" 
              v-model="fuzziness.otherGiven" 
              value="8" 
              label="Fuzzy"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="otherGivenInitials" 
              v-model="fuzziness.otherGiven" 
              value="16" 
              label="Initials"
            ></v-checkbox>
          </v-row>
          <v-row class="pa-0 ma-0">
            <v-text-field outlined dense v-model="query.otherSurname" type="text" placeholder="Other person's surname" class="ma-0 mb-n2"
            :value="value" @input="updateValue" v-bind="$attrs" v-on="listeners"
            ></v-text-field>
          </v-row>
          <v-row class="pa-0 ma-0 pl-1 mt-n5">
            <v-checkbox 
              class="mt-0" 
              type="checkbox" 
              id="otherSurnameAlt" 
              v-model="fuzziness.otherSurname" 
              value="1"
              label="Alternate spellings"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="otherSurnameSoundsLikeNarrow"
              v-model="fuzziness.otherSurname"
              value="2"
              label="Sounds like (narrow)"
            ></v-checkbox>
            <v-checkbox
              class="mt-0 ml-3"
              type="checkbox"
              id="otherSurnameSoundsLikeBroad"
              v-model="fuzziness.otherSurname"
              value="4"
              label="Sounds like (broad)"
            ></v-checkbox>
            <v-checkbox 
              class="mt-0 ml-3" 
              type="checkbox" 
              id="otherSurnameFuzzy" 
              v-model="fuzziness.otherSurname" 
              value="8" 
              label="Fuzzy"
            ></v-checkbox>
          </v-row>          
        </v-col>
      </v-row>
      <v-btn type="submit" color="primary">Go</v-btn>
    </v-form>

    <v-row class="pa-3" v-if="search.searchTotal === 0">
      <p>No results found</p>
    </v-row>
    <v-row v-if="search.searchTotal > 0">
      <p>Showing 1 - {{ search.searchList.length }} of {{ search.searchTotal }}</p>
      <SearchResult v-for="(result, $ix) in search.searchList" :key="$ix" :result="result" />
    </v-row>
  </div>
</template>

<script>
import SearchResult from "../components/SearchResult.vue";
import NProgress from "nprogress";
import { mapState } from "vuex";
import { formFieldMixin } from "../mixins/formFieldMixin";

export default {
  mixins: [formFieldMixin],
  components: {
    SearchResult
  },
  data() {
    return {
      query: {},
      fuzziness: {
        principalGiven: [],
        principalSurname: [],
        fatherGiven: [],
        fatherSurname: [],
        motherGiven: [],
        motherSurname: [],
        spouseGiven: [],
        spouseSurname: [],
        otherGiven: [],
        otherSurname: []
      }
    };
  },
  computed: 
    mapState(["search"]),
    listeners() {
      return {
        ...this.$listeners,
        input: this.updateValue
      };
    },
  methods: {
    go() {
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
  padding:0px;
  font-size:50%;
}
.v-checkbox label {
  font-size:50%;
}
</style>