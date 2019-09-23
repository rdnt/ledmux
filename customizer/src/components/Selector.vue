<template>
  <div id="selector">
    <div
      class="display reveal"
      v-for="display in displays"
      :key="display.id"
      :class="{
        shrink: display.shrink,
        expand: display.expand,
        visible: display.visible,
        ready: display.width > 0 && display.height > 0,
        rapid: display.rapid,
        clickable: display.clickable
      }"
    >
      <div class="inner" v-on:click="editDisplay(display.id)" s>
        <div class="id">{{ display.id }}</div>
        <div v-if="display.width > 0 && display.height > 0" class="resolution">
          {{ display.width }}
          <span class="separator"> x </span>
          {{ display.height }}
        </div>
        <div v-if="!(display.width > 0 && display.height > 0)" class="prompt">
          Click to configure
        </div>
      </div>
    </div>
    <div
      class="display new"
      :class="{
        shrink: newDisplayBtn.shrink,
        expand: newDisplayBtn.expand,
        visible: newDisplayBtn.visible,
        rapid: newDisplayBtn.rapid,
        clickable: newDisplayBtn.clickable
      }"
    >
      <div class="inner" v-on:click="addDisplay()">
        <div class="plus"></div>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState } from "vuex";
export default {
  name: "Selector",
  mounted() {},
  computed: mapState({
    displays: state => state.displays,
    selected: state => state.selected
  }),
  data() {
    return {
      newDisplayBtn: {
        shrink: false,
        expand: false,
        visible: true,
        rapid: false,
        clickable: true
      }
    };
  },
  methods: {
    addDisplay() {
      var size = this.displays.length;
      if (size == 4) {
        return;
      }
      this.$store.commit("addDisplay");
      if (size == 3) {
        this.newDisplayBtn.visible = false;
        this.newDisplayBtn.shrink = true;
        this.newDisplayBtn.clickable = false;
      }
      setTimeout(() => {
        this.displays[size].visible = true;
      }, 500);
    },
    removeDisplay(index) {
      this.displays.splice(index, 1);
    },
    editDisplay(id) {
      this.newDisplayBtn.clickable = false;
      this.newDisplayBtn.visible = false;
      this.displays.forEach(function(display) {
        if (display.id !== id) {
          display.visible = false;
        }
        display.clickable = false;
      });
      setTimeout(() => {
        this.newDisplayBtn.shrink = true;
        this.displays.forEach(function(display) {
          if (display.id !== id) {
            display.shrink = true;
          }
        });
      }, 500);
      setTimeout(() => {
        this.displays[id - 1].expand = true;
        // this.selected = id;
        this.$store.commit("selectDisplay", id);
      }, 1000);
    }
  }
};
</script>
