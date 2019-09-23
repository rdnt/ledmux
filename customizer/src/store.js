import Vue from "vue";
import Vuex from "vuex";

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    selected: null,
    displays: []
  },
  mutations: {
    selectDisplay(state, id) {
      state.selected = id;
      var width = Math.max(
        document.documentElement.clientWidth,
        window.innerWidth || 0
      );
      var height = Math.max(
        document.documentElement.clientHeight,
        window.innerHeight || 0
      );
      state.displays[id - 1].width = width;
      state.displays[id - 1].height = height;
    },
    addDisplay(state) {
      var size = state.displays.length;
      if (size == 4) {
        return;
      }
      state.displays.push({
        id: size + 1,
        width: 0,
        height: 0,
        shrink: false,
        expand: false,
        visible: false,
        rapid: false,
        clickable: true
      });
      setTimeout(() => {
        // this.newDisplayBtn.rapid = true;
        state.displays[size].visible = true;
      }, 500);
    }
  },
  actions: {}
});
