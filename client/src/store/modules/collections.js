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
      state.collectionsList = state.collectionsList.map(c => (c.id === coll.id ? coll : c));
    }
  }
};

export const actions = {
  collectionsCreate({ commit, dispatch, rootGetters }, collection) {
    return Server.collectionsCreate(rootGetters.currentSocietyId, collection)
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
          error,
          type: "error",
          message: "There was a problem creating your collection: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  collectionsUpdate({ commit, dispatch, rootGetters }, coll) {
    return Server.collectionsUpdate(rootGetters.currentSocietyId, coll)
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
          error,
          type: "error",
          message: "There was a problem updating your collection: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  collectionsDelete({ commit, dispatch, rootGetters }, id) {
    return Server.collectionsDelete(rootGetters.currentSocietyId, id)
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
          error,
          type: "error",
          message: "There was a problem deleting the collection: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  collectionsGetAll({ commit, dispatch, rootGetters }) {
    return Server.collectionsGetAll(rootGetters.currentSocietyId)
      .then(response => {
        commit("COLLECTIONS_SET", response.data.collections);
        return response.data.collections;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading collections: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  collectionsGetOne({ commit, dispatch, rootGetters }, id) {
    return Server.collectionsGetOne(rootGetters.currentSocietyId, id)
      .then(response => {
        commit("COLLECTION_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading collection: " + id + " " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};
