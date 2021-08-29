<template>
  <div>{{ message }}</div>
</template>

<script>
import Server from "@/services/Server.js";
import store from "@/store";

export default {
  name: "Downloads",
  created() {
    Server.contentGetRequest(store.getters.currentSocietyId, this.$route.params.key).then(result => {
      window.location.href = result.data.signedURL;
      this.message = "The file has been downloaded";
    });
  },
  data() {
    return {
      message: "Please wait..."
    };
  }
};
</script>
