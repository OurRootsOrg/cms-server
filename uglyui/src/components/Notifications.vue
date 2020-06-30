<template>
  <div class="notifications">
    <Notification
      v-for="notification in notifications.notificationsList"
      :key="notification.id"
      :notification="notification"
    />
  </div>
</template>

<script>
import Notification from "@/components/Notification.vue";
import { mapState } from "vuex";

export default {
  components: {
    Notification
  },
  computed: mapState(["notifications"])
};
</script>

<style scoped>
.notifications {
  position: fixed;
  top: 0;
  right: 0;
  padding-right: 40px;
}
</style>
