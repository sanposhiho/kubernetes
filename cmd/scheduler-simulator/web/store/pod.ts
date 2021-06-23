import { reactive } from "@nuxtjs/composition-api";
import { createPod, listPod } from "~/api/v1/pod";
import { V1Pod, V1PodList } from "@kubernetes/client-node";

type stateType = {
  count: number;
  pods: {
    [key: string]: Array<V1Pod>;
  };
};

export default function podStore() {
  const state: stateType = reactive({
    count: 0,
    pods: { unscheduled: [] },
  });

  return {
    get count() {
      return state.count;
    },

    get pods() {
      console.log(state.pods);
      return state.pods;
    },

    increment() {
      state.count += 1;
    },

    decrement() {
      state.count -= 1;
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

    async createNewPod(name: string) {
      await createPod({
        metadata: {
          name: name,
        },
      });
      await this.getPods();
    },
  };
}

export type PodStore = ReturnType<typeof podStore>;
