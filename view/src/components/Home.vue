<template>
  <v-container>
    <v-cols>
      <ul>
        <li v-for="(item, key) in queryData " :key="key">
          <p v-if="item.Up==true">{{ key }} - Server Up</p>
          <p v-else>{{ key }} - Server Down</p>
        </li>
      </ul>
      <h2 class="text-center">
        <u>Server Logs</u>
      </h2>
    </v-cols>
  </v-container>
</template>

<script>
import axios from "axios";

export default {
  data: () => ({
    queryData: null
  }),

  components: {},

  mounted() {
    // eslint-disable-next-line no-console
    console.log("Hello");
    axios
      .get("http://127.0.0.1:3000/query", {
        withCredentials: true
      })
      .then(
        response => {
          this.queryData = response.data;
        },

        error => {
          if (error.response.status == 425) {
            // eslint-disable-next-line no-console
            console.log("error so signed out");
            axios
              .get("http://127.0.0.1:3000/logout")
              .then(this.$router.push("/signin"));
          }
        }
      );
  }
};
</script>
