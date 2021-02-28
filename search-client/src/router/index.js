import Vue from "vue";
import VueRouter from "vue-router";
import NProgress from "nprogress";
import Image from "../views/Image.vue";
import Search from "../views/Search.vue";
import SearchDetail from "../views/SearchDetail.vue";
import NotFound from "../views/NotFound.vue";
import NetworkIssue from "../views/NetworkIssue.vue";

Vue.use(VueRouter);

const routes = [
  {
    path: "/search",
    name: "search",
    component: Search
  },
  {
    path: "/images/:societyId/:pid/:path",
    name: "image",
    component: Image
  },
  {
    path: "/search/:rid",
    name: "search-detail",
    component: SearchDetail
  },
  {
    path: "/404",
    name: "404",
    component: NotFound,
    props: true
  },
  {
    path: "/network-issue",
    name: "network-issue",
    component: NetworkIssue
  },
  {
    path: "*",
    redirect: { name: "search" }
  }
];

const router = new VueRouter({
  mode: "hash",
  base: process.env.BASE_URL,
  routes
});

router.beforeEach(async (routeTo, routeFrom, next) => {
  NProgress.start();
  next();
});

router.afterEach(() => {
  NProgress.done();
});

export default router;
