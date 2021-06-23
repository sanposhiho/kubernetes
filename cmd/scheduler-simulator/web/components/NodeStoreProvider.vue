<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { provide, defineComponent } from "@nuxtjs/composition-api";
import nodeStore from "../store/node";
import NodeStoreKey from "./node-store-key";

export default defineComponent({
  setup() {
    provide(NodeStoreKey, nodeStore());
    return {};
  },
});
</script>
