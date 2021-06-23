import { reactive } from "@nuxtjs/composition-api";
import { createNode, listNode } from "~/api/v1/node";
import { V1Node, V1NodeList, V1Pod, V1PodList } from "@kubernetes/client-node";

type stateType = {
  count: number;
  nodes: V1Node[];
};

export default function nodeStore() {
  const state: stateType = reactive({
    count: 0,
    nodes: [],
  });

  return {
    get count() {
      return state.count;
    },

    get nodes() {
      return state.nodes;
    },

    increment() {
      state.count += 1;
    },

    decrement() {
      state.count -= 1;
    },

    async getNodes() {
      state.nodes = (await listNode({})).items;
    },

    async createNewNode(name: string) {
      await createNode({
        metadata: {
          name: name,
        },
      });
      await this.getNodes();
    },
  };
}

export type NodeStore = ReturnType<typeof nodeStore>;
