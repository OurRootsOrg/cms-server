<template>
  <div class="search">
    <h1>Search</h1>
    <form @submit.prevent="go">
      <BaseInput label="Given" v-model="query.given" type="text" placeholder="Given name" class="field">
        <input class="fuzzy" type="checkbox" id="principalGivenAlt" v-model="fuzziness.principalGiven" value="1" />
        <label for="principalGivenAlt">Alternate spellings</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="principalGivenSoundsLikeNarrow"
          v-model="fuzziness.principalGiven"
          value="2"
        />
        <label for="principalGivenSoundsLikeNarrow">Sounds like (narrow)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="principalGivenSoundsLikeBroad"
          v-model="fuzziness.principalGiven"
          value="4"
        />
        <label for="principalGivenSoundsLikeBroad">Sounds like (broad)</label>
        <input class="fuzzy" type="checkbox" id="principalGivenFuzzy" v-model="fuzziness.principalGiven" value="8" />
        <label for="principalGivenFuzzy">Fuzzy</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="principalGivenInitials"
          v-model="fuzziness.principalGiven"
          value="16"
        />
        <label for="principalGivenInitials">Initials</label>
      </BaseInput>
      <BaseInput label="Surname" v-model="query.surname" type="text" placeholder="Surname" class="field">
        <input class="fuzzy" type="checkbox" id="principalSurnameAlt" v-model="fuzziness.principalSurname" value="1" />
        <label for="principalSurnameAlt">Alternate spellings</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="principalSurnameSoundsLikeNarrow"
          v-model="fuzziness.principalSurname"
          value="2"
        />
        <label for="principalSurnameSoundsLikeNarrow">Sounds like (narrow)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="principalSurnameSoundsLikeBroad"
          v-model="fuzziness.principalSurname"
          value="4"
        />
        <label for="principalSurnameSoundsLikeBroad">Sounds like (broad)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="principalSurnameFuzzy"
          v-model="fuzziness.principalSurname"
          value="8"
        />
        <label for="principalSurnameFuzzy">Fuzzy</label>
      </BaseInput>

      <h4>Father</h4>
      <BaseInput label="Given" v-model="query.fatherGiven" type="text" placeholder="Given name" class="field">
        <input class="fuzzy" type="checkbox" id="fatherGivenAlt" v-model="fuzziness.fatherGiven" value="1" />
        <label for="fatherGivenAlt">Alternate spellings</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="fatherGivenSoundsLikeNarrow"
          v-model="fuzziness.fatherGiven"
          value="2"
        />
        <label for="fatherGivenSoundsLikeNarrow">Sounds like (narrow)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="fatherGivenSoundsLikeBroad"
          v-model="fuzziness.fatherGiven"
          value="4"
        />
        <label for="fatherGivenSoundsLikeBroad">Sounds like (broad)</label>
        <input class="fuzzy" type="checkbox" id="fatherGivenFuzzy" v-model="fuzziness.fatherGiven" value="8" />
        <label for="fatherGivenFuzzy">Fuzzy</label>
        <input class="fuzzy" type="checkbox" id="fatherGivenInitials" v-model="fuzziness.fatherGiven" value="16" />
        <label for="fatherGivenInitials">Initials</label>
      </BaseInput>
      <BaseInput label="Surname" v-model="query.fatherSurname" type="text" placeholder="Surname" class="field">
        <input class="fuzzy" type="checkbox" id="fatherSurnameAlt" v-model="fuzziness.fatherSurname" value="1" />
        <label for="fatherSurnameAlt">Alternate spellings</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="fatherSurnameSoundsLikeNarrow"
          v-model="fuzziness.fatherSurname"
          value="2"
        />
        <label for="fatherSurnameSoundsLikeNarrow">Sounds like (narrow)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="fatherSurnameSoundsLikeBroad"
          v-model="fuzziness.fatherSurname"
          value="4"
        />
        <label for="fatherSurnameSoundsLikeBroad">Sounds like (broad)</label>
        <input class="fuzzy" type="checkbox" id="fatherSurnameFuzzy" v-model="fuzziness.fatherSurname" value="8" />
        <label for="fatherSurnameFuzzy">Fuzzy</label>
      </BaseInput>

      <h4>Mother</h4>
      <BaseInput label="Given" v-model="query.motherGiven" type="text" placeholder="Given name" class="field">
        <input class="fuzzy" type="checkbox" id="motherGivenAlt" v-model="fuzziness.motherGiven" value="1" />
        <label for="motherGivenAlt">Alternate spellings</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="motherGivenSoundsLikeNarrow"
          v-model="fuzziness.motherGiven"
          value="2"
        />
        <label for="motherGivenSoundsLikeNarrow">Sounds like (narrow)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="motherGivenSoundsLikeBroad"
          v-model="fuzziness.motherGiven"
          value="4"
        />
        <label for="motherGivenSoundsLikeBroad">Sounds like (broad)</label>
        <input class="fuzzy" type="checkbox" id="motherGivenFuzzy" v-model="fuzziness.motherGiven" value="8" />
        <label for="motherGivenFuzzy">Fuzzy</label>
        <input class="fuzzy" type="checkbox" id="motherGivenInitials" v-model="fuzziness.motherGiven" value="16" />
        <label for="motherGivenInitials">Initials</label>
      </BaseInput>
      <BaseInput label="Surname" v-model="query.motherSurname" type="text" placeholder="Surname" class="field">
        <input class="fuzzy" type="checkbox" id="motherSurnameAlt" v-model="fuzziness.motherSurname" value="1" />
        <label for="motherSurnameAlt">Alternate spellings</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="motherSurnameSoundsLikeNarrow"
          v-model="fuzziness.motherSurname"
          value="2"
        />
        <label for="motherSurnameSoundsLikeNarrow">Sounds like (narrow)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="motherSurnameSoundsLikeBroad"
          v-model="fuzziness.motherSurname"
          value="4"
        />
        <label for="motherSurnameSoundsLikeBroad">Sounds like (broad)</label>
        <input class="fuzzy" type="checkbox" id="motherSurnameFuzzy" v-model="fuzziness.motherSurname" value="8" />
        <label for="motherSurnameFuzzy">Fuzzy</label>
      </BaseInput>

      <h4>Spouse</h4>
      <BaseInput label="Given" v-model="query.spouseGiven" type="text" placeholder="Given name" class="field">
        <input class="fuzzy" type="checkbox" id="spouseGivenAlt" v-model="fuzziness.spouseGiven" value="1" />
        <label for="spouseGivenAlt">Alternate spellings</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="spouseGivenSoundsLikeNarrow"
          v-model="fuzziness.spouseGiven"
          value="2"
        />
        <label for="spouseGivenSoundsLikeNarrow">Sounds like (narrow)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="spouseGivenSoundsLikeBroad"
          v-model="fuzziness.spouseGiven"
          value="4"
        />
        <label for="spouseGivenSoundsLikeBroad">Sounds like (broad)</label>
        <input class="fuzzy" type="checkbox" id="spouseGivenFuzzy" v-model="fuzziness.spouseGiven" value="8" />
        <label for="spouseGivenFuzzy">Fuzzy</label>
        <input class="fuzzy" type="checkbox" id="spouseGivenInitials" v-model="fuzziness.spouseGiven" value="16" />
        <label for="spouseGivenInitials">Initials</label>
      </BaseInput>
      <BaseInput label="Surname" v-model="query.spouseSurname" type="text" placeholder="Surname" class="field">
        <input class="fuzzy" type="checkbox" id="spouseSurnameAlt" v-model="fuzziness.spouseSurname" value="1" />
        <label for="spouseSurnameAlt">Alternate spellings</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="spouseSurnameSoundsLikeNarrow"
          v-model="fuzziness.spouseSurname"
          value="2"
        />
        <label for="spouseSurnameSoundsLikeNarrow">Sounds like (narrow)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="spouseSurnameSoundsLikeBroad"
          v-model="fuzziness.spouseSurname"
          value="4"
        />
        <label for="spouseSurnameSoundsLikeBroad">Sounds like (broad)</label>
        <input class="fuzzy" type="checkbox" id="spouseSurnameFuzzy" v-model="fuzziness.spouseSurname" value="8" />
        <label for="spouseSurnameFuzzy">Fuzzy</label>
      </BaseInput>

      <h4>Other person</h4>
      <BaseInput label="Given" v-model="query.otherGiven" type="text" placeholder="Given name" class="field">
        <input class="fuzzy" type="checkbox" id="otherGivenAlt" v-model="fuzziness.otherGiven" value="1" />
        <label for="otherGivenAlt">Alternate spellings</label>
        <input class="fuzzy" type="checkbox" id="otherGivenSoundsLikeNarrow" v-model="fuzziness.otherGiven" value="2" />
        <label for="otherGivenSoundsLikeNarrow">Sounds like (narrow)</label>
        <input class="fuzzy" type="checkbox" id="otherGivenSoundsLikeBroad" v-model="fuzziness.otherGiven" value="4" />
        <label for="otherGivenSoundsLikeBroad">Sounds like (broad)</label>
        <input class="fuzzy" type="checkbox" id="otherGivenFuzzy" v-model="fuzziness.otherGiven" value="8" />
        <label for="otherGivenFuzzy">Fuzzy</label>
        <input class="fuzzy" type="checkbox" id="otherGivenInitials" v-model="fuzziness.otherGiven" value="16" />
        <label for="otherGivenInitials">Initials</label>
      </BaseInput>
      <BaseInput label="Surname" v-model="query.otherSurname" type="text" placeholder="Surname" class="field">
        <input class="fuzzy" type="checkbox" id="otherSurnameAlt" v-model="fuzziness.otherSurname" value="1" />
        <label for="otherSurnameAlt">Alternate spellings</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="otherSurnameSoundsLikeNarrow"
          v-model="fuzziness.otherSurname"
          value="2"
        />
        <label for="otherSurnameSoundsLikeNarrow">Sounds like (narrow)</label>
        <input
          class="fuzzy"
          type="checkbox"
          id="otherSurnameSoundsLikeBroad"
          v-model="fuzziness.otherSurname"
          value="4"
        />
        <label for="otherSurnameSoundsLikeBroad">Sounds like (broad)</label>
        <input class="fuzzy" type="checkbox" id="otherSurnameFuzzy" v-model="fuzziness.otherSurname" value="8" />
        <label for="otherSurnameFuzzy">Fuzzy</label>
      </BaseInput>

      <BaseButton type="submit" class="submit-button" buttonClass="-fill-gradient -size-small">Go</BaseButton>
    </form>
    <p v-if="search.searchTotal === 0">
      No results found
    </p>
    <div v-if="search.searchTotal > 0">
      <p>Showing 1 - {{ search.searchList.length }} of {{ search.searchTotal }}</p>
      <SearchResult v-for="(result, $ix) in search.searchList" :key="$ix" :result="result" />
    </div>
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
  computed: mapState(["search"]),
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
.field {
  margin: 8px;
}
.fuzzy {
  margin-left: 8px;
}
.submit-button {
  margin-top: 16px;
}
</style>
