<template>
  <v-content justify="center">{{queryData}}</v-content>
</template>

<script>
import axios from "axios";

export default {
  data: () => ({
    queryData: ""
  }),

  components: {},

  mounted() {
    // eslint-disable-next-line no-console
    console.log("Hello");
    axios.get("http://127.0.0.1:3000/query").then(response => {
      this.queryData = response.data.results;
    });
  }
};
</script>
