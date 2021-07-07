<template>
  <v-card class="ma-2" outlined v-if="pods['unscheduled'].length !== 0">
    <v-card-title class="mb-1"> Unscheduled Pods </v-card-title>
    <PodList nodeName="unscheduled" />
  </v-card>
</template>

<script lang="ts">
import { V1Pod } from "@kubernetes/client-node";
import {
  ref,
  computed,
  inject,
  onMounted,
  PropType,
  defineComponent,
} from "@nuxtjs/composition-api";
import { getSimulatorIDFromPath } from "./lib/util";
import PodStoreKey from "./PodStoreKey";
export default defineComponent({
  setup(_, context) {
    const store = inject(PodStoreKey);
    if (!store) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

     const getPodList = async () => {
      const route = context.root.$route;
      await store.list(getSimulatorIDFromPath(route.path));
    };
    onMounted(getPodList);

    const pods = computed(() => store.pods);
    return {
      pods,
    };
  },
});
</script>