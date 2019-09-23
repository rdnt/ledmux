<template>
  <div id="app">
    <fullscreen
      ref="fullscreen"
      @change="fullscreenChange"
      :background="background"
    >
      <Selector></Selector>
      <Customizer></Customizer>
      <button type="button" @click="toggle">Fullscreen</button>
    </fullscreen>
  </div>
</template>
<script>
import "./assets/scss/App.scss";
import Selector from "@/components/Selector.vue";
import Customizer from "@/components/Customizer.vue";
export default {
  name: "App",
  components: {
    Selector,
    Customizer
  },
  methods: {
    toggle() {
      this.$refs["fullscreen"].toggle();
    },
    fullscreenChange(fullscreen) {
      this.fullscreen = fullscreen;
    }
  },
  data() {
    return {
      fullscreen: false,
      background: "linear-gradient(to right, #1d1c2c, #1d1c2c)"
    };
  }
};
</script>
