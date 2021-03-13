<template>
  <v-container class="users-list">
    <v-row>
      <v-col cols="12">
        <h1>Users and Invitations</h1>
      </v-col>
    </v-row>

    <v-row no-gutters>
      <v-col cols="12">
        <h3 style="margin-top: 16px;">Users</h3>
        <v-data-table
          :headers="societyUserColumns"
          :items="societyUsersList"
          item-key="id"
          :show-select="false"
          :disable-pagination="true"
          dense
          v-columns-resizable
        >
          <template v-slot:body>
            <tr v-for="societyUser in societyUsersList" :key="societyUser.id">
              <td>{{ societyUser.name }}</td>
              <td>{{ societyUser.email }}</td>
              <td>{{ societyUser.levelName }}</td>
              <td>
                <v-icon small @click="editSocietyUser(societyUser)" class="mr-3">mdi-pencil</v-icon>
                <v-icon v-if="notMe(societyUser)" small @click="deleteSocietyUser(societyUser.id)">mdi-delete</v-icon>
              </td>
            </tr>
          </template>
        </v-data-table>
        <v-dialog v-model="dialogSocietyUser" max-width="600px">
          <v-card>
            <v-card-title class="pb-5 mb-0">User</v-card-title>
            <v-card-text>
              <v-container class="pl-0">
                <v-row>
                  <v-col cols="12">
                    <v-text-field
                      dense
                      v-model="editedSocietyUser.name"
                      label="Name"
                      placeholder="Name of the person to invite"
                    >
                    </v-text-field>
                  </v-col>
                  <v-col cols="12">
                    <v-select
                      v-model="editedSocietyUser.level"
                      label="Authorization level (Editor or Admin)"
                      :items="authLevels"
                      item-text="name"
                      item-value="id"
                      :disabled="!notMe(editedSocietyUser)"
                    >
                    </v-select>
                  </v-col>
                </v-row>
              </v-container>
            </v-card-text>
            <v-card-actions class="pb-5 pr-5">
              <v-spacer></v-spacer>
              <v-btn color="primary" text @click="closeSocietyUser" class="mr-5">Cancel</v-btn>
              <v-btn
                color="primary"
                @click="saveSocietyUser"
                :disabled="!editedSocietyUser.name || !editedSocietyUser.level"
                >Save</v-btn
              >
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-col>
    </v-row>

    <v-row no-gutters>
      <v-col cols="12">
        <h3 style="margin-top: 16px;">Invitations</h3>
        <p class="caption">
          OurRoots does not send invitations automatically. After adding an invitation, send the URL in the
          &quot;Invitation URL&quot; column to the invitee and ask them to click on it.
        </p>
        <v-data-table
          :headers="invitationColumns"
          :items="invitationsList"
          item-key="id"
          :show-select="false"
          :disable-pagination="true"
          dense
          v-columns-resizable
        >
          <template v-slot:body>
            <tr v-for="invitation in invitationsList" :key="invitation.id">
              <td>{{ invitation.name }}</td>
              <td>{{ invitation.levelName }}</td>
              <td>{{ invitation.url }}</td>
              <td>
                <v-icon small @click="deleteInvitation(invitation.id)">mdi-delete</v-icon>
              </td>
            </tr>
          </template>
          <template v-slot:footer>
            <v-toolbar flat class="ml-n3">
              <v-dialog v-model="dialogInvitation" max-width="600px">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn class="secondary primary--text mr-3" v-bind="attrs" v-on="on" small>Add an invitation</v-btn>
                </template>
                <v-card>
                  <v-card-title class="pb-5 mb-0">Invitation</v-card-title>
                  <v-card-text>
                    <v-container class="pl-0">
                      <v-row>
                        <v-col cols="12">
                          <v-text-field
                            dense
                            v-model="editedInvitation.name"
                            label="Name"
                            placeholder="Name of the person to invite"
                          >
                          </v-text-field>
                        </v-col>
                        <v-col cols="12">
                          <v-select
                            v-model="editedInvitation.level"
                            label="Authorization level (Editor or Admin)"
                            :items="authLevels"
                            item-text="name"
                            item-value="id"
                          >
                          </v-select>
                        </v-col>
                      </v-row>
                    </v-container>
                  </v-card-text>
                  <v-card-actions class="pb-5 pr-5">
                    <v-spacer></v-spacer>
                    <v-btn color="primary" text @click="closeInvitation" class="mr-5">Cancel</v-btn>
                    <v-btn
                      color="primary"
                      @click="saveInvitation"
                      :disabled="!editedInvitation.name || !editedInvitation.level"
                      >Save</v-btn
                    >
                  </v-card-actions>
                </v-card>
              </v-dialog>
            </v-toolbar>
          </template>
        </v-data-table>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import store from "@/store";
