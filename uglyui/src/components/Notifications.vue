<template>
  <Snackbars
    :messages="notifications.notificationsList"
    id-key="id"
    color-key="type"
    text-key="message"
    timeout="5000"
    @remove="notificationsRemove"
    top
    right
  ></Snackbars>
</template>

<script>
import Snackbars from "./Snackbars.vue";
import { mapState, mapActions } from "vuex";

export default {
  components: {
    Snackbars: Snackbars
  },
  computed: mapState(["notifications"]),
  methods: mapActions(["notificationsRemove"])
};
</script>

<style scoped>
.notifications {
  position: fixed;
  top: 80px;
  right: 0;
  padding-right: 32px;
}
</style>
