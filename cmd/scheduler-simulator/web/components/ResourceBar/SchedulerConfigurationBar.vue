<template>
  <v-navigation-drawer absolute right temporary bottom width="70%" v-model="d">
    <BarHeader
      title="Scheduler Configuration"
      :applyOnClick="applyOnClick"
      :enableDeleteBtn="false"
      :enableEditmodeSwitch="false"
    />

    <v-divider></v-divider>

    <v-spacer v-for="n in 3" :key="n" />
    <v-divider></v-divider>

    <YamlEditor v-model="formData" />
  </v-navigation-drawer>
</template>

<script lang="ts">
import { ref, watch, defineComponent, computed } from "@nuxtjs/composition-api";
import yaml from "js-yaml";
import { getSimulatorIDFromPath } from "../lib/util";
import YamlEditor from "./YamlEditor.vue";
import SchedulingResults from "./SchedulingResults.vue";
import ResourceDefinitionTree from "./DefinitionTree.vue";
import BarHeader from "./BarHeader.vue";
import {
  applySchedulerConfiguration,
  getSchedulerConfiguration,
} from "~/api/v1/schedulerconfiguration";
import { SchedulerConfiguration } from "~/api/v1/types";

export default defineComponent({
  components: {
    YamlEditor,
    BarHeader,
    ResourceDefinitionTree,
    SchedulingResults,
  },
  props: {
    value: Boolean,
  },
  setup(props, context) {
    const formData = ref("");

    const route = context.root.$route;

    const fetch = async () => {
      getSchedulerConfiguration(getSimulatorIDFromPath(route.path)).then(
        (value: SchedulerConfiguration) => {
          formData.value = yaml.dump(value);
          d.value = props.value;
        }
      );
    };

    const applyOnClick = async () => {
      const cfg = yaml.load(formData.value);
      applySchedulerConfiguration(cfg, getSimulatorIDFromPath(route.path));
      d.value = false;
    };

    const d = ref(false);

    watch(props, (newvalue, _) => {
      if (newvalue) {
        fetch();
      }
    });

    watch(d, () => {
      context.emit("input", d.value);
    });

    return {
      formData,
      fetch,
      applyOnClick,
      d,
    };
  },
});
</script>
