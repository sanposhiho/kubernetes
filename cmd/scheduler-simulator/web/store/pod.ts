import { reactive } from "@nuxtjs/composition-api";
import { applyPod, deletePod, getPod, listPod } from "~/api/v1/pod";
import { V1Pod, V1PodList } from "@kubernetes/client-node";

type stateType = {
  selectedPod: SelectedPod | null;
  pods: {
    // key is node name or "unscheduled"
    [key: string]: Array<V1Pod>;
  };
};

export type SelectedPod = {
  // isNew represents whether this is a new one or not.
  isNew: boolean;
  item: V1Pod;
  resourceKind: string;
};

export default function podStore() {
  const state: stateType = reactive({
    selectedPod: null,
    pods: { unscheduled: [] },
  });

  return {
    get pods() {
      return state.pods;
    },

    get count(): number {
      var num = 0;
      Object.keys(state.pods).forEach((key) => {
        num += state.pods[key].length;
      });
      return num;
    },

    get selected() {
      return state.selectedPod;
    },

    select(p: V1Pod | null, isNew: boolean) {
      if (p !== null) {
        state.selectedPod = {
          isNew: isNew,
          item: p,
          resourceKind: "Pod",
        };
      }
    },

    resetSelected() {
      state.selectedPod = null;
    },

    async fetchlist(simulatorID: string) {
      const pods = (await listPod(simulatorID)).items;
      var result: { [key: string]: Array<V1Pod> } = {};
      result["unscheduled"] = [];
      pods.forEach((p) => {
        if (!p.spec?.nodeName) {
          // unscheduled pod
          result["unscheduled"].push(p);
        } else if (!result[p.spec?.nodeName as string]) {
          // first pod on the node
          result[p.spec?.nodeName as string] = [p];
        } else {
          result[p.spec?.nodeName as string].push(p);
        }
      });
      state.pods = result;
    },

    async fetchSelected(simulatorID: string) {
      if (this.selected?.item.metadata?.name && !this.selected?.isNew) {
        const p = await getPod(this.selected.item.metadata.name, simulatorID);
        this.select(p, this.selected?.isNew);
      }
    },

    async apply(p: V1Pod, simulatorID: string) {
      await applyPod(p, simulatorID);
      await this.fetchlist(simulatorID);
    },

    async delete(name: string, simulatorID: string) {
      await deletePod(name, simulatorID);
      await this.fetchlist(simulatorID);
    },
  };
}

export type PodStore = ReturnType<typeof podStore>;
