import Server from "@/services/Server.js";

export const state = {
  categoriesList: []
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
  }
};

export const actions = {
  categoriesCreate({ commit, dispatch }, category) {
    return Server.categoriesCreate(category)
      .then(category => {
        commit("CATEGORIES_ADD", category);
        const notification = {
          type: "success",
          message: "Your category has been created!"
        };
        dispatch("notificationsAdd", notification, { root: true });
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
  categoriesGetAll({ commit, dispatch }) {
    return Server.categoriesGetAll()
      .then(response => {
        commit("CATEGORIES_SET", response.data.categories);
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem fetching categories: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  },
  categoriesDelete({ commit, dispatch }, id) {
    return Server.categoriesDelete(id)
      .then(() => {
        commit("CATEGORIES_REMOVE", id);
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
