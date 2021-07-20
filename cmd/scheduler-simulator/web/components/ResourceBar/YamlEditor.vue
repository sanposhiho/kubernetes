<template>
      <monaco-editor
        v-model="formData"
        class="editor mt-1"
        language="yaml"
        @change="onChange"
      ></monaco-editor>
</template>

<script lang="ts">
import {
  defineComponent,
  ref,
} from "@nuxtjs/composition-api";
//@ts-ignore // it is ok to ignore.
import MonacoEditor from "vue-monaco";

export default defineComponent({
components: {
    MonacoEditor,
},
props: {
    value: Array,
  },
setup(props, { emit } ) {
    const formData = ref(props.value)

    const onChange = () => {
        emit("input", formData.value)
    }

    return {
        formData,
      onChange,
    };
  },
});
</script>
