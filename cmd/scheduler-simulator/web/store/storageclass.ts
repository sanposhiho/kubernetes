import { reactive } from "@nuxtjs/composition-api";
import {
  applyStorageClass,
  deleteStorageClass,
  getStorageClass,
  listStorageClass,
} from "~/api/v1/storageclass";
import {
  V1StorageClass,
  V1StorageClassList,
  V1Pod,
  V1PodList,
} from "@kubernetes/client-node";

type stateType = {
  selectedStorageClass: selectedStorageClass | null;
  storageclasses: V1StorageClass[];
};

type selectedStorageClass = {
  // isNew represents whether this is a new StorageClass or not.
  isNew: boolean;
  item: V1StorageClass;
  resourceKind: string;
};

export default function storageclassStore() {
  const state: stateType = reactive({
    selectedStorageClass: null,
    storageclasses: [],
  });

  return {
    get storageclasses() {
      return state.storageclasses;
    },

    get selected() {
      return state.selectedStorageClass;
    },

    select(n: V1StorageClass | null, isNew: boolean) {
      if (n !== null) {
        state.selectedStorageClass = {
          isNew: isNew,
          item: n,
          resourceKind: "SC",
        };
      }
    },

    resetSelected() {
      state.selectedStorageClass = null;
    },

    async fetchlist(simulatorID: string) {
      state.storageclasses = (await listStorageClass(simulatorID)).items;
    },

    async apply(n: V1StorageClass, simulatorID: string) {
      await applyStorageClass(n, simulatorID);
      await this.fetchlist(simulatorID);
    },

    async fetchSelected(simulatorID: string) {
      if (
        state.selectedStorageClass?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        state.selectedStorageClass.item = await getStorageClass(
          state.selectedStorageClass.item.metadata.name,
          simulatorID
        );
      }
    },

    async delete(name: string, simulatorID: string) {
      await deleteStorageClass(name, simulatorID);
      await this.fetchlist(simulatorID);
    },
  };
}

export type StorageClassStore = ReturnType<typeof storageclassStore>;
