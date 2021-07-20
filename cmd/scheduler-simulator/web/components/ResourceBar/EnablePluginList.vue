<template>
  <v-expansion-panels accordion multiple>
    <v-expansion-panel>
      <v-expansion-panel-header>
        enable filter plugins
      </v-expansion-panel-header>
      <v-expansion-panel-content>
        <v-row>
          <v-col v-for="plugin in filterplugins" :key="plugin.name" cols="auto">
            <v-checkbox
              v-model="selectedplugins"
              :label="plugin.name"
              :value="plugin.name"
              @change="onChange"
            ></v-checkbox>
          </v-col>
        </v-row>
      </v-expansion-panel-content>
    </v-expansion-panel>
    <v-expansion-panel>
      <v-expansion-panel-header>
        enable score plugins
      </v-expansion-panel-header>
      <v-expansion-panel-content>
        <v-row>
          <v-col
            v-for="plugin in scoreplugins"
            :key="plugin.name"
            :value="plugin.name"
            cols="auto"
          >
            <v-checkbox
              v-model="selectedplugins"
              :label="plugin.name"
              :value="plugin.name"
              @change="onChange"
            ></v-checkbox>
          </v-col>
        </v-row>
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-expansion-panels>
</template>
>
<script lang="ts">
import { defineComponent, ref } from "@nuxtjs/composition-api";
import { filterPlugins, scorePlugins } from "~/components/lib/plugins";

export default defineComponent({
  props: {
    value: Array,
  },
  setup(props, { emit }) {
    const filterplugins = filterPlugins.map((p) => {
      return {
        name: p,
        enabled: true,
      };
    });
    const scoreplugins = scorePlugins.map((p) => {
      return {
        name: p,
        enabled: true,
      };
    });
    const selectedplugins = ref(props.value);

    const onChange = () => {
      emit("input", selectedplugins.value);
    };

    return {
      scoreplugins,
      filterplugins,
      selectedplugins,
      onChange,
    };
  },
});
</script>
