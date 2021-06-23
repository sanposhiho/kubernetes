<template>
  <v-btn @click="createNode">Add New Node</v-btn>
</template>

<script lang="ts">
import { computed, inject, defineComponent } from "@nuxtjs/composition-api";
import NodeStoreKey from "./node-store-key";

export default defineComponent({
  setup() {
    const store = inject(NodeStoreKey);
    if (!store) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const createNode = async () => {
      await store.createNewNode("sample-node-" + store.count);
      store.increment();
    };

    return {
      createNode,
    };
  },
});
</script>
