import { reactive } from "@nuxtjs/composition-api";
import { applyPod, deletePod, listPod } from "~/api/v1/pod";
import { V1Pod, V1PodList } from "@kubernetes/client-node";

type stateType = {
  selectedPod: selectedPod | null;
  pods: {
    [key: string]: Array<V1Pod>;
  };
};

type selectedPod = {
  // isNew represents whether this Pod is a new Pod or not.
  isNew: boolean;
  pod: V1Pod;
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

    get selectedPod() {
      return state.selectedPod;
    },

    selectPod(p: V1Pod | null, isNew: boolean) {
      if (p !== null) {
        state.selectedPod = {
          isNew: isNew,
          pod: p,
        };
      } else {
        state.selectedPod = null;
      }
    },

    async listPod() {
      const pods = (await listPod()).items;
      var result: { [key: string]: Array<V1Pod> } = { unscheduled: [] };
      pods.forEach((p) => {
        if (!p.spec) {
          return;
        } else {
          if (p.spec?.nodeName == null) {
            // unscheduled pod
            result["unscheduled"].push(p);
          } else if (!result[p.spec?.nodeName as string]) {
            // first pod on the node
            result[p.spec?.nodeName as string] = [p];
          } else {
            result[p.spec?.nodeName as string].push(p);
          }
        }
      });
      state.pods = result;
    },

    async applyPod(p: V1Pod) {
      await applyPod(p);
      await this.listPod();
    },

    async deletePod(name: string) {
      await deletePod(name);
      await this.listPod();
    },
  };
}

export type PodStore = ReturnType<typeof podStore>;
