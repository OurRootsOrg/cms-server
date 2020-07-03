import Vue from "vue";
import Vuetify, { colors } from "vuetify/lib";

Vue.use(Vuetify);

export default new Vuetify({
    theme: {
        themes: {
          light: {
            primary: colors.cyan.darken3,
            secondary: colors.orange.darken3,
            accent: colors.lime.darken3,
            error: '#b71c1c',
          },
        },
      },
});
