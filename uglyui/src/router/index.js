import Vue from "vue";
import VueRouter from "vue-router";
import NProgress from "nprogress";
import Home from "../views/Home.vue";
import NotFound from "../views/NotFound.vue";
import NetworkIssue from "../views/NetworkIssue.vue";
import Server from "@/services/Server.js";
import { getAuth } from "../plugins/auth0";

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
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/Dashboard.vue")
  },
  {
    path: "/categories",
    name: "categories-list",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/CategoriesList.vue")
  },
  {
    path: "/categories/create",
    name: "categories-create",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/CategoriesCreateEdit.vue")
  },
  {
    path: "/categories/:cid",
    name: "category-edit",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/CategoriesCreateEdit.vue")
  },
  {
    path: "/collections",
    name: "collections-list",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/CollectionsList.vue")
  },
  {
    path: "/collections/create",
    name: "collections-create",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/CollectionsCreateEdit.vue")
  },
  {
    path: "/collections/:cid",
    name: "collection-edit",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/CollectionsCreateEdit.vue")
  },
  {
    path: "/posts",
    name: "posts-list",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/PostsList.vue")
  },
  {
    path: "/posts/create",
    name: "posts-create",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/PostsCreateEdit.vue")
  },
  {
    path: "/posts/:pid",
    name: "post-edit",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/PostsCreateEdit.vue")
  },
  {
    path: "/users",
    name: "users-list",
    meta: { requiresAuth: true },
    component: () => import(/* webpackChunkName: "about" */ "../views/UsersList.vue")
  },
  {
    path: "/search",
    name: "search",
    component: () => import(/* webpackChunkName: "about" */ "../views/Search.vue")
  },
  {
    path: "/search/:rid",
    name: "search-detail",
    component: () => import(/* webpackChunkName: "about" */ "../views/SearchDetail.vue")
  },
  {
    path: "/settings",
    name: "settings",
    meta: { requiresAuth: true },
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

async function checkLogin() {
  if (Server.isLoggedIn()) {
    return true;
  }
  let auth0 = await getAuth();
  if (!auth0.isAuthenticated) {
    return false;
  }
  let token = await auth0.getTokenSilently();
  Server.login(token);
  return true;
}

router.beforeEach((routeTo, routeFrom, next) => {
  NProgress.start();
  checkLogin().then(isLoggedIn => {
    if (!isLoggedIn && routeTo.matched.some(record => record.meta.requiresAuth)) {
      next("/");
    } else {
      next();
    }
  });
});

router.afterEach(() => {
  NProgress.done();
});

export default router;
