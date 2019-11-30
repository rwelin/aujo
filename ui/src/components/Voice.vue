<template>
  <div class="voice">
    <input
      v-for="(v, k) in voice"
      v-bind:key="k"
      type="range"
      orient="vertical"
      min="0"
      max="1000"
      v-model="voice[k]"
      @input="updateInstrument"
    />
  </div>
</template>

<script lang="ts">
import { Component, Prop, Vue, Watch } from "vue-property-decorator";

@Component
export default class Voice extends Vue {
  voice: Record<number, number> = {
    0: 0,
    1: 0,
    2: 0,
    3: 0,
    4: 0,
    5: 0,
    6: 0,
    7: 0,
    8: 0,
    9: 0,
    10: 0,
    11: 0,
    12: 0,
    13: 0,
    14: 0,
    15: 0,
    16: 0,
    17: 0,
    18: 0,
    19: 0
  };

  beforeMount() {
    this.$http
      .get("http://localhost:7999/mix")
      .then(response => {
        return response.json();
      })
      .then(j => {
        const harm = j["Instruments"][0]["Harmonics"];
        let voice: Record<number, number> = {};
        for (let i = 0; i < harm.length; i++) {
          voice[i] = Math.sqrt(harm[i] * 1000000);
        }
        this.voice = voice;
      });
  }

  updateInstrument() {
    let voice = Object.values(this.voice).map(i => (i * i) / 1000000);
    this.$http.put("http://localhost:7999/instruments/0", voice);
  }
}
</script>

<style scoped lang="scss">
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
