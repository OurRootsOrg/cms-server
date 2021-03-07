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
          :headers="userColumns"
          :items="usersList"
          item-key="id"
          :show-select="false"
          :disable-pagination="true"
          dense
          v-columns-resizable
        >
          <template v-slot:body>
            <tr v-for="user in usersList" :key="user.id">
              <td>{{ user.name }}</td>
              <td>{{ user.email }}</td>
              <td>{{ user.levelName }}</td>
              <td>
                <v-icon v-if="notMe(user)" small @click="editUser(user)" class="mr-3">mdi-pencil</v-icon>
                <v-icon v-if="notMe(user)" small @click="deleteUser(user.id)">mdi-delete</v-icon>
              </td>
            </tr>
          </template>
        </v-data-table>
        <v-dialog v-model="dialogUser" max-width="600px">
          <v-card>
            <v-card-title class="pb-5 mb-0">User</v-card-title>
            <v-card-text>
              <v-container class="pl-0">
                <v-row>
                  <v-col cols="12">
                    <v-text-field
                      dense
                      v-model="editedUser.name"
                      label="Name"
                      placeholder="Name of the person to invite"
                    >
                    </v-text-field>
                  </v-col>
                  <v-col cols="12">
                    <v-select
                      v-model="editedUser.level"
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
              <v-btn color="primary" text @click="closeUser" class="mr-5">Cancel</v-btn>
              <v-btn color="primary" @click="saveUser" :disabled="!editedUser.name || !editedUser.level">Save</v-btn>
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
  Promise.all([store.dispatch("invitationsGetAll"), store.dispatch("usersGetAll")])
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
      // start of data for users table
      dialogUser: false,
      editedUser: {},
      defaultUser: {
        id: 0,
        name: "",
        level: 0,
        levelName: ""
      },
      userColumns: [
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
      val || this.closeUser();
    }
  },
  computed: {
    usersList() {
      return this.users.usersList.map(user => {
        return {
          id: user.id,
          name: user.name,
          email: user.email,
          level: user.level,
          levelName: getAuthLevelName(user.level)
        };
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
    ...mapState(["invitations", "users"])
  },
  methods: {
    //methods for users
    notMe(user) {
      return user.id !== store.getters.userId;
    },
    editUser(user) {
      this.editedUser = Object.assign({}, user);
      this.dialogUser = true;
    },
    deleteUser(id) {
      NProgress.start();
      this.$store
        .dispatch("usersDelete", id)
        .then(() => {
          NProgress.done();
        })
        .catch(() => {
          NProgress.done();
        });
    },
    closeUser() {
      this.dialogUser = false;
      this.$nextTick(() => {
        this.editedUser = Object.assign({}, this.defaultUser);
      });
    },
    saveUser() {
      let user = Object.assign({}, this.editedUser);
      NProgress.start();
      this.$store
        .dispatch("usersUpdate", user)
        .then(() => {
          NProgress.done();
          this.closeUser();
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
