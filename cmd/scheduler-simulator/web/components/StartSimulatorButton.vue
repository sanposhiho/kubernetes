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
import { applyNamespace } from "~/api/v1/namespace";

export default defineComponent({
  setup(_, context) {
    const router = context.root.$router;

    const onClick = async () => {
      let name: string = UUID.generate();
      await applyNamespace({ metadata: { name: name } });
      router.push(`/${name}`);
    };
    return {
      onClick,
    };
  },
});
</script>
