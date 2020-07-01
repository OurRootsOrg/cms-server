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
  COLLECTIONS_REMOVE(state, id) {
    state.collectionsList = state.collectionsList.filter(coll => coll.id !== id);
  },
  COLLECTION_SET(state, coll) {
    if (!coll.fields) coll.fields = []; // force empty array
    if (!coll.mappings) coll.mappings = []; // force empty array
    state.collection = coll;
  },
  COLLECTION_UPDATE(state, coll) {
    if (state.collection && state.collection.id === coll.id) {
      state.collection = coll;
    }
    if (state.collectionsList) {
      for (let i = 0; i < state.collectionsList.length; i++) {
        if (state.collectionsList[i].id === coll.id) {
          Object.assign(state.collectionsList[i], coll);
        }
      }
    }
  }
};

export const actions = {
  collectionsCreate({ commit, dispatch }, collection) {
    return Server.collectionsCreate(collection)
      .then(response => {
        commit("COLLECTIONS_ADD", response.data);
        const notification = {
          type: "success",
          message: "Your collection has been created"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
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
          message: "Your collection has been updated"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem updating your collection: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  },
  collectionsDelete({ commit, dispatch }, id) {
    return Server.collectionsDelete(id)
      .then(() => {
        commit("COLLECTIONS_REMOVE", id);
        const notification = {
          type: "success",
          message: "Your collection has been deleted"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem deleting the collection: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  },
  collectionsGetAll({ commit, dispatch }) {
    return Server.collectionsGetAll()
      .then(response => {
        commit("COLLECTIONS_SET", response.data.collections);
        return response.data.collections;
      })
      .catch(error => {
        const notification = {
          type: "error",
          message: "There was a problem fetching collections: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
      });
  },
  collectionsGetOne({ commit }, id) {
    return Server.collectionsGetOne(id).then(response => {
      commit("COLLECTION_SET", response.data);
      return response.data;
    });
  }
};
