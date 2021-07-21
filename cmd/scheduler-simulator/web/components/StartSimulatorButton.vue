<template>
  <v-btn color="primary" @click="onClick">Start Simulator</v-btn>
</template>

<script lang="ts">
import {
  defineComponent,
  reactive,
  watchEffect,
} from "@nuxtjs/composition-api";
import { createNamespace } from "~/api/v1/namespace";

export default defineComponent({
  setup(_, context) {
    const router = context.root.$router;

    const onClick = async () => {
      var n = await createNamespace();
      router.push(`/${n.metadata?.name}`);
    };
    return {
      onClick,
    };
  },
});
</script>
