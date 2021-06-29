<template>
  <PodStoreProvider>
    <NodeStoreProvider>
      <ResourceBar />
      <NodeList />
      <NodeAddButton />
      <PodAddButton />
    </NodeStoreProvider>
  </PodStoreProvider>
</template>

<script>
import { computed, defineComponent, inject } from "@nuxtjs/composition-api";
import NodeStoreProvider from "~/components/NodeStoreProvider.vue";
import NodeList from "~/components/NodeList.vue";
import NodeAddButton from "~/components/NodeAddButton.vue";
import PodStoreProvider from "~/components/PodStoreProvider.vue";
import PodAddButton from "~/components/PodAddButton.vue";
import ResourceBar from "~/components/ResourceBar.vue";

export default defineComponent({
  components: {
    NodeStoreProvider,
    NodeList,
    NodeAddButton,
    PodStoreProvider,
    PodAddButton,
    ResourceBar,
  },
  setup() {
    return {};
  },
});
</script>
