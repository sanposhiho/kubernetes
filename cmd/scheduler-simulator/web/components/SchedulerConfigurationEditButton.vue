<template>
  <v-btn class="ma-2" @click="onClick">
    <v-icon> mdi-cog </v-icon>
  </v-btn>
</template>

<script>
import { defineComponent } from "@vue/composition-api";

export default defineComponent({
  props: {
    value: Boolean,
  },
  setup(props, { emit }) {
    const onClick = () => {
      emit("input", true);
    };
    return {
      onClick,
    };
  },
});
</script>
