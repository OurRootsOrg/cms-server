<template>
  <!--if the results move to a new page, take out the wrapper row as it won't be needed to provide the single template element anymore-->
  <v-row>
    <v-col cols="12" md="9" class="search">
      <h1>Search</h1>
      <v-form @submit.prevent="go">
        <!--name row-->
        <v-row no-gutters>
          <v-col cols="12" md="6" class="pr-3 mt-3">
            <h4>First &amp; Middle Name(s)</h4>
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
              <v-menu offset-x :close-on-content-click="false" v-model="givenOptionsMenu">
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
          <v-col cols="12" md="6" class="mt-3">
            <h4>Last Name</h4>
            <v-row no-gutters>
              <v-text-field dense outlined v-model="query.surname" type="text" placeholder="Surname"></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.surname">
              <v-menu offset-x :close-on-content-click="false" v-model="surnameOptionsMenu">
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
        <v-row no-gutters>
          <v-col cols="12" md="6" class="pr-3 mt-3">
            <h4>Place your ancestor might have lived</h4>
            <v-row no-gutters>
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
                @change="anyPlaceChanged()"
              ></v-autocomplete>
            </v-row>
            <v-row no-gutters v-if="query.anyPlace">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="query.anyPlaceFuzziness"
                  :value="1"
                  class="mt-0 mr-1 ml-n1 primary--text shrink smallCheckbox"
                >
                </v-checkbox>
                <span class="mt-2 ml-n3 primary--text">Exact location</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col cols="6" md="2" class="mt-3">
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
              <v-menu offset-x :close-on-content-click="false" v-model="birthOptionsMenu">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn color="primary" text x-small v-bind="attrs" v-on="on" class="pa-0 mt-0">
                    <v-icon v-if="query.birthDateFuzziness === 0" small class="mr-1">mdi-checkbox-blank-outline</v-icon>
                    <v-icon v-if="query.birthDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                    <span class="ml-1">{{
                      query.birthDateFuzziness === 0
                        ? "Exactness"
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
        <v-row no-gutters class="my-5">
          <strong>Add event details:</strong>
          <v-btn text color="primary" class="eventButton" :disabled="showEvent.birth" @click="showEvent.birth = true"
            >Birth</v-btn
          >
          <v-btn
            text
            color="primary"
            class="eventButton"
            :disabled="showEvent.marriage"
            @click="showEvent.marriage = true"
            >Marriage</v-btn
          >
          <v-btn text color="primary" class="eventButton" :disabled="showEvent.death" @click="showEvent.death = true"
            >Death</v-btn
          >
          <v-btn
            text
            color="primary"
            class="eventButton"
            :disabled="showEvent.residence"
            @click="showEvent.residence = true"
            >Lived In</v-btn
          >
          <v-btn text color="primary" class="eventButton" :disabled="showEvent.any" @click="showEvent.any = true"
            >Any Event</v-btn
          >
        </v-row>
        <!--Birth-->
        <v-row no-gutters class="my-3" v-if="showEvent.birth">
          <v-col cols="2" class="pl-5 pt-3">
            <h4>Birth</h4>
          </v-col>
          <v-col cols="12" md="3">
            <v-row no-gutters class="pr-3">
              <v-text-field
                outlined
                dense
                v-model="query.birthDate"
                type="text"
                placeholder="Birth date"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.birthDate">
              <v-menu offset-x :close-on-content-click="false" v-model="birthOptionsMenu2">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn color="primary" text x-small v-bind="attrs" v-on="on" class="pa-0 mt-0">
                    <v-icon v-if="query.birthDateFuzziness === 0" small class="mr-1">mdi-checkbox-blank-outline</v-icon>
                    <v-icon v-if="query.birthDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                    <span class="ml-1">{{
                      query.birthDateFuzziness === 0
                        ? "Exactness"
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
          <v-col cols="12" md="6">
            <v-row no-gutters>
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
              ></v-autocomplete>
            </v-row>
            <v-row no-gutters class="mt-0" v-if="query.birthPlace">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="query.birthPlaceFuzziness"
                  :value="1"
                  class="shrink mr-2 mt-0 smallCheckbox"
                  dense
                  primary
                >
                </v-checkbox
                ><span class="mt-2 ml-n3 primary--text">Exact location</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col cols="1" class="ma-0">
            <v-btn text @click="clearEvent('birth')" class="grey--text mt-0"
              ><v-icon>mdi-close-circle-outline</v-icon></v-btn
            >
          </v-col>
        </v-row>
        <!--Marriage-->
        <v-row no-gutters class="my-3" v-if="showEvent.marriage">
          <v-col cols="2" class="pl-5 pt-3">
            <h4>Marriage</h4>
          </v-col>
          <v-col cols="12" md="3">
            <v-row no-gutters class="pr-3">
              <v-text-field
                outlined
                dense
                v-model="query.marriageDate"
                type="text"
                placeholder="Marriage date"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.marriageDate">
              <v-menu offset-x :close-on-content-click="false" v-model="marriageOptionsMenu">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn color="primary" text x-small v-bind="attrs" v-on="on" class="pa-0 mt-0">
                    <v-icon v-if="query.marriageDateFuzziness === 0" small class="mr-1"
                      >mdi-checkbox-blank-outline</v-icon
                    >
                    <v-icon v-if="query.marriageDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                    <span class="ml-1">{{
                      query.marriageDateFuzziness === 0
                        ? "Exactness"
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
            </v-row>
          </v-col>
          <v-col cols="12" md="6">
            <v-row no-gutters>
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
              ></v-autocomplete>
            </v-row>
            <v-row no-gutters class="mt-0 exactCheck" v-if="query.marriagePlace">
              <v-checkbox
                v-model="query.marriagePlaceFuzziness"
                :value="1"
                class="shrink mr-2 mt-0 smallCheckbox"
                dense
                primary
              >
              </v-checkbox
              ><span class="mt-2 ml-n3 primary--text">Exact location</span>
            </v-row>
          </v-col>
          <v-col cols="1" class="ma-0">
            <v-btn text @click="clearEvent('marriage')" class="grey--text mt-0"
              ><v-icon>mdi-close-circle-outline</v-icon></v-btn
            >
          </v-col>
        </v-row>
        <!--Death-->
        <v-row no-gutters class="my-3" v-if="showEvent.death">
          <v-col cols="2" class="pl-5 pt-3">
            <h4>Death</h4>
          </v-col>
          <v-col cols="12" md="3">
            <v-row no-gutters class="pr-3">
              <v-text-field
                outlined
                dense
                v-model="query.deathDate"
                type="text"
                placeholder="Death date"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.deathDate">
              <v-menu offset-x :close-on-content-click="false" v-model="deathOptionsMenu">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn color="primary" text x-small v-bind="attrs" v-on="on" class="pa-0 mt-0">
                    <v-icon v-if="query.deathDateFuzziness === 0" small class="mr-1">mdi-checkbox-blank-outline</v-icon>
                    <v-icon v-if="query.deathDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                    <span class="ml-1">{{
                      query.deathDateFuzziness === 0
                        ? "Exactness"
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
            </v-row>
          </v-col>
          <v-col cols="12" md="6">
            <v-row no-gutters>
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
              ></v-autocomplete>
            </v-row>
            <v-row no-gutters class="mt-0 exactCheck" v-if="query.deathPlace">
              <v-checkbox
                v-model="query.deathPlaceFuzziness"
                :value="1"
                class="shrink mr-2 mt-0 smallCheckbox"
                dense
                primary
              >
              </v-checkbox
              ><span class="mt-2 ml-n3 primary--text">Exact location</span>
            </v-row>
          </v-col>
          <v-col cols="1" class="ma-0">
            <v-btn text @click="clearEvent('death')" class="grey--text mt-0"
              ><v-icon>mdi-close-circle-outline</v-icon></v-btn
            >
          </v-col>
        </v-row>
        <!--Residence-->
        <v-row no-gutters class="my-3" v-if="showEvent.residence">
          <v-col cols="2" class="pl-5 pt-3">
            <h4>Lived in</h4>
          </v-col>
          <v-col cols="12" md="3">
            <v-row no-gutters class="pr-3">
              <v-text-field
                outlined
                dense
                v-model="query.residenceDate"
                type="text"
                placeholder="Residence date"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.residenceDate">
              <v-menu offset-x :close-on-content-click="false" v-model="residenceOptionsMenu">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn color="primary" text x-small v-bind="attrs" v-on="on" class="pa-0 mt-0">
                    <v-icon v-if="query.residenceDateFuzziness === 0" small class="mr-1"
                      >mdi-checkbox-blank-outline</v-icon
                    >
                    <v-icon v-if="query.residenceDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                    <span class="ml-1">{{
                      query.residenceDateFuzziness === 0
                        ? "Exactness"
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
            </v-row>
          </v-col>
          <v-col cols="12" md="6">
            <v-row no-gutters>
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
              ></v-autocomplete>
            </v-row>
            <v-row no-gutters class="mt-0 exactCheck" v-if="query.residencePlace">
              <v-checkbox
                v-model="query.residencePlaceFuzziness"
                :value="1"
                class="shrink mr-2 mt-0 smallCheckbox"
                dense
                primary
              >
              </v-checkbox
              ><span class="mt-2 ml-n3 primary--text">Exact location</span>
            </v-row>
          </v-col>
          <v-col cols="1" class="ma-0">
            <v-btn text @click="clearEvent('residence')" class="grey--text mt-0"
              ><v-icon>mdi-close-circle-outline</v-icon></v-btn
            >
          </v-col>
        </v-row>
        <!--Any-->
        <v-row no-gutters class="my-3" v-if="showEvent.any">
          <v-col cols="2" class="pl-5 pt-3">
            <h4>Any event</h4>
          </v-col>
          <v-col cols="12" md="3">
            <v-row no-gutters class="pr-3">
              <v-text-field
                outlined
                dense
                v-model="query.anyDate"
                type="text"
                placeholder="Any event date"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.anyDate">
              <v-menu offset-x :close-on-content-click="false" v-model="anyOptionsMenu">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn color="primary" text x-small v-bind="attrs" v-on="on" class="pa-0 mt-0">
                    <v-icon v-if="query.anyDateFuzziness === 0" small class="mr-1">mdi-checkbox-blank-outline</v-icon>
                    <v-icon v-if="query.anyDateFuzziness > 0" small class="mr-1">mdi-checkbox-marked</v-icon>
                    <span class="ml-1">{{
                      query.anyDateFuzziness === 0
                        ? "Exactness"
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
            </v-row>
          </v-col>
          <v-col cols="12" md="6">
            <v-row no-gutters>
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
                placeholder="Any event place"
              ></v-autocomplete>
            </v-row>
            <v-row no-gutters class="mt-0 exactCheck" v-if="query.anyPlace">
              <v-checkbox
                v-bind="attrs"
                v-on="on"
                v-model="query.anyPlaceFuzziness"
                :value="1"
                class="shrink mr-2 mt-0 smallCheckbox"
                dense
                primary
              >
              </v-checkbox
              ><span class="mt-2 ml-n3 primary--text">Exact location</span>
            </v-row>
          </v-col>
          <v-col cols="1" class="ma-0">
            <v-btn text @click="clearEvent('any')" class="grey--text mt-0"
              ><v-icon>mdi-close-circle-outline</v-icon></v-btn
            >
          </v-col>
        </v-row>
        <!--Relative buttons-->
        <v-row no-gutters class="mt-5">
          <strong>Add family member:</strong>
          <v-btn
            text
            color="primary"
            class="eventButton"
            :disabled="showRelative.father"
            @click="showRelative.father = true"
            >Father</v-btn
          >
          <v-btn
            text
            color="primary"
            class="eventButton"
            :disabled="showRelative.mother"
            @click="showRelative.mother = true"
            >Mother</v-btn
          >
          <v-btn
            text
            color="primary"
            class="eventButton"
            :disabled="showRelative.spouse"
            @click="showRelative.spouse = true"
            >Spouse</v-btn
          >
          <v-btn
            text
            color="primary"
            class="eventButton"
            :disabled="showRelative.other"
            @click="showRelative.other = true"
            >Other</v-btn
          >
        </v-row>
        <!--Father-->
        <v-row no-gutters class="my-5" v-if="showRelative.father">
          <v-col cols="2" class="pl-5 pt-3">
            <h4 class="mb-1">Father</h4>
          </v-col>
          <v-col cols="12" md="3">
            <v-row no-gutters class="pr-3">
              <v-text-field
                outlined
                dense
                v-model="query.fatherGiven"
                type="text"
                placeholder="Father's given name"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.fatherGiven">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="fuzziness.fatherGiven"
                  :value="1"
                  class="shrink mr-2 mt-0 smallCheckbox"
                  dense
                  primary
                >
                </v-checkbox
                ><span class="mt-2 ml-n3 primary--text">Exact spelling</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col>
            <v-row no-gutters class="pa-0 ma-0">
              <v-text-field
                outlined
                dense
                v-model="query.fatherSurname"
                type="text"
                placeholder="Father's surname"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.fatherSurname">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="fuzziness.fatherSurname"
                  :value="1"
                  class="shrink mr-2 mt-0 smallCheckbox"
                  dense
                  primary
                >
                </v-checkbox
                ><span class="mt-2 ml-n3 primary--text">Exact spelling</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col cols="1">
            <v-btn text @click="clearRelative('father')" class="grey--text mt-0"
              ><v-icon>mdi-close-circle-outline</v-icon></v-btn
            >
          </v-col>
        </v-row>
        <!--Mother-->
        <v-row no-gutters class="my-5" v-if="showRelative.mother">
          <v-col cols="2" class="pl-5 pt-3">
            <h4 class="mb-1">Mother</h4>
          </v-col>
          <v-col cols="12" md="3">
            <v-row no-gutters class="pr-3">
              <v-text-field
                outlined
                dense
                v-model="query.motherGiven"
                type="text"
                placeholder="Mother's given name"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.motherGiven">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="fuzziness.motherGiven"
                  :value="1"
                  class="shrink mr-2 mt-0 smallCheckbox"
                  dense
                  primary
                >
                </v-checkbox
                ><span class="mt-2 ml-n3 primary--text">Exact spelling</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col>
            <v-row no-gutters class="pa-0 ma-0">
              <v-text-field
                outlined
                dense
                v-model="query.motherSurname"
                type="text"
                placeholder="Mother's surname"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.motherSurname">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="fuzziness.motherSurname"
                  :value="1"
                  class="shrink mr-2 mt-0 smallCheckbox"
                  dense
                  primary
                >
                </v-checkbox
                ><span class="mt-2 ml-n3 primary--text">Exact spelling</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col cols="1">
            <v-btn text @click="clearRelative('mother')" class="grey--text mt-0"
              ><v-icon>mdi-close-circle-outline</v-icon></v-btn
            >
          </v-col>
        </v-row>
        <!--Spouse-->
        <v-row no-gutters class="my-5" v-if="showRelative.spouse">
          <v-col cols="2" class="pl-5 pt-3">
            <h4 class="mb-1">Spouse</h4>
          </v-col>
          <v-col cols="12" md="3">
            <v-row no-gutters class="pr-3">
              <v-text-field
                outlined
                dense
                v-model="query.spouseGiven"
                type="text"
                placeholder="Spouse's given name"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.spouseGiven">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="fuzziness.spouseGiven"
                  :value="1"
                  class="shrink mr-2 mt-0 smallCheckbox"
                  dense
                  primary
                >
                </v-checkbox
                ><span class="mt-2 ml-n3 primary--text">Exact spelling</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col>
            <v-row no-gutters class="pa-0 ma-0">
              <v-text-field
                outlined
                dense
                v-model="query.spouseSurname"
                type="text"
                placeholder="Spouse's surname"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.spouseSurname">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="fuzziness.spouseSurname"
                  :value="1"
                  class="shrink mr-2 mt-0 smallCheckbox"
                  dense
                  primary
                >
                </v-checkbox
                ><span class="mt-2 ml-n3 primary--text">Exact spelling</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col cols="1">
            <v-btn text @click="clearRelative('spouse')" class="grey--text mt-0"
              ><v-icon>mdi-close-circle-outline</v-icon></v-btn
            >
          </v-col>
        </v-row>
        <!--Other-->
        <v-row no-gutters class="my-5" v-if="showRelative.other">
          <v-col cols="2" class="pl-5 pt-3">
            <h4 class="mb-1">Other person</h4>
          </v-col>
          <v-col cols="12" md="3">
            <v-row no-gutters class="pr-3">
              <v-text-field
                outlined
                dense
                v-model="query.otherGiven"
                type="text"
                placeholder="Other person's given name"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.otherGiven">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="fuzziness.otherGiven"
                  :value="1"
                  class="shrink mr-2 mt-0 smallCheckbox"
                  dense
                  primary
                >
                </v-checkbox
                ><span class="mt-2 ml-n3 primary--text">Exact spelling</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col>
            <v-row no-gutters class="pa-0 ma-0">
              <v-text-field
                outlined
                dense
                v-model="query.otherSurname"
                type="text"
                placeholder="Other person's surname"
                class="ma-0 mb-n2"
              ></v-text-field>
            </v-row>
            <v-row no-gutters class="mt-n5" v-if="query.otherSurname">
              <v-col cols="12" class="exactCheck d-flex flex-row">
                <v-checkbox
                  v-model="fuzziness.otherSurname"
                  :value="1"
                  class="shrink mr-2 mt-0 smallCheckbox"
                  dense
                  primary
                >
                </v-checkbox
                ><span class="mt-2 ml-n3 primary--text">Exact spelling</span>
              </v-col>
            </v-row>
          </v-col>
          <v-col cols="1">
            <v-btn text @click="clearRelative('other')" class="grey--text mt-0"
              ><v-icon>mdi-close-circle-outline</v-icon></v-btn
            >
          </v-col>
        </v-row>
        <!--Keywords-->
        <v-row no-gutters class="mt-5">
          <h4 class="mt-2 mr-2">Keyword:</h4>
          <v-text-field
            outlined
            dense
            v-model="query.keywords"
            type="text"
            placeholder="Occupation, etc."
            class="ma-0 mb-n2"
          ></v-text-field>
        </v-row>
        <v-btn class="mt-2 mb-4" type="submit" color="primary">Go</v-btn>
      </v-form>
    </v-col>

    <!--search results page-->
    <v-col cols="12">
      <v-row v-if="searchPerformed && search.searchTotal === 0">
        <p>No results found</p>
      </v-row>

      <v-row no-gutters v-if="searchPerformed && search.searchTotal > 0">
        <v-col cols="12" align="center" class="my-5">
          <h1>All search results</h1>
        </v-col>
        <!--search results facets-->
        <v-col cols="12" md="3" class="no-underline">
          <h3 class="mb-3">Filter by</h3>
          <v-row no-gutters>
            <h4>All Categories</h4>
            <v-col cols="12" class="pa-0 ma-0">
              <v-row v-if="query.category" no-gutters>
                <v-col cols="1">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('category', null) }"
                    x-small
                    icon
                    class="grey--text pr-2 mt-n1"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                </v-col>
                <v-col cols="11">
                  <span>{{ query.category }}</span>
                </v-col>
              </v-row>
              <v-row v-if="query.collection" no-gutters>
                <v-col cols="1" offset-md="1">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('collection', null) }"
                    x-small
                    icon
                    class="grey--text pr-2 mt-n1"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                </v-col>
                <v-col cols="10">
                  <span>{{ query.collection }}</span>
                </v-col>
              </v-row>
              <v-row v-if="categoryFacet" no-gutters>
                <v-col v-for="(bucket, $ix) in categoryFacet.buckets" :key="$ix" cols="12">
                  <v-row no-gutters>
                    <v-col cols="1">
                      <v-btn
                        :to="{ name: 'search', query: getQuery(categoryFacet.key, bucket.label) }"
                        x-small
                        icon
                        class="grey--text pr-2 mt-n1"
                      >
                        <v-icon>mdi-chevron-right</v-icon>
                      </v-btn>
                    </v-col>
                    <v-col cols="9">
                      <router-link :to="{ name: 'search', query: getQuery(categoryFacet.key, bucket.label) }">{{
                        bucket.label
                      }}</router-link>
                    </v-col>
                    <v-col cols="2">
                      {{ bucket.count }}
                    </v-col>
                  </v-row>
                </v-col>
              </v-row>
            </v-col>
          </v-row>
          <v-row no-gutters class="mt-5">
            <h4>Collection Location</h4>
            <v-col cols="12" class="pa-0 ma-0">
              <v-row v-if="query.collectionPlace1" no-gutters>
                <v-col cols="1">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('collectionPlace1', null) }"
                    x-small
                    icon
                    class="grey--text pr-2 mt-n1"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                </v-col>
                <v-col cols="11">
                  <span>{{ query.collectionPlace1 }}</span>
                </v-col>
              </v-row>
              <v-row v-if="query.collectionPlace2" no-gutters>
                <v-col cols="1">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('collectionPlace2', null) }"
                    x-small
                    icon
                    class="grey--text pr-2 mt-n1"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                </v-col>
                <v-col cols="11">
                  <span>{{ query.collectionPlace2 }}</span>
                </v-col>
              </v-row>
              <v-row v-if="query.collectionPlace3" no-gutters>
                <v-col cols="1">
                  <v-btn
                    :to="{ name: 'search', query: getQuery('collectionPlace3', null) }"
                    x-small
                    icon
                    class="grey--text pr-2 mt-n1"
                  >
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                </v-col>
                <v-col cols="11">
                  <span>{{ query.collectionPlace3 }}</span> &nbsp;
                </v-col>
              </v-row>
              <v-row v-if="placeFacet" no-gutters>
                <v-col v-for="(bucket, $ix) in placeFacet.buckets" :key="$ix" cols="12">
                  <v-row no-gutters>
                    <v-col cols="1">
                      <v-btn
                        :to="{ name: 'search', query: getQuery(placeFacet.key, bucket.label) }"
                        x-small
                        icon
                        class="grey--text pr-2 mt-n1"
                      >
                        <v-icon>mdi-chevron-right</v-icon>
                      </v-btn>
                    </v-col>
                    <v-col cols="9">
                      <router-link :to="{ name: 'search', query: getQuery(placeFacet.key, bucket.label) }">{{
                        bucket.label
                      }}</router-link>
                    </v-col>
                    <v-col cols="2">
                      {{ bucket.count }}
                    </v-col>
                  </v-row>
                </v-col>
              </v-row>
            </v-col>
          </v-row>
        </v-col>
        <!--search results-->
        <v-col cols="12" md="9">
          <v-row v-if="searchPerformed && search.searchTotal > 0" no-gutters class="pl-3 pb-3">
            Showing results {{ query.from + 1 }} - {{ query.from + search.searchList.length }} of
            {{ search.searchTotal }}
          </v-row>
          <v-card>
            <v-row no-gutters class="no-underline pl-3 resultsHeader">
              <v-col cols="12" md="4">
                Name
              </v-col>
              <v-col cols="12" md="3">
                Events
              </v-col>
              <v-col cols="12" md="4">
                Relationship
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
  </v-row>
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
        dlg: [],
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
        { value: 2, text: "+/- 1 year" },
        { value: 3, text: "+/- 2 years" },
        { value: 4, text: "+/- 5 years" },
        { value: 5, text: "+/- 10 years" }
      ],
      givenFuzzinessLevels: [
        { value: 0, text: "Default" },
        { value: 1, text: "Exact spelling only" },
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
      anyPlaceLoading: false,
      //option menus
      givenOptionsMenu: false,
      surnameOptionsMenu: false,
      placeOptionsMenu: false,
      birthOptionsMenu: false,
      birthOptionsMenu2: false
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
      this.placeOptionsMenu = false;
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
.smallCheckbox i {
  font-size: 8px !important;
  color: #0097a7 !important;
}

/* .result:nth-child(odd) {
  background-color: #f7f7f7;
}
.result:nth-child(even) {
  background-color: #ffffff;
} */
</style>
