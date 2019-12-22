<template>
  <v-app>
    <v-content justify="center">
      <img alt="Vue logo" src="./assets/logo.jpg" />
      <SignIn />
    </v-content>
  </v-app>
</template>

<script>
import SignIn from "./components/SignIn.vue";

export default {
  name: "app",
  components: {
    SignIn
  }
};
</script>

<style>
#app {
  font-family: "Avenir", Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
