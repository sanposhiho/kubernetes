<template>
  <v-btn @click="createPod">Add New Pod</v-btn>
</template>

<script lang="ts">
import { computed, inject, defineComponent } from "@nuxtjs/composition-api";
import PodStoreKey from "./pod-store-key";

export default defineComponent({
  setup() {
    const store = inject(PodStoreKey);
    if (!store) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const createPod = async () => {
      await store.createPod("sample-pod-" + store.count);
      store.increment();
    };

    return {
      createPod,
    };
  },
});
</script>
