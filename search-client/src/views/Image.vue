<template>
  <v-container>
    <div>
      <v-btn text class="pl-0 ml-0 primary--text" @click="$router.go(-1)"><v-icon>mdi-chevron-left</v-icon> Back</v-btn>
      <v-progress-circular v-if="loading" :size="24" indeterminate></v-progress-circular>
    </div>
    <div v-if="isPrivate">
      <p style="margin: 32px 8px;">
        The full image is available to members of the society.
        <a v-if="loginURL" :href="loginURL">Click here to become a member</a>
      </p>
    </div>
    <div v-else class="wrapper">
      <div id="openseadragonToolbar" class="toolbar"></div>
      <div id="openseadragon" class="openseadragon"></div>
    </div>
  </v-container>
</template>
<script>
import Server from "@/services/Server";
import OpenSeadragon from "openseadragon";

export default {
  props: {
    pid: {
      type: String
    },
    path: {
      type: String
    }
  },
  mounted() {
    if ((this.$route.params && this.$route.params.societyId, this.$route.params.pid && this.$route.params.path)) {
      Server.postsGetImage(this.$route.params.societyId, this.$route.params.pid, this.$route.params.path, false).then(
        result => {
          if (result.data.private) {
            this.isPrivate = true;
            this.loginURL = result.data.loginURL;
          } else {
            let pyramid = {
              type: "legacy-image-pyramid",
              levels: [
                {
                  url: result.data.url,
                  height: result.data.height,
                  width: result.data.width
                }
              ]
            };
            console.log("pyramid", pyramid);
            this.osd = OpenSeadragon({
              id: "openseadragon",
              toolbar: "openseadragonToolbar",
              tileSources: pyramid,
              prefixUrl: (window.ourroots.images_directory || "") + "img/seadragon/",
              autoHideControls: false,
              showRotationControl: true //ROTATION
            });
            this.loading = true;
            this.osd.addOnceHandler("tile-drawn", () => {
              this.loading = false;
            });
          }
        }
      );
    }
  },
  data() {
    return {
      loading: false,
      isPrivate: false,
      loginURL: ""
    };
  }
};
</script>

<style scoped>
.wrapper {
  position: relative;
  padding-top: 12px;
}
.openseadragon {
  width: 100%;
  /* max-width: 1024px; */
  height: 600px;
  border: 1px solid black;
  color: #333; /* text color for messages */
  background-color: #f1f1f1;
}
.toolbar {
  width: 100%;
  /* max-width: 1024px; */
  height: 33px;
  border: none;
  color: #333;
  padding: 4px;
  background-color: transparent;
}
.toolbar.fullpage {
  width: 100%;
  border: none;
  position: fixed;
  z-index: 999999;
  left: 0;
  top: 0;
  background-color: #ccc;
}
</style>
