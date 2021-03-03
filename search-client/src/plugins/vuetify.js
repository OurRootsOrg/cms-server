import Vue from "vue";
import Vuetify, { colors } from "vuetify/lib";

Vue.use(Vuetify);

export default new Vuetify({
  theme: {
    themes: {
      light: {
        primary: "#607D8B",
        secondary: "#455A64",
        accent: colors.orange.darken3,
        error: "#b71c1c"
      }
    }
  }
});
