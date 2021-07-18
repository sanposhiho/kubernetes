<template>
  <v-navigation-drawer
    absolute
    right
    temporary
    bottom
    width="70%"
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
      
      <EnablePluginList v-model="selectedplugins" />

      <v-spacer v-for="n in 3" :key="n" />
      <v-divider></v-divider>
      <monaco-editor
        v-model="formData"
        class="editor mt-1"
        language="yaml"
      ></monaco-editor>
    </template>

    <template v-if="!editmode">

      <SchedulingResults />
      
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
import {
  getSimulatorIDFromPath,
  objectToTreeViewData,
} from "./lib/util";
import NodeStoreKey from "./NodeStoreKey";
import PersistentVolumeStoreKey from "./PVStoreKey";
import PersistentVolumeClaimStoreKey from "./PVCStoreKey";
import StorageClassStoreKey from "./StorageClassStoreKey";
import EnablePluginList from "~/components/EnablePluginList.vue"
import ResourceDeleteButton from "~/components/ResourceDeleteButton.vue";
import SchedulingResults from "~/components/SchedulingResults.vue"
import { filterPlugins, scorePlugins } from "~/components/lib/plugins";
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
  fetchSelected(simulatorID: string): Promise<void>;
}

interface SelectedItem {
  isNew: boolean;
  item: Resource;
}

export default defineComponent({
  components: {
    EnablePluginList,
    MonacoEditor,
    ResourceDeleteButton,
    SchedulingResults,
  },
  setup(_, context) {
    var store: Store | null = null;

    // inject stores
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

    const selectedplugins = ref(filterPlugins.concat(scorePlugins));

    // boolean to switch some view
    const drawer = ref(false);
    const editmode = ref(false);
    const dialog = ref(false);

    const pod = computed(() => podstore.selected);
    watch(pod, () => {
      store = podstore;
      selected.value = pod.value;
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

    watch(selected, (newVal, oldVal) => {
      if (selected.value) {
        if (!oldVal) {
          // first change
          fetchSelected();
        }
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

    const fetchSelected = async () => {
      if (store) {
        await store.fetchSelected(getSimulatorIDFromPath(route.path));
      }
    };

    const applyOnClick = () => {
      if (store) {
        const y = yaml.load(formData.value);
        y.metadata.annotations = {};
        y.metadata.annotations["scheduler-simulator/enabled-plugins"] =
          JSON.stringify(selectedplugins.value);
        store.apply(y, getSimulatorIDFromPath(route.path));
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
      deleteOnClick,
      selectedplugins,
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
