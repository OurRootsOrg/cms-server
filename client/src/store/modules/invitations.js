import Server from "@/services/Server.js";

export const state = {
  invitationsList: [],
  invitation: {}
};

export const mutations = {
  INVITATIONS_ADD(state, invitation) {
    state.invitationsList.push(invitation);
  },
  INVITATIONS_SET(state, invitations) {
    state.invitationsList = invitations;
  },
  INVITATIONS_REMOVE(state, id) {
    state.invitationsList = state.invitationsList.filter(inv => inv.id !== id);
  },
  INVITATION_SET(state, invitation) {
    state.invitation = invitation;
  }
};

export const actions = {
  invitationsCreate({ commit, dispatch, rootGetters }, invitation) {
    return Server.invitationsCreate(rootGetters.currentSocietyId, invitation)
      .then(response => {
        commit("INVITATIONS_ADD", response.data);
        const notification = {
          type: "success",
          message: "Your invitation has been created - remember to send the URL to the invitee"
        };
        dispatch("notificationsAdd", notification, { root: true });
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem creating your invitation: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  invitationsGetAll({ commit, dispatch, rootGetters }) {
    return Server.invitationsGetAll(rootGetters.currentSocietyId)
      .then(response => {
        commit("INVITATIONS_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading invitations: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  invitationsDelete({ commit, dispatch, rootGetters }, id) {
    return Server.invitationsDelete(rootGetters.currentSocietyId, id)
      .then(() => {
        commit("INVITATIONS_REMOVE", id);
        const notification = {
          type: "success",
          message: "Your invitation has been deleted"
        };
        dispatch("notificationsAdd", notification, { root: true });
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem deleting the invitation: " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  invitationGetForCode({ commit, dispatch }, code) {
    return Server.invitationGetForCode(code)
      .then(response => {
        commit("INVITATION_SET", response.data);
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem reading your invitation - have you already accepted it? " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  },
  invitationAccept({ commit, dispatch }, code) {
    return Server.invitationAccept(code)
      .then(response => {
        commit("INVITATION_SET", {});
        return response.data;
      })
      .catch(error => {
        const notification = {
          error,
          type: "error",
          message: "There was a problem accepting your invitation - have you already accepted it? " + error.message
        };
        dispatch("notificationsAdd", notification, { root: true });
        throw error;
      });
  }
};
