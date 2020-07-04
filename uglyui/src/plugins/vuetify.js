import Vue from "vue";
import Vuetify, { colors } from "vuetify/lib";

Vue.use(Vuetify);

export default new Vuetify({
  theme: {
    themes: {
      light: {
        primary: colors.cyan.darken2,
        secondary: colors.lime.darken3,
        accent: colors.orange.darken3,
        error: "#b71c1c"
      }
    }
  }
});
