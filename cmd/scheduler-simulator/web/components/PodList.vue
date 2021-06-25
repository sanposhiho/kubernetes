<template>
  <v-card-actions>
    <v-chip
      class="ma-2"
      v-for="(p, i) in pods[nodeName]"
      :key="i"
      @click.stop="onClick(p)"
      color="primary"
      outlined
      large
      label
    >
      <img src="/pod.png" height="40" alt="p.metadata.name" />
      {{ p.metadata.name }}
    </v-chip>
  </v-card-actions>
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
import PodStoreKey from "./pod-store-key";
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
    const onClick = (pod: V1Pod) => {
      store.selectPod(pod);
    };
    onMounted(getPodList);
    const pods = computed(() => store.pods);
    return {
      pods,
      onClick,
    };
  },
});
</script>
