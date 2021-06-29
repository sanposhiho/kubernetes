import { reactive } from "@nuxtjs/composition-api";
import { applyNode, deleteNode, listNode } from "~/api/v1/node";
import { V1Node, V1NodeList, V1Pod, V1PodList } from "@kubernetes/client-node";

type stateType = {
  selectedNode: selectedNode | null;
  nodes: V1Node[];
};

type selectedNode = {
  // isNew represents whether this Pod is a new Node or not.
  isNew: boolean;
  node: V1Node;
};

export default function nodeStore() {
  const state: stateType = reactive({
    selectedNode: null,
    nodes: [],
  });

  return {
    get nodes() {
      return state.nodes;
    },

    get selectedNode() {
      return state.selectedNode;
    },

    selectNode(n: V1Node | null, isNew: boolean) {
      if (n !== null) {
        state.selectedNode = {
          isNew: isNew,
          node: n,
        };
      }
    },

    resetSelectNode() {
      state.selectedNode = null;
    },

    async listNode(simulatorID: string) {
      state.nodes = (await listNode(simulatorID)).items;
    },

    async applyNode(n: V1Node, simulatorID: string) {
      await applyNode(n, simulatorID);
      await this.listNode(simulatorID);
    },

    async deleteNode(name: string, simulatorID: string) {
      await deleteNode(name, simulatorID);
      await this.listNode(simulatorID);
    },
  };
}

export type NodeStore = ReturnType<typeof nodeStore>;
