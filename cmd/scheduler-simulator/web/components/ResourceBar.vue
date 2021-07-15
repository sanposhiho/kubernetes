<template>
  <v-navigation-drawer
    absolute
    right
    temporary
    bottom
    width="80%"
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
            <v-spacer v-for="n in 3" :key="n" />
            <v-col>
              <v-btn class="ma-5 mb-0" @click="applyOnClick"> Apply </v-btn>
            </v-col>
            <v-col>
              <ResourceDeleteButton
                @deleteOnClick="deleteOnClick"
                v-if="selected && !selected.isNew"
              />
            </v-col>
          </v-row>
        </v-list-item-content>
      </v-list-item>
    </template>

    <v-divider></v-divider>

    <template v-if="editmode">
      <monaco-editor v-model="formData" class="editor" language="yaml"></monaco-editor>
    </template>

    <template v-if="!editmode">
    <template v-if="podFilterResultTreeData.length != 0 ">
    <v-card-title>Filter</v-card-title>
    <v-treeview
      dense
      v-if="!editmode"
      :items="podFilterResultTreeData"
    ></v-treeview>

    <v-divider></v-divider>
    </template>


    <template v-if="podScoringResultTreeData.length != 0">
    <v-card-title>Score</v-card-title>
    <v-treeview
      dense
      v-if="!editmode"
      :items="podScoringResultTreeData"
    ></v-treeview>

    <v-divider></v-divider>
    </template>

    <template v-if="podNormalizedScoringResultTreeData.length != 0">
    <v-card-title>Normalize Score</v-card-title>
    <v-treeview
      dense
      v-if="!editmode"
      :items="podNormalizedScoringResultTreeData"
    ></v-treeview>

    <v-divider></v-divider>
    </template>


    <v-card-title>Resource Definition</v-card-title>
    <v-treeview
      dense
      open-all
      ref="tree"
      v-if="!editmode"
      :items="treeData"
    ></v-treeview>
    </template>
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
//@ts-ignore // it is ok to ignore.
import MonacoEditor from "vue-monaco";
import yaml from "js-yaml";
import PodStoreKey from "./PodStoreKey";
import { getSimulatorIDFromPath, objectToTreeViewData } from "./lib/util";
import NodeStoreKey from "./NodeStoreKey";
import PersistentVolumeStoreKey from "./PVStoreKey";
import PersistentVolumeClaimStoreKey from "./PVCStoreKey";
import StorageClassStoreKey from "./StorageClassStoreKey";
import ResourceDeleteButton from "~/components/ResourceDeleteButton.vue";
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
  resetSelected(): void;
  apply(r: Resource, simulatorID: string): Promise<void>;
  delete(name: string, simulatorID: string): Promise<void>;
}

interface SelectedItem {
  isNew: boolean;
  item: Resource;
}

export default defineComponent({
  components: {
    MonacoEditor,
    ResourceDeleteButton,
  },
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

    const selected = ref(null as SelectedItem | null);

    const formData = ref("");

    const tree: any = ref(null);
    const treeData = ref(objectToTreeViewData(null));
    const podScoringResultTreeData = ref(objectToTreeViewData(null));
    const podNormalizedScoringResultTreeData = ref(objectToTreeViewData(null));
    const podFilterResultTreeData = ref(objectToTreeViewData(null));

    const drawer = ref(false);
    const editmode = ref(false);
    const dialog = ref(false);

    const pod = computed(() => podstore.selected);
    watch(pod, () => {
      store = podstore;
      selected.value = pod.value;
      if (pod.value?.item.metadata?.annotations) {
        var score = JSON.parse(pod.value?.item.metadata?.annotations["scheduler-simulator/score-result"])
        var nscore = JSON.parse(pod.value?.item.metadata?.annotations["scheduler-simulator/normalizedscore-result"])
        var filter = JSON.parse(pod.value?.item.metadata?.annotations["scheduler-simulator/filter-result"])
        podScoringResultTreeData.value = objectToTreeViewData(score);
        podNormalizedScoringResultTreeData.value = objectToTreeViewData(nscore);
        podFilterResultTreeData.value = objectToTreeViewData(filter);
      } else {
        podScoringResultTreeData.value = objectToTreeViewData(null);
        podNormalizedScoringResultTreeData.value = objectToTreeViewData(null);
        podFilterResultTreeData.value = objectToTreeViewData(null);
      }
    });

    const node = computed(() => nodestore.selected);
    watch(node, () => {
      store = nodestore;
      selected.value = node.value;
    });

    const pv = computed(() => pvstore.selected);
    watch(pv, () => {
      store = pvstore;
      selected.value = pv.value;
    });

    const pvc = computed(() => pvcstore.selected);
    watch(pvc, () => {
      store = pvcstore;
      selected.value = pvc.value;
    });

    const sc = computed(() => storageclassstore.selected);
    watch(sc, () => {
      store = storageclassstore;
      selected.value = sc.value;
    });

    watch(selected, () => {
      if (selected.value) {
        editmode.value = selected.value.isNew;

        formData.value = yaml.dump(selected.value.item);
        treeData.value = objectToTreeViewData(selected.value.item);
      }
    });

    watch(treeData, () => {
      // make sidebar visible, after treeData changed to open treeview correctly.
      drawer.value = true;
    });

    watch(drawer, (newValue, _) => {
      if (!newValue) {
        // reset editmode.
        editmode.value = false;
        if (store) {
          store.resetSelected();
        }
        store = null;
        selected.value = null;
      } else {
        // open all tree when sidebar be visible.
        if (tree.value) {
          tree.value.updateAll(true);
        }
      }
    });

    const route = context.root.$route;

    const applyOnClick = () => {
      if (store) {
        store.apply(
          yaml.load(formData.value),
          getSimulatorIDFromPath(route.path)
        );
      }
      drawer.value = false;
    };

    const deleteOnClick = () => {
      if (selected.value?.item.metadata?.name && store) {
        store.delete(
          selected.value.item.metadata.name,
          getSimulatorIDFromPath(route.path)
        );
      }
      drawer.value = false;
    };

    return {
      drawer,
      editmode,
      dialog,
      tree,
      selected,
      formData,
      treeData,
      applyOnClick,
      podScoringResultTreeData,
      podFilterResultTreeData,
      podNormalizedScoringResultTreeData,
      deleteOnClick,
    };
  },
});
</script>

<style>
.editor {
  width: auto;
  height: 100%;
}
</style>