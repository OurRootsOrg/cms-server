<!-- Copied from https://github.com/Aymkdn/v-snackbars -->
<template>
  <div>
    <v-snackbar
      v-bind="$attrs"
      v-model="snackbar.show"
      :key="snackbar.id"
      :class="'snackbars snackbars-' + idx + '-' + identifier"
      :timeout="-1"
      :color="snackbar.color"
      v-for="(snackbar, idx) in snackbars"
    >
      {{ snackbar.text }}
      <template v-slot:action>
        <slot name="action" :close="removeMessage" :id="snackbar.id" :index="idx" :message="snackbar.text">
          <v-btn text @click="removeMessage(snackbar.id)">close</v-btn>
        </slot>
      </template>
    </v-snackbar>
    <css-style> .snackbars .v-snack__wrapper { transition: {{ topOrBottom }} 500ms; {{ topOrBottom }}: 0; } </css-style>
    <css-style :key="'snackbars-css' + idx" v-for="idx in len">
      .snackbars.snackbars-{{ idx }}-{{ identifier }} > .v-snack__wrapper { {{ topOrBottom }}:{{ idx * distance }}px; }
    </css-style>
  </div>
</template>

<script>
export default {
  name: "snackbars",
  props: {
    messages: {
      type: Array,
      default: () => []
    },
    idKey: {
      type: String
    },
    textKey: {
      type: String,
      default: "text"
    },
    colorKey: {
      type: String,
      default: "color"
    },
    timeout: {
      type: [Number, String],
      default: 5000
    },
    distance: {
      type: [Number, String],
      default: 55
    }
  },
  data() {
    return {
      topOrBottom: "bottom",
      len: 0, // we need it to have a css transition
      snackbars: [], // array of {id, text, color, show(true)}
      identifier: Date.now() + (Math.random() + "").slice(2)
    };
  },
  components: {
    "css-style": {
      render: function(createElement) {
        return createElement("style", this.$slots.default);
      }
    }
  },
  watch: {
    messages() {
      console.log("snackbars.watch");
      this.setSnackbars();
    }
  },
  methods: {
    setSnackbars() {
      this.snackbars = this.snackbars.filter(snackbar =>
        this.messages.some(message => message[this.idKey] === snackbar.id)
      );
      this.messages.forEach(message => {
        let snackbar = this.snackbars.find(snackbar => snackbar.id === message.id);
        if (snackbar) {
          snackbar.text = message[this.textKey];
          snackbar.color = message[this.colorKey];
        } else {
          let id = message[this.idKey];
          this.snackbars.push({
            id: id,
            text: message[this.textKey],
            color: message[this.colorKey],
            show: true
          });
          if (this.timeout > 0) {
            setTimeout(() => this.removeMessage(id), this.timeout * 1);
          }
        }
      });
      if (this.snackbars.length > this.len) this.len = this.snackbars.length;
    },
    removeMessage(id) {
      this.$emit("remove", id);
    }
  },
  created() {
    if (typeof this.$attrs.top !== "undefined" && this.$attrs.top !== false) this.topOrBottom = "top";
    this.setSnackbars();
  }
};
</script>
