import Server from "@/services/Server.js";

export const state = {
  collectionsList: []
};

export const mutations = {
  COLLECTIONS_ADD(state, collection) {
    state.collectionsList.push(collection);
  },
  COLLECTIONS_SET(state, collections) {
    state.collectionsList = collections;
  }
};

export const actions = {
  collectionsCreate({ commit, dispatch }, collection) {
    return Server.collectionsCreate(collection)
      .then(collection => {
        commit("COLLECTIONS_ADD", collection);
        const notification = {
          type: "success",
          message: "Your collection has been created!"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem creating your collection: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  collectionsGetAll({ commit, dispatch }) {
    return Server.collectionsGetAll()
      .then(response => {
        commit("COLLECTIONS_SET", response.data.collections);
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem fetching collections: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  }
};
