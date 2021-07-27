import { reactive } from "@nuxtjs/composition-api";
import {
  applyPersistentVolumeClaim,
  deletePersistentVolumeClaim,
  getPersistentVolumeClaim,
  listPersistentVolumeClaim,
} from "~/api/v1/pvc";
import {
  V1PersistentVolumeClaim,
  V1PersistentVolumeClaimList,
  V1Pod,
  V1PodList,
} from "@kubernetes/client-node";

type stateType = {
  selectedPersistentVolumeClaim: selectedPersistentVolumeClaim | null;
  pvcs: V1PersistentVolumeClaim[];
};

type selectedPersistentVolumeClaim = {
  // isNew represents whether this is a new PersistentVolumeClaim or not.
  isNew: boolean;
  item: V1PersistentVolumeClaim;
  resourceKind: string;
};

export default function pvcStore() {
  const state: stateType = reactive({
    selectedPersistentVolumeClaim: null,
    pvcs: [],
  });

  return {
    get pvcs() {
      return state.pvcs;
    },

    get count(): number {
      return state.pvcs.length;
    },

    get selected() {
      return state.selectedPersistentVolumeClaim;
    },

    select(n: V1PersistentVolumeClaim | null, isNew: boolean) {
      if (n !== null) {
        state.selectedPersistentVolumeClaim = {
          isNew: isNew,
          item: n,
          resourceKind: "PVC",
        };
      }
    },

    resetSelected() {
      state.selectedPersistentVolumeClaim = null;
    },

    async fetchlist(simulatorID: string) {
      state.pvcs = (await listPersistentVolumeClaim(simulatorID)).items;
    },

    async apply(n: V1PersistentVolumeClaim, simulatorID: string) {
      await applyPersistentVolumeClaim(n, simulatorID);
      await this.fetchlist(simulatorID);
    },

    async fetchSelected(simulatorID: string) {
      if (
        state.selectedPersistentVolumeClaim?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        state.selectedPersistentVolumeClaim.item =
          await getPersistentVolumeClaim(
            state.selectedPersistentVolumeClaim.item.metadata.name,
            simulatorID
          );
      }
    },

    async delete(name: string, simulatorID: string) {
      await deletePersistentVolumeClaim(name, simulatorID);
      await this.fetchlist(simulatorID);
    },
  };
}

export type PersistentVolumeClaimStore = ReturnType<typeof pvcStore>;
