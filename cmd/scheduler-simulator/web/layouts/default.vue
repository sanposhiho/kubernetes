<template>
  <v-app :style="{ background: $vuetify.theme.themes.light.background }">
    <v-app-bar color="primary" app>
      <v-toolbar-title
        v-text="title"
        class="white--text"
        @click="onClick"
        :style="{ cursor: 'pointer' }"
      />
    </v-app-bar>
    <v-main>
      <v-container>
        <nuxt />
      </v-container>
    </v-main>
    <v-footer app>
      <span>&copy; {{ new Date().getFullYear() }}</span>
    </v-footer>
  </v-app>
</template>

<script>
import { defineComponent } from "@nuxtjs/composition-api";

export default defineComponent({
  setup(_, context) {
    const router = context.root.$router;

    const onClick = () => {
      router.push(`/`);
    };
    return {
      onClick,
    };
  },
  data() {
    return {
      title: "Kubernetes scheduler simulator",
    };
  },
});
</script>
