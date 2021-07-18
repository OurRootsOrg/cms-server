import Vue from "vue";
import Vuetify, { colors } from "vuetify/lib";

Vue.use(Vuetify);

export default new Vuetify({
  theme: {
    themes: {
      light: {
        primary: colors.cyan.darken2,
        secondary: colors.cyan.lighten4,
        accent: colors.orange.darken3,
        error: "#fff8f8"
      }
    }
  }
});
