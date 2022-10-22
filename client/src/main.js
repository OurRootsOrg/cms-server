import Vue from "vue";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import "nprogress/nprogress.css";
import Vuelidate from "vuelidate";
import vuetify from "./plugins/vuetify";
import VueSanitize from "vue-sanitize";
import VueColumnsResizableVuetify from "vue-columns-resizable-vuetify";
import VueCookies from "vue-cookies";

Vue.use(Vuelidate);
Vue.use(VueColumnsResizableVuetify);
Vue.use(VueCookies);
let defaultOptions = {
  allowedTags: ["a", "li", "ol", "p", "ul", "b", "br", "em", "i", "small", "strong", "sub", "sup", "u"],
  allowedAttributes: { a: ["href"] }
};
Vue.use(VueSanitize, defaultOptions);

Vue.config.productionTip = false;

const requireComponent = require.context("./components", false, /Base[A-Z]\w+\.(vue|js)$/);

requireComponent.keys().forEach(fileName => {
  const componentConfig = requireComponent(fileName);
  // const componentName = upperFirst(
  //     camelCase(fileName.replace(/^\.\/(.*)\.\w+$/, '$1'))
  // )
  const componentName = fileName.replace(/^\.\/(.*)\.\w+$/, "$1");
  Vue.component(componentName, componentConfig.default || componentConfig);
});

Vue.filter("formatSize", function(size) {
  if (size > 1024 * 1024 * 1024 * 1024) {
    return (size / 1024 / 1024 / 1024 / 1024).toFixed(2) + " TB";
  } else if (size > 1024 * 1024 * 1024) {
    return (size / 1024 / 1024 / 1024).toFixed(2) + " GB";
  } else if (size > 1024 * 1024) {
    return (size / 1024 / 1024).toFixed(2) + " MB";
  } else if (size > 1024) {
    return (size / 1024).toFixed(2) + " KB";
  }
  return size.toString() + " B";
});

new Vue({
  router,
  store,
  vuetify,
  render: h => h(App)
}).$mount("#app");
