<template>
  <v-form ref="form" v-model="valid" lazy-validation>
    <v-row justify="center">
      <v-col cols="4">
        <v-text-field v-model="name" :rules="nameRules" placeholder="Username" required></v-text-field>
        <v-text-field
          v-model="password"
          :append-icon="show1 ? 'mdi-eye' : 'mdi-eye-off'"
          :rules="[rules.required, rules.min]"
          :type="show1 ? 'text' : 'password'"
          name="input-10-1"
          counter
          @click:append="show1 = !show1"
        ></v-text-field>
      </v-col>
    </v-row>
    <v-btn outlined @click="resetValidation">Sign In</v-btn>
  </v-form>
</template>

<script>
export default {
  data: () => ({
    valid: true,
    name: "",
    nameRules: [v => !!v || "Name is required"],
    email: "",
    emailRules: [
      v => !!v || "E-mail is required",
      v => /.+@.+\..+/.test(v) || "E-mail must be valid"
    ],
    password: "Password",
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
    }
  }
};
</script>
