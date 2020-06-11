import Server from "@/services/Server.js";

export const state = {
  collectionsList: [],
  collection: {}
};

export const mutations = {
  COLLECTIONS_ADD(state, collection) {
    state.collectionsList.push(collection);
  },
  COLLECTIONS_SET(state, collections) {
    state.collectionsList = collections;
  },
  COLLECTION_SET(state, coll) {
    if (!coll.fields) coll.fields = []; // force empty array
    state.collection = coll;
  },
  COLLECTION_UPDATE(state, coll) {
    if (state.collection && state.collection.id === coll.id) {
      state.collection = coll;
    }
    if (state.collectionsList) {
      for (let i = 0; i < state.collectionsList.length; i++) {
        if (state.collectionsList[i].id === coll.id) {
          state.collectionsList[i] = coll;
        }
      }
    }
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
  collectionsUpdate({ commit, dispatch }, coll) {
    return Server.collectionsUpdate(coll)
      .then(response => {
        commit("COLLECTION_UPDATE", response.data);
        const notification = {
          type: "success",
          message: "Your collection has been updated!"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem updating your collection: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
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
  },
  collectionsGetOne({ commit, getters }, id) {
    let coll = getters.getCollectionById(id);

    if (coll) {
      commit("COLLECTION_SET", coll);
      return coll;
    } else {
      return Server.collectionsGetOne(id).then(response => {
        commit("COLLECTION_SET", response.data);
        return response.data;
      });
    }
  }
};
export const getters = {
  getCollectionById: state => id => {
    return state.collectionsList ? state.collectionsList.find(coll => coll.id === id) : null;
  }
};
