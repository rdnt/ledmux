<template>
  <div id="customizer" :class="{ show: selected != null }">
    <div class="border" v-on:click="addBounds">
      <div
        v-if="from.x && from.y"
        class="point show"
        :style="{ left: from.x, top: from.y }"
      ></div>
      <div
        v-if="to.x && to.y"
        class="point show"
        :style="{ left: to.x, top: to.y }"
      ></div>
    </div>
    <div class="overlay">
      <!-- from {{ from }} to {{ to }} -->
    </div>
  </div>
</template>

<script>
import { mapState } from "vuex";
export default {
  name: "Customizer",
  mounted() {},
  computed: mapState({
    displays: state => state.displays,
    selected: state => state.selected
  }),
  data() {
    return {
      step: 0,
      from: { x: null, y: null },
      to: { x: null, y: null }
    };
  },
  methods: {
    addBounds(event) {
      var width = Math.max(
        document.documentElement.clientWidth,
        window.innerWidth || 0
      );
      var height = Math.max(
        document.documentElement.clientHeight,
        window.innerHeight || 0
      );

      var x = event.x,
        y = event.y;

      if (this.step < 2) {
        if (x <= 20) {
          x = 10;
        }
        if (y <= 20) {
          y = 10;
        }
        if (width - x <= 20) {
          x = width - 11;
        }
        if (height - y <= 20) {
          y = height - 11;
        }

        var percentX = (x / (width - 1)) * 100 + "%";
        var percentY = (y / (height - 1)) * 100 + "%";
        var coords = { x: percentX, y: percentY };
        if (this.step == 0) {
          this.from = coords;
        } else if (this.step == 1) {
          this.to = coords;
        }
        this.step++;
      }
    }
  }
};
</script>
