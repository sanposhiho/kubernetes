<template>
  <v-dialog v-model="dialog" width="500">
    <template v-slot:activator="{ on, attrs }">
      <v-btn
        color="primary ma-2"
        dark
        v-bind="attrs"
        v-on="on"
        v-for="(rn, i) in resourceNames"
        @click="target = rn"
        :key="i"
      >
        New {{ rn }}
      </v-btn>
    </template>

    <v-card>
      <v-card-title class="text-h5 grey lighten-2">
        Create New {{ target }}
      </v-card-title>
      <v-divider></v-divider>
      <v-divider></v-divider>
      <v-card-subtitle> Choose the option below to create. </v-card-subtitle>

      <template>
        <v-btn @click="create" block class="pa-2">Create with template</v-btn>
        <v-spacer />
        <v-btn disabled block class="pa-2">Create with local file</v-btn>
        <v-spacer />
      </template>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { ref, watch, inject, defineComponent } from "@nuxtjs/composition-api";
import { podTemplate, nodeTemplate, pvTemplate, pvcTemplate, storageclassTemplate } from "./lib/template";
import { getSimulatorIDFromPath } from "./lib/util";
import PodStoreKey from "./PodStoreKey";
import NodeStoreKey from "./NodeStoreKey";
import PersistentVolumeStoreKey from "./PVStoreKey";
import PersistentVolumeClaimStoreKey from "./PVCStoreKey";
import StorageClassStoreKey from "./StorageClassStoreKey";
import {
  V1Node,
  V1PersistentVolumeClaim,
  V1PersistentVolume,
  V1Pod,
  V1StorageClass,
} from "@kubernetes/client-node";

type Resource =
  | V1Pod
  | V1Node
  | V1PersistentVolumeClaim
  | V1PersistentVolume
  | V1StorageClass;

interface Store {
  readonly selected: object | null;
  select(name: Resource | null, isNew: boolean): void;
}

export default defineComponent({
  setup(_, context) {
    var store: Store | null = null;

    const podstore = inject(PodStoreKey);
    if (!podstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const nodestore = inject(NodeStoreKey);
    if (!nodestore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const pvstore = inject(PersistentVolumeStoreKey);
    if (!pvstore) {
      throw new Error(`${pvstore} is not provided`);
    }

    const pvcstore = inject(PersistentVolumeClaimStoreKey);
    if (!pvcstore) {
      throw new Error(`${pvcstore} is not provided`);
    }

    const storageclassstore = inject(StorageClassStoreKey);
    if (!storageclassstore) {
      throw new Error(`${StorageClassStoreKey} is not provided`);
    }

    const dialog = ref(false);
    const resourceNames = [
      "StorageClass",
      "PersistentVolumeClaim",
      "PersistentVolume",
      "Node",
      "Pod",
    ];

    const targetTemplate = ref(null as Resource | null);
    const target = ref("");

    watch(target, () => {
      const route = context.root.$route;
      switch (target.value) {
        case "Pod":
          targetTemplate.value = podTemplate(
            getSimulatorIDFromPath(route.path)
          );
          store = podstore;
          break;
        case "Node":
          targetTemplate.value = nodeTemplate();
          store = nodestore;
          break;
        case "PersistentVolume":
          targetTemplate.value = pvTemplate();
          store = pvstore;
          break;
        case "PersistentVolumeClaim":
          targetTemplate.value = pvcTemplate();
          store = pvcstore;
          break;
        case "StorageClass":
          targetTemplate.value = storageclassTemplate();
          store = storageclassstore;
          break;
      }
    });

    const create = () => {
      if (store) {
        store.select(targetTemplate.value, true);
      }
      dialog.value = false;
    };

    return {
      create,
      dialog,
      resourceNames,
      target,
    };
  },
});
</script>
