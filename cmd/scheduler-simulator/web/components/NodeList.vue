<template>
<v-container>
  <v-row>
          <v-col >
    <v-card>
      <v-card-title> Nodes </v-card-title>
      <v-row no-gutters>
        <template v-for="(n, i) in nodes">
          <v-col tile :key="i">
            <v-card class="pa-2" outlined>
              <v-card-title v-text="n.metadata.name"> </v-card-title>
              <PodList :nodeName="n.metadata.name" />
            </v-card>
          </v-col>
          <v-responsive
            v-if="(i % 3) ==2"
            :key="`width-${n}`"
            width="100%"
          ></v-responsive>
        </template>
      </v-row>
    </v-card>
    <v-col class="mx-auto" tile>
      <v-card>
        <v-card-title> Unscheduled Pods </v-card-title>
        <PodList nodeName="unscheduled" />
      </v-card>
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

export default defineComponent({
  components: { PodList },
  setup() {
    const store = inject(NodeStoreKey);
    if (!store) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const getNodeList = async () => {
      await store.getNodes();
    };

    onMounted(getNodeList);

    const nodes = computed(() => store.nodes);

    return {
      nodes,
    };
  },
});
</script>
