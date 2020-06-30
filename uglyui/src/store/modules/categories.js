import Server from "@/services/Server.js";

export const state = {
  categoriesList: [],
  category: {}
};

export const mutations = {
  CATEGORIES_ADD(state, category) {
    state.categoriesList.push(category);
  },
  CATEGORIES_SET(state, categories) {
    state.categoriesList = categories;
  },
  CATEGORIES_REMOVE(state, id) {
    state.categoriesList = state.categoriesList.filter(cat => cat.id !== id);
  },
  CATEGORY_SET(state, cat) {
    state.category = cat;
  },
  CATEGORY_UPDATE(state, cat) {
    if (state.category && state.category.id === cat.id) {
      state.category = cat;
    }
    if (state.categoriesList) {
      for (let i = 0; i < state.categoriesList.length; i++) {
        if (state.categoriesList[i].id === cat.id) {
          Object.assign(state.categoriesList[i], cat);
        }
      }
    }
  }
};

export const actions = {
  categoriesCreate({ commit, dispatch }, category) {
    return Server.categoriesCreate(category)
      .then(response => {
        commit("CATEGORIES_ADD", response.data);
        const notification = {
          type: "success",
          message: "Your category has been created"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem creating your category: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  categoriesUpdate({ commit, dispatch }, coll) {
    return Server.categoriesUpdate(coll)
      .then(response => {
        commit("CATEGORY_UPDATE", response.data);
        const notification = {
          type: "success",
          message: "Your category has been updated"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem updating your category: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  },
  categoriesGetAll({ commit, dispatch }) {
    return Server.categoriesGetAll()
      .then(response => {
        commit("CATEGORIES_SET", response.data.categories);
        return response.data.categories;
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem fetching categories: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  },
  categoriesGetOne({ commit }, id) {
    return Server.categoriesGetOne(id).then(response => {
      commit("CATEGORY_SET", response.data);
      return response.data;
    });
  },
  categoriesDelete({ commit, dispatch }, id) {
    return Server.categoriesDelete(id)
      .then(() => {
        commit("CATEGORIES_REMOVE", id);
        const notification = {
          type: "success",
          message: "Your category has been deleted"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem deleting the category: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  }
};
