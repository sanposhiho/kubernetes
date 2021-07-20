<template>
<v-expansion-panels accordion multiple>
        <v-expansion-panel v-if="filterTableData.length > 1">
          <v-expansion-panel-header> Filter </v-expansion-panel-header>
          <v-expansion-panel-content>
            <v-data-table
              dense
              :headers="filterTableHeader"
              :items="filterTableData"
              item-key="Node"
            >
            </v-data-table>
          </v-expansion-panel-content>
        </v-expansion-panel>
        <v-expansion-panel v-if="scoreTableData.length > 1">
          <v-expansion-panel-header> Score </v-expansion-panel-header>
          <v-expansion-panel-content>
            <v-data-table
              dense
              :headers="scoreTableHeader"
              :items="scoreTableData"
              item-key="Node"
            >
            </v-data-table>
          </v-expansion-panel-content>
        </v-expansion-panel>
        <v-expansion-panel v-if="normalizedTableData.length > 1">
          <v-expansion-panel-header> Normalize Score </v-expansion-panel-header>
          <v-expansion-panel-content>
            <v-data-table
              dense
              :headers="normalizedTableHeader"
              :items="normalizedTableData"
              item-key="Node"
            >
            </v-data-table>
          </v-expansion-panel-content>
        </v-expansion-panel>
      </v-expansion-panels>
</template>
<script lang="ts">
import {
  ref,
  defineComponent,
  inject,
  computed,
  watch,
} from "@nuxtjs/composition-api";
import { extractTableHeader, schedulingResultToTableData } from "../lib/util";
import PodStoreKey from "../StoreKey/PodStoreKey";

export default defineComponent({
  setup() {
    const podstore = inject(PodStoreKey);
    if (!podstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    // scheduling results
    const filterTableHeader = ref(
      [] as Array<{
        text: string;
        value: string;
      }>
    );
    const filterTableData = ref(
      [] as Array<{ [name: string]: string | number }>
    );
    const scoreTableHeader = ref(
      [] as Array<{
        text: string;
        value: string;
      }>
    );
    const scoreTableData = ref(
      [] as Array<{ [name: string]: string | number }>
    );
    const normalizedTableHeader = ref(
      [] as Array<{
        text: string;
        value: string;
      }>
    );
    const normalizedTableData = ref(
      [] as Array<{ [name: string]: string | number }>
    );

    const pod = computed(() => podstore.selected);
    watch(pod, () => {
      if (pod.value?.item.metadata?.annotations) {
        if (
          "scheduler-simulator/score-result" in
          pod.value.item.metadata.annotations
        ) {
          var score = JSON.parse(
            pod.value?.item.metadata?.annotations[
              "scheduler-simulator/score-result"
            ]
          );
        }
        if (
          "scheduler-simulator/normalizedscore-result" in
          pod.value.item.metadata.annotations
        ) {
          var nscore = JSON.parse(
            pod.value?.item.metadata?.annotations[
              "scheduler-simulator/normalizedscore-result"
            ]
          );
        }
        if (
          "scheduler-simulator/filter-result" in
          pod.value.item.metadata.annotations
        ) {
          var filter = JSON.parse(
            pod.value?.item.metadata?.annotations[
              "scheduler-simulator/filter-result"
            ]
          );
        }

        filterTableHeader.value = extractTableHeader(filter);
        filterTableData.value = schedulingResultToTableData(filter);
        scoreTableHeader.value = extractTableHeader(score);
        scoreTableData.value = schedulingResultToTableData(score);
        normalizedTableHeader.value = extractTableHeader(nscore);
        normalizedTableData.value = schedulingResultToTableData(nscore);
      }
    });

    return {
      filterTableHeader,
      filterTableData,
      scoreTableHeader,
      scoreTableData,
      normalizedTableHeader,
      normalizedTableData,
    };
  },
});
</script>