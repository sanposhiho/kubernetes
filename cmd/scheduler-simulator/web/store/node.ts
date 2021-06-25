import { reactive } from "@nuxtjs/composition-api";
import { applyNode, listNode } from "~/api/v1/node";
import { V1Node, V1NodeList, V1Pod, V1PodList } from "@kubernetes/client-node";

type stateType = {
  count: number;
  selectedNode: V1Node | null;
  nodes: V1Node[];
};

export default function nodeStore() {
  const state: stateType = reactive({
    count: 0,
    selectedNode: null,
    nodes: [],
  });

  return {
    get count() {
      return state.count;
    },

    get nodes() {
      return state.nodes;
    },

    get selectedNode() {
      return state.selectedNode;
    },

    increment() {
      state.count += 1;
    },

    decrement() {
      state.count -= 1;
    },

    selectNode(n: V1Node | null) {
      state.selectedNode = n;
    },

    async getNodes() {
      state.nodes = (await listNode({})).items;
    },

    async createNode(name: string) {
      await applyNode({
        metadata: {
          name: name,
        },
      });
      await this.getNodes();
    },

    async editNode(n: V1Node) {
      await applyNode(n);
      await this.getNodes();
    },
  };
}

export type NodeStore = ReturnType<typeof nodeStore>;
