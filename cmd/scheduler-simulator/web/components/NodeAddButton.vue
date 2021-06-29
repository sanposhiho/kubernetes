<template>
  <v-dialog v-model="dialog" width="500">
    <template v-slot:activator="{ on, attrs }">
      <v-btn color="primary" v-bind="attrs" v-on="on"> Create New Node </v-btn>
    </template>

    <v-card>
      <v-card-title class="text-h5 grey lighten-2">
        Create New Node
      </v-card-title>
      <v-divider></v-divider>
      <v-divider></v-divider>
      <v-card-subtitle> Choose the option below to create. </v-card-subtitle>
      <v-card-text>
        Note: Since the name of the node needs to be unique, including other
        users, it will be automatically given a suffix.</v-card-text
      >

      <template>
        <v-btn @click="createNode" block class="pa-2"
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
import { nodeTemplate, podTemplate } from "./lib/template";
import NodeStoreKey from "./node-store-key";

export default defineComponent({
  setup() {
    const store = inject(NodeStoreKey);
    if (!store) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const dialog = ref(false);

    const createNode = async () => {
      store.selectNode(nodeTemplate(), true);
      dialog.value = false;
    };

    return {
      createNode,
      dialog,
    };
  },
});
</script>
