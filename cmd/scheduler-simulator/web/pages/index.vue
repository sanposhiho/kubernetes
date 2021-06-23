<template>
  <v-row justify="center" align="center">
    <v-col cols="12" sm="8" md="6">
      <div class="text-center">
        <kubernetes-logo />
      </div>
      <v-card>
        <v-card-title class="headline">
          Welcome to Kubernetes scheduler simulator
        </v-card-title>
        <v-card-text>
          <p>Now, it is under developping...</p>
          <p>
            For more information, check out
            <a
              href="https://github.com/kubernetes/kubernetes/issues/99605"
              target="_blank"
              rel="noopener noreferrer"
            >
              the issue</a
            >.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <start-simulator-button />
          <v-spacer />
        </v-card-actions>
      </v-card>
    </v-col>
  </v-row>
</template>

<script>
import KubernetesLogo from "~/components/KubernetesLogo.vue";
import StartSimulatorButton from "~/components/StartSimulatorButton.vue";

export default {
  components: {
    KubernetesLogo,
    StartSimulatorButton,
  },
};
</script>
