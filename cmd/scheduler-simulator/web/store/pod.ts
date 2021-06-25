import { reactive } from "@nuxtjs/composition-api";
import { applyPod, listPod } from "~/api/v1/pod";
import { V1Pod, V1PodList } from "@kubernetes/client-node";

type stateType = {
  count: number;
  selectedPod: V1Pod | null;
  pods: {
    [key: string]: Array<V1Pod>;
  };
};

export default function podStore() {
  const state: stateType = reactive({
    count: 0,
    selectedPod: null,
    pods: { unscheduled: [] },
  });

  return {
    get count() {
      return state.count;
    },

    get pods() {
      return state.pods;
    },

    get selectedPod() {
      return state.selectedPod;
    },

    increment() {
      state.count += 1;
    },

    decrement() {
      state.count -= 1;
    },

    selectPod(p: V1Pod | null) {
      state.selectedPod = p;
    },

    async getPods() {
      const pods = (await listPod({})).items;
      var result: { [key: string]: Array<V1Pod> } = { unscheduled: [] };
      pods.forEach((p) => {
        if (!p.spec) {
          return;
        } else {
          if (p.spec?.nodeName == null) {
            result["unscheduled"].push(p);
          }
          if (!result[p.spec?.nodeName as string]) {
            result[p.spec?.nodeName as string] = [p];
          } else {
            result[p.spec?.nodeName as string].push(p);
          }
        }
      });
      state.pods = result;
    },

    async createPod(name: string) {
      await applyPod({
        metadata: {
          name: name,
        },
      });
      await this.getPods();
    },
    async editPod(p: V1Pod) {
      await applyPod(p);
      await this.getPods();
    },
  };
}

export type PodStore = ReturnType<typeof podStore>;
