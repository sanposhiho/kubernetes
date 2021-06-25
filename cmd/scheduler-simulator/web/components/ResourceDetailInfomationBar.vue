<template>
  <v-navigation-drawer
    absolute
    right
    temporary
    bottom
    width="60%"
    v-model="drawer"
  >
    <template v-slot:prepend>
      <v-list-item two-line>
        <v-list-item-content>
          <v-list-item-title>Resource</v-list-item-title>
          <v-row>
            <v-col>
              <v-switch
                class="ma-5 mb-0"
                v-model="editmode"
                inset
                label="edit"
              />
            </v-col>
            <v-spacer></v-spacer>
            <v-spacer></v-spacer>
            <v-spacer></v-spacer>
            <v-col>
              <v-btn
                class="ma-5 mb-0"
                v-if="editmode && pod != null"
                @click="podFormOnClick"
              >
                Submit
              </v-btn>
              <v-btn
                class="ma-5 mb-0"
                v-if="editmode && node != null"
                @click="nodeFormOnClick"
              >
                Submit
              </v-btn>
            </v-col>
          </v-row>
        </v-list-item-content>
      </v-list-item>
    </template>

    <v-divider></v-divider>

    <template v-if="editmode && pod != null">
      <v-textarea filled auto-grow v-model="formPodJSON"></v-textarea>
    </template>

    <template v-if="editmode && node != null">
      <v-textarea filled auto-grow v-model="formNodeJSON"></v-textarea>
    </template>

    <v-treeview
      dense
      open-all
      ref="podTree"
      v-if="!editmode"
      :items="objectToTreeViewData(pod)"
    ></v-treeview>
    <v-treeview
      dense
      open-all
      v-if="!editmode"
      ref="nodeTree"
      :items="objectToTreeViewData(node)"
    ></v-treeview>
  </v-navigation-drawer>
</template>
<script lang="ts">
import {
  ref,
  computed,
  inject,
  watch,
  defineComponent,
} from "@nuxtjs/composition-api";
import PodStoreKey from "./pod-store-key";
import { objectToTreeViewData } from "./lib/util";
import NodeStoreKey from "./node-store-key";

export default defineComponent({
  setup() {
    const pstore = inject(PodStoreKey);
    if (!pstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const nstore = inject(NodeStoreKey);
    if (!nstore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const formPodJSON = ref("");
    const formNodeJSON = ref("");

    const nodeTree: any = ref(null);
    const podTree: any = ref(null);
    const drawer = ref(false);
    const editmode = ref(false);

    const pod = computed(() => pstore.selectedPod);
    const node = computed(() => nstore.selectedNode);

    watch(pod, () => {
      if (pod.value !== null) {
        // make sidebar visible, only when a pod is selected.
        drawer.value = true;
        nstore.selectNode(null);

        formPodJSON.value = JSON.stringify(pod.value, null, "   ");
      }
    });

    watch(node, () => {
      if (node.value !== null) {
        // make sidebar visible, only when a node is selected.
        drawer.value = true;
        pstore.selectPod(null);

        formNodeJSON.value = JSON.stringify(node.value, null, "   ");
      }
    });

    watch(drawer, (newValue, _) => {
      if (!newValue) {
        // reset pod and node selection when sidebar be invisible.
        pstore.selectPod(null);
        nstore.selectNode(null);
        editmode.value = false;
      } else {
      // open all tree when sidebar be visible.
      podTree.value.updateAll(true);
      nodeTree.value.updateAll(true);
      }
    });

    const podFormOnClick = () => {
      pstore.editPod(JSON.parse(formPodJSON.value));
      drawer.value = false;
    };

    const nodeFormOnClick = () => {
      nstore.editNode(JSON.parse(formNodeJSON.value));
      drawer.value = false;
    };

    return {
      drawer,
      editmode,
      nodeTree,
      podTree,
      pod,
      node,
      formPodJSON,
      formNodeJSON,
      objectToTreeViewData,
      podFormOnClick,
      nodeFormOnClick,
    };
  },
});
</script>
