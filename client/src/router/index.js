import Vue from "vue";
import VueRouter from "vue-router";
import store from "@/store";
import Auth from "@/services/Auth";
import NProgress from "nprogress";
import Home from "../views/Home.vue";
import Society from "../views/Society.vue";
import NotFound from "../views/NotFound.vue";
import NetworkIssue from "../views/NetworkIssue.vue";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    name: "home",
    component: Home
  },
  {
    path: "/invitation/:code",
    name: "accept-invitation",
    meta: { requiresAuth: false },
    component: () => import(/* webpackChunkName: "about" */ "../views/AcceptInvitation.vue")
  },
  {
    path: "/society/:society",
    component: Society,
    children: [
      {
        path: "home",
        name: "society-home",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/SocietyHome.vue")
      },
      {
        path: "categories",
        name: "categories-list",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/CategoriesList.vue")
      },
      {
        path: "categories/create",
        name: "categories-create",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/CategoriesCreateEdit.vue")
      },
      {
        path: "categories/:cid",
        name: "category-edit",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/CategoriesCreateEdit.vue")
      },
      {
        path: "collections",
        name: "collections-list",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/CollectionsList.vue")
      },
      {
        path: "collections/create",
        name: "collections-create",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/CollectionsCreateEdit.vue")
      },
      {
        path: "collections/:cid",
        name: "collection-edit",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/CollectionsCreateEdit.vue")
      },
      {
        path: "images/:pid/:path",
        name: "image",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/Image.vue")
      },
      {
        path: "record-sets",
        name: "posts-list",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/PostsList.vue")
      },
      {
        path: "downloads/:key",
        name: "downloads",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/Downloads.vue")
      },
      {
        path: "record-sets/create",
        name: "posts-create",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/PostsCreateEdit.vue")
      },
      {
        path: "record-sets/:pid",
        name: "post-edit",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/PostsCreateEdit.vue")
      },
      {
        path: "records/:rid",
        name: "records-view",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/RecordsView.vue")
      },
      {
        path: "settings",
        name: "settings",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/Settings.vue")
      },
      {
        path: "users",
        name: "users-list",
        meta: { requiresAuth: true },
        component: () => import(/* webpackChunkName: "about" */ "../views/UsersList.vue")
      }
    ]
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

router.beforeEach(async (routeTo, routeFrom, next) => {
  NProgress.start();
  await Auth.isLoaded();
  if (!store.getters.userIsLoggedIn && routeTo.matched.some(record => record.meta.requiresAuth)) {
    next("/");
  } else {
    next();
  }
});

router.afterEach(() => {
  NProgress.done();
});

export default router;
