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
        <v-list-item-content >
          <v-list-item-title>Resource</v-list-item-title>
          <v-row>
            <v-col>
          <v-switch class="ma-5 mb-0" v-model="editmode" inset label="edit" />
            </v-col>
            <v-spacer v-for="n in 3" :key="n" />
            <v-col>
          <v-btn class="ma-5 mb-0" @click="applyOnClick"> Apply </v-btn>
            </v-col>
            <v-col>
          <v-btn class="ma-5 mb-0" color="error" @click="deleteOnClick">
            Delete
          </v-btn>
            </v-col>
          </v-row>
        </v-list-item-content>
      </v-list-item>
    </template>

    <v-divider></v-divider>

    <template v-if="editmode ">
      <v-textarea filled auto-grow v-model="formData"></v-textarea>
    </template>

    <v-treeview
      dense
      open-all
      ref="tree"
      v-if="!editmode"
      :items="treeData"
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
import yaml from "js-yaml";
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

    const formData = ref("");

    // value for the treeview
    const tree: any = ref(null);
    const treeData = ref(objectToTreeViewData(null))

    const drawer = ref(false);
    const editmode = ref(false);

    const pod = computed(() => pstore.selectedPod);
    const node = computed(() => nstore.selectedNode);

    watch(pod, () => {
      if (pod.value !== null) {
        editmode.value = pod.value.isNew;

        formData.value = yaml.dump(pod.value.pod);
        treeData.value = objectToTreeViewData(pod.value.pod);
      }
    });

    watch(node, () => {
      if (node.value !== null) {
        editmode.value = node.value.isNew;

        formData.value = yaml.dump(node.value.node);
        treeData.value = objectToTreeViewData(node.value.node);
      }
    });

    watch(treeData, () => {
        // make sidebar visible, after treeData changed to open treeview correctly.
        drawer.value = true;
    })

    watch(drawer, (newValue, _) => {
      if (!newValue) {
        // reset editmode.
        editmode.value = false;
      } else {
        // open all tree when sidebar be visible.
        tree.value.updateAll(true);
      }
    });

    const applyOnClick = () => {
      if (pod.value !== null) {
        pstore.applyPod(yaml.load(formData.value));
      } else if (node.value !== null) {
        nstore.applyNode(yaml.load(formData.value));
      }
      drawer.value = false;
    };

    const deleteOnClick = () => {
      if (pod.value !== null) {
        podDeleteOnClick();
      } else if (node.value !== null) {
        nodeDeleteOnClick();
      }
    };

    const podDeleteOnClick = () => {
      if (pod.value?.pod.metadata?.name) {
        pstore.deletePod(pod.value.pod.metadata.name);
        drawer.value = false;
      }
    };

    const nodeDeleteOnClick = () => {
      if (node.value?.node.metadata?.name) {
        nstore.deleteNode(node.value.node.metadata.name);
        drawer.value = false;
      }
    };

    return {
      drawer,
      editmode,
      tree,
      pod,
      node,
      formData,
      treeData, 
      applyOnClick,
      deleteOnClick,
    };
  },
});
</script>
