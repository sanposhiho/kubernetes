<template>
  <PodStoreProvider>
    <NodeStoreProvider>
      <PVStoreProvider>
        <PVCStoreProvider>
          <StorageClassStoreProvider>
            <ResourceBar />
            <NodeList />
            <UnscheduledPodList />
            <PVList />
            <PVCList />
            <StorageClassList />
            <ResourceAddButton />
          </StorageClassStoreProvider>
        </PVCStoreProvider>
      </PVStoreProvider>
    </NodeStoreProvider>
  </PodStoreProvider>
</template>

<script>
import { computed, defineComponent, inject } from "@nuxtjs/composition-api";
import NodeStoreProvider from "~/components/NodeStoreProvider.vue";
import PVStoreProvider from "~/components/PVStoreProvider.vue";
import PVCStoreProvider from "~/components/PVCStoreProvider.vue";
import StorageClassStoreProvider from "~/components/StorageClassStoreProvider.vue";
import NodeList from "~/components/NodeList.vue";
import UnscheduledPodList from "~/components/UnscheduledPodList.vue";
import PVList from "~/components/PVList.vue";
import PVCList from "~/components/PVCList.vue";
import StorageClassList from "~/components/StorageClassList.vue";
import PodStoreProvider from "~/components/PodStoreProvider.vue";
import ResourceAddButton from "~/components/ResourceAddButton.vue";
import ResourceBar from "~/components/ResourceBar.vue";

export default defineComponent({
  components: {
    NodeStoreProvider,
    PVList,
    PVCList,
    StorageClassList,
    NodeList,
    UnscheduledPodList,
    PodStoreProvider,
    ResourceAddButton,
    ResourceBar,
    StorageClassStoreProvider,
    PVStoreProvider,
    PVCStoreProvider,
  },
  setup() {
    return {};
  },
});
</script>
