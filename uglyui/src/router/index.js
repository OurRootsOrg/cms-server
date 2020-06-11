import Vue from "vue";
import VueRouter from "vue-router";
import NProgress from "nprogress";
import Home from "../views/Home.vue";
import NotFound from "../views/NotFound.vue";
import NetworkIssue from "../views/NetworkIssue.vue";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    name: "Home",
    component: Home
  },
  {
    path: "/dashboard",
    name: "dashboard",
    component: () => import(/* webpackChunkName: "about" */ "../views/Dashboard.vue")
  },
  {
    path: "/categories",
    name: "categories-list",
    component: () => import(/* webpackChunkName: "about" */ "../views/CategoriesList.vue")
  },
  {
    path: "/categories/create",
    name: "categories-create",
    component: () => import(/* webpackChunkName: "about" */ "../views/CategoriesCreate.vue")
  },
  {
    path: "/collections",
    name: "collections-list",
    component: () => import(/* webpackChunkName: "about" */ "../views/CollectionsList.vue")
  },
  {
    path: "/collections/create",
    name: "collections-create",
    component: () => import(/* webpackChunkName: "about" */ "../views/CollectionsCreateEdit.vue")
  },
  {
    path: "/collections/:cid",
    name: "collection-edit",
    component: () => import(/* webpackChunkName: "about" */ "../views/CollectionsCreateEdit.vue")
  },
  {
    path: "/posts",
    name: "posts-list",
    component: () => import(/* webpackChunkName: "about" */ "../views/PostsList.vue")
  },
  {
    path: "/posts/create",
    name: "posts-create",
    component: () => import(/* webpackChunkName: "about" */ "../views/PostsCreate.vue")
  },
  {
    path: "/posts/:pid",
    name: "post-show",
    component: () => import(/* webpackChunkName: "about" */ "../views/PostShow.vue")
  },
  {
    path: "/users",
    name: "users-list",
    component: () => import(/* webpackChunkName: "about" */ "../views/UsersList.vue")
  },
  {
    path: "/search",
    name: "search",
    component: () => import(/* webpackChunkName: "about" */ "../views/Search.vue")
  },
  {
    path: "/settings",
    name: "settings",
    component: () => import(/* webpackChunkName: "about" */ "../views/Settings.vue")
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
    redirect: { name: "404", params: { resource: "page" } }
  }
];

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes
});

router.beforeEach((routeTo, routeFrom, next) => {
  NProgress.start();
  next();
});

router.afterEach(() => {
  NProgress.done();
});

export default router;
