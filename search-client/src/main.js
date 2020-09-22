import Vue from "vue";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import "nprogress/nprogress.css";
import vuetify from "./plugins/vuetify";
import VueColumnsResizableVuetify from "vue-columns-resizable-vuetify";

Vue.use(VueColumnsResizableVuetify);

Vue.config.productionTip = false;

const requireComponent = require.context("./components", false, /Base[A-Z]\w+\.(vue|js)$/);

requireComponent.keys().forEach(fileName => {
  const componentConfig = requireComponent(fileName);
  const componentName = fileName.replace(/^\.\/(.*)\.\w+$/, "$1");
  Vue.component(componentName, componentConfig.default || componentConfig);
});

new Vue({
  router,
  store,
  vuetify,
  render: h => h(App)
}).$mount("#app");
