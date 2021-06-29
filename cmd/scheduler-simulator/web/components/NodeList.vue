<template>
  <v-container>
    <v-row>
      <v-col>
        <v-col>
          <v-card-title class="text-h5 mb-1"> Nodes </v-card-title>
          <v-row no-gutters>
            <template v-for="(n, i) in nodes">
              <v-col tile :key="i" cols="auto">
                <v-card class="ma-2" outlined @click="onClick(n)">
                  <v-card-title v-text="n.metadata.name" class="mb-1">
                  </v-card-title>
                  <PodList :nodeName="n.metadata.name" />
                </v-card>
              </v-col>
            </template>
          </v-row>
        </v-col>
        <v-col>
          <v-card-title class="text-h5 mb-1"> Unscheduled Pods </v-card-title>
          <v-row no-gutters>
            <v-col>
              <v-card class="ma-2" outlined>
                <PodList nodeName="unscheduled" />
              </v-card>
            </v-col>
          </v-row>
        </v-col>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import NodeStoreKey from "./node-store-key";
import PodList from "~/components/PodList.vue";
import { V1Node } from "@kubernetes/client-node";
import PodStoreKey from "./pod-store-key";
import { getSimulatorIDFromPath } from "./lib/util";

export default defineComponent({
  components: { PodList },
  setup(_, context) {
    const pstore = inject(PodStoreKey);
    if (!pstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const nstore = inject(NodeStoreKey);
    if (!nstore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const route = context.root.$route;

    const getNodeList = async () => {
      await nstore.listNode(getSimulatorIDFromPath(route.path));
    };

    onMounted(getNodeList);

    const nodes = computed(() => nstore.nodes);
    const pods = computed(() => pstore.pods);

    const onClick = (node: V1Node) => {
      nstore.selectNode(node, false);
    };

    return {
      pods,
      nodes,
      onClick,
    };
  },
});
</script>
