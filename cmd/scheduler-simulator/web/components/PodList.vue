<template>
  <v-list>
    <v-list-group v-for="(p, i) in pods[nodeName]" :key="i">
      <template v-slot:activator>
        <v-list-item-title v-text="p.metadata.name"> </v-list-item-title>
      </template>
      <v-list-item v-for="(v, i) in p.spec.containers" :key="i">
        <v-list-item-content>
          <v-list-item-title v-text="v.name"></v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list-group>
  </v-list>
</template>

<script lang="ts">
import {
  ref,
  computed,
  inject,
  onMounted,
  PropType,
  defineComponent,
} from "@nuxtjs/composition-api";
import PodStoreKey from "./pod-store-key";
import { V1Pod } from "@kubernetes/client-node";

type Props = {
  nodeName: string;
};
export default defineComponent({
  props: {
    nodeName: {
      type: String,
      required: true,
    },
  },
  setup(props: Props) {
    const store = inject(PodStoreKey);
    if (!store) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const getPodList = async () => {
      await store.getPods();
    };

    onMounted(getPodList);

    const pods = computed(() => store.pods);

    return {
      pods,
    };
  },
});
</script>
