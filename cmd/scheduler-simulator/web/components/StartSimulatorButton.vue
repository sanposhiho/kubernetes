<template>
  <v-btn color="primary" @click="onClick">Start Simulator</v-btn>
</template>

<script lang="ts">
import UUID from "uuidjs";
import {
  defineComponent,
  reactive,
  watchEffect,
} from "@nuxtjs/composition-api";

export default defineComponent({
  setup(_, context) {
    const router = context.root.$router;

    const onClick = () => {
      let id: string = UUID.generate();
      router.push(`/simulator/${id}`);
    };
    return {
      onClick,
    };
  },
});
</script>
