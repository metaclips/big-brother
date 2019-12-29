<template>
  <v-form>
    <v-row justify="center">
      <v-col cols="4">
        <v-text-field
          v-model="name"
          :rules="nameRules"
          id="username"
          placeholder="Username"
          required
        ></v-text-field>

        <v-text-field
          v-model="password"
          placeholder="Password"
          :append-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
          :rules="[rules.required, rules.min]"
          :type="show1 ? 'text' : 'password'"
          id="password"
          counter
          @click:append="show1 = !show1"
        ></v-text-field>

        <v-alert v-if="wrongDetails" dense outlined type="error">
          <strong>Invalid login credential, please try again.</strong>
        </v-alert>
        <v-alert v-if="internalError" dense outlined type="error">
          <strong>Internal error</strong>
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
    internalError: false,
    buttonDisabled: false,
    //valid: true,
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
      // eslint-disable-next-line no-console
      console.log("rrrrr");
      this.buttonDisabled = true;

      var bodyFormData = new FormData();
      bodyFormData.append("password", this.password);
      bodyFormData.append("username", this.name);
      var querystring = require("querystring");

      axios
        .post(
          "http://127.0.0.1:3000/signin",
          querystring.stringify({
            username: this.name,
            password: this.password
          }),
          {
            withCredentials: true
          }
        )
        .then(
          response => {
            // eslint-disable-next-line no-console
            console.log(response.statusText);
            if (response.status == 200) {
              this.wrongDetails = false;
              this.$router.push("/");
            }
          },
          error => {
            if (error.response.status == 422) {
              this.buttonDisabled = false;
              this.wrongDetails = true;
            } else {
              this.internalError = true;
              this.wrongDetails = false;
            }
          }
        );
    }
  },

  mounted() {
    axios
      .get("http://127.0.0.1:3000/islogged", {
        withCredentials: true
      })
      .then(response => {
        // eslint-disable-next-line no-console
        console.log("Trying to check log in");
        if (response.status == 200) {
          // eslint-disable-next-line no-console
          console.log("Trying log in");
          this.$router.push("/");
        }
      });
  }
};
</script>
