<template>
  <v-form v-model="valid" method="POST" action="http://127.0.0.1:3000/signin">
    <v-row justify="center">
      <v-col cols="4">
        <v-text-field
          v-model="name"
          :rules="nameRules"
          name="username"
          placeholder="Username"
          required
        ></v-text-field>

        <v-text-field
          v-model="password"
          placeholder="Password"
          :append-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
          :rules="[rules.required, rules.min]"
          :type="show1 ? 'text' : 'password'"
          name="password"
          counter
          @click:append="show1 = !show1"
        ></v-text-field>

        <v-alert v-if="wrongDetails" dense outlined type="error">
          <strong>Invalid login credential, please try again.</strong>
        </v-alert>
      </v-col>
    </v-row>

    <v-btn outlined v-on:click="attemptLogin" :disabled="buttonDisabled">Sign In</v-btn>
  </v-form>
</template>

<script>
import axios from "axios";

export default {
  data: () => ({
    info: null,
    wrongDetails: false,
    buttonDisabled: false,
    valid: true,
    name: "",
    nameRules: [v => !!v || "Name is required"],
    email: "",
    emailRules: [
      v => !!v || "E-mail is required",
      v => /.+@.+\..+/.test(v) || "E-mail must be valid"
    ],
    password: "",
    show1: false,
    rules: {
      required: value => !!value || "Required.",
      emailMatch: () => "The email and password you entered don't match"
    }
  }),

  methods: {
    validate() {
      if (this.$refs.form.validate()) {
        this.snackbar = true;
      }
    },
    reset() {
      this.$refs.form.reset();
    },
    resetValidation() {
      this.$refs.form.resetValidation();
    },
    attemptLogin() {
      this.buttonDisabled = true;
      var login_data = {
        username: this.name,
        password: this.password
      };

      axios.post("http://127.0.0.1:3000/signin", login_data).then(
        response => {
          // eslint-disable-next-line no-console
          console.log(response.statusText);
          if (response.status == 200) {
            this.wrongDetails = false;
          }
        },
        error => {
          if (error.response.status == 422) {
            this.buttonDisabled = false;
            this.wrongDetails = true;
          }
        }
      );
    }
  }

  // mounted() {
  //   axios
  //     .get("https://api.coindesk.com/v1/bpi/currentprice.json")
  //     .then(response => (this.info = response));
  // }
};
</script>
