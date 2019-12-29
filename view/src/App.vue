<template>
  <v-app>
    <v-content>
      <v-row>
        <v-spacer></v-spacer>
        <img class="justify-center" alt="UNILAG logo" src="./assets/logo.jpg" />
        <v-spacer></v-spacer>
      </v-row>

      <router-view></router-view>
    </v-content>
  </v-app>
</template>

<script>
export default {
  name: "app",
  components: {}
};
</script>