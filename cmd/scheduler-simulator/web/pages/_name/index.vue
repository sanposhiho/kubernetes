<template>
  <PodStoreProvider>
    <NodeStoreProvider>
      <PVStoreProvider>
        <PVCStoreProvider>
          <StorageClassStoreProvider>
            <ResourceBar />
            <ResourceAddButton />
            <NodeList />
            <UnscheduledPodList />
            <PVList />
            <PVCList />
            <StorageClassList />
          </StorageClassStoreProvider>
        </PVCStoreProvider>
      </PVStoreProvider>
    </NodeStoreProvider>
  </PodStoreProvider>
</template>

<script>
import { computed, defineComponent, inject } from "@nuxtjs/composition-api";
import NodeStoreProvider from "~/components/StoreProvider/NodeStoreProvider.vue";
import PVStoreProvider from "~/components/StoreProvider/PVStoreProvider.vue";
import PVCStoreProvider from "~/components/StoreProvider/PVCStoreProvider.vue";
import StorageClassStoreProvider from "~/components/StoreProvider/StorageClassStoreProvider.vue";
import NodeList from "~/components/NodeList.vue";
import UnscheduledPodList from "~/components/UnscheduledPodList.vue";
import PVList from "~/components/PVList.vue";
import PVCList from "~/components/PVCList.vue";
import StorageClassList from "~/components/StorageClassList.vue";
import PodStoreProvider from "~/components/StoreProvider/PodStoreProvider.vue";
import ResourceAddButton from "~/components/ResourceAddButton.vue";
import ResourceBar from "~/components/ResourceBar/ResourceBar.vue";

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
