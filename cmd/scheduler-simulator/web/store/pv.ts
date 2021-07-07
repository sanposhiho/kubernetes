import { reactive } from "@nuxtjs/composition-api";
import {
  applyPersistentVolume,
  deletePersistentVolume,
  listPersistentVolume,
} from "~/api/v1/pv";
import {
  V1PersistentVolume,
  V1PersistentVolumeList,
  V1Pod,
  V1PodList,
} from "@kubernetes/client-node";

type stateType = {
  selectedPersistentVolume: selectedPersistentVolume | null;
  pvs: V1PersistentVolume[];
};

type selectedPersistentVolume = {
  // isNew represents whether this Pod is a new PersistentVolume or not.
  isNew: boolean;
  item: V1PersistentVolume;
};

export default function pvStore() {
  const state: stateType = reactive({
    selectedPersistentVolume: null,
    pvs: [],
  });

  return {
    get pvs() {
      return state.pvs;
    },

    get selected() {
      return state.selectedPersistentVolume;
    },

    select(n: V1PersistentVolume | null, isNew: boolean) {
      if (n !== null) {
        state.selectedPersistentVolume = {
          isNew: isNew,
          item: n,
        };
      }
    },

    resetSelected() {
      state.selectedPersistentVolume = null;
    },

    async list(simulatorID: string) {
      state.pvs = (await listPersistentVolume(simulatorID)).items;
    },

    async apply(n: V1PersistentVolume, simulatorID: string) {
      await applyPersistentVolume(n, simulatorID);
      await this.list(simulatorID);
    },

    async delete(name: string, simulatorID: string) {
      await deletePersistentVolume(name, simulatorID);
      await this.list(simulatorID);
    },
  };
}

export type PersistentVolumeStore = ReturnType<typeof pvStore>;