import { mapState } from "vuex";
import NProgress from "nprogress";
import { getAuthLevelName, getAuthLevelOptions } from "@/utils/authLevels";

const invitationURLPrefix = "https://db.ourroots.org?code=";

function getContent(next) {
  Promise.all([store.dispatch("invitationsGetAll"), store.dispatch("societyUsersGetAll")])
    .then(() => {
      next();
    })
    .catch(() => {
      next("/");
    });
}

export default {
  beforeRouteEnter: function(routeTo, routeFrom, next) {
    console.log("usersList.beforeRouteEnter");
    getContent(next);
  },
  beforeRouteUpdate: function(routeTo, routeFrom, next) {
    console.log("usersList.beforeRouteUpdate");
    getContent(next);
  },
  data() {
    return {
      // start of data for society users table
      dialogSocietyUser: false,
      editedSocietyUser: {},
      defaultSocietyUser: {
        id: 0,
        name: "",
        level: 0,
        levelName: ""
      },
      societyUserColumns: [
        {
          text: "Name",
          value: "name"
        },
        {
          text: "Email",
          value: "email"
        },
        {
          text: "Level",
          value: "levelName"
        },
        {
          text: "",
          value: "actions",
          width: 40,
          align: "right"
        }
      ],
      // start of data for invitations table
      dialogInvitation: false,
      editedInvitation: {},
      defaultInvitation: {
        id: 0,
        name: "",
        level: 0,
        levelName: "",
        code: "",
        url: ""
      },
      authLevels: getAuthLevelOptions(),
      invitationColumns: [
        {
          text: "Name",
          value: "name"
        },
        {
          text: "Level",
          value: "levelName"
        },
        {
          text: "URL (send this to the invitee)",
          value: "url"
        },
        {
          text: "",
          value: "actions",
          width: 40,
          align: "right"
        }
      ]
    };
  },
  watch: {
    dialogInvitation(val) {
      val || this.closeInvitation();
    },
    dialogUser(val) {
      val || this.closeSocietyUser();
    }
  },
  computed: {
    societyUsersList() {
      return this.societyUsers.societyUsersList.map(societyUser => {
        return Object.assign({ levelName: getAuthLevelName(societyUser.level) }, societyUser);
      });
    },
    invitationsList() {
      return this.invitations.invitationsList.map(inv => {
        return {
          id: inv.id,
          name: inv.name,
          level: inv.level,
          levelName: getAuthLevelName(inv.level),
          code: inv.code,
          url: invitationURLPrefix + inv.code
        };
      });
    },
    ...mapState(["invitations", "societyUsers"])
  },
  methods: {
    //methods for users
    notMe(societyUser) {
      return societyUser.userId !== store.getters.userId;
    },
    editSocietyUser(societyUser) {
      this.editedSocietyUser = Object.assign({}, societyUser);
      this.dialogSocietyUser = true;
    },
    deleteSocietyUser(id) {
      NProgress.start();
      this.$store
        .dispatch("societyUsersDelete", id)
        .then(() => {
          NProgress.done();
        })
        .catch(() => {
          NProgress.done();
        });
    },
    closeSocietyUser() {
      this.dialogSocietyUser = false;
      this.$nextTick(() => {
        this.editedSocietyUser = Object.assign({}, this.defaultSocietyUser);
      });
    },
    saveSocietyUser() {
      let societyUser = Object.assign({}, this.editedSocietyUser);
      NProgress.start();
      this.$store
        .dispatch("societyUsersUpdate", societyUser)
        .then(() => {
          NProgress.done();
          this.closeSocietyUser();
        })
        .catch(() => {
          NProgress.done();
        });
    },
    //methods for invitations
    deleteInvitation(id) {
      NProgress.start();
      this.$store
        .dispatch("invitationsDelete", id)
        .then(() => {
          NProgress.done();
        })
        .catch(() => {
          NProgress.done();
        });
    },
    closeInvitation() {
      this.dialogInvitation = false;
      this.$nextTick(() => {
        this.editedInvitation = Object.assign({}, this.defaultInvitation);
      });
    },
    saveInvitation() {
      let invitation = Object.assign({}, this.editedInvitation);
      NProgress.start();
      this.$store
        .dispatch("invitationsCreate", invitation)
        .then(() => {
          NProgress.done();
          this.closeInvitation();
        })
        .catch(() => {
          NProgress.done();
        });
    }
  }
};
</script>

<style scoped></style>
