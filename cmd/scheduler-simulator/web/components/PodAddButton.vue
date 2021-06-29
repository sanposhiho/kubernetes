<template>
  <v-dialog v-model="dialog" width="500">
    <template v-slot:activator="{ on, attrs }">
      <v-btn color="primary" dark v-bind="attrs" v-on="on">
        Create New Pod
      </v-btn>
    </template>

    <v-card>
      <v-card-title class="text-h5 grey lighten-2">
        Create New Pod
      </v-card-title>
      <v-divider></v-divider>
      <v-divider></v-divider>
      <v-card-subtitle> Choose the option below to create. </v-card-subtitle>

      <template>
        <v-btn @click="createPod" block class="pa-2"
          >Create with template</v-btn
        >
        <v-spacer />
        <v-btn disabled block class="pa-2">Create with local file</v-btn>
        <v-spacer />
      </template>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import {
  ref,
  computed,
  inject,
  defineComponent,
} from "@nuxtjs/composition-api";
import { podTemplate } from "./lib/template";
import { getSimulatorIDFromPath } from "./lib/util";
import PodStoreKey from "./pod-store-key";

export default defineComponent({
  setup(_, context) {
    const store = inject(PodStoreKey);
    if (!store) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const dialog = ref(false);

    const createPod = async () => {
      const route = context.root.$route;
      store.selectPod(podTemplate(getSimulatorIDFromPath(route.path)), true);
      dialog.value = false;
    };

    return {
      createPod,
      dialog,
    };
  },
});
</script>
