<template>
  <v-content>
    <v-container fill-height fluid>
      <v-layout align-center justify-center>
        <v-flex md4>
          <v-card>
            <v-card-text>
              <v-form>
                <v-text-field
                  :error="errors.hasErrors('username')"
                  :error-messages="errors.getErrors('username')"
                  :readonly="loading"
                  label="Username"
                  name="username"
                  type="text"
                  v-model="username"
                  v-on:keyup.enter="submit"/>
                <v-text-field
                  :error="errors.hasErrors('password')"
                  :error-messages="errors.getErrors('password')"
                  :readonly="loading"
                  :type="showPassword ? 'text' : 'password'"
                  @click:append="toggleShowPassword"
                  label="Password"
                  v-model="password"
                  v-on:keyup.enter="submit"/>
              </v-form>
              <v-alert :value="errors.hasErrors('message')" outlined
                       transition="message-transition"
                       type="error">{{ errors.getErrors('message') }}
              </v-alert>
            </v-card-text>
            <v-card-actions>
              <v-btn :loading="loading" @click="submit" block outlined>Login</v-btn>
            </v-card-actions>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </v-content>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import { AxiosResponse } from 'axios';
import { UnexpectedStatusError, ValidationError } from '@/utils/api';
import apiService from '@/apiService';


  @Component
export default class Login extends Vue {
    private username: string = '';

    private password: string = '';

    private loading: boolean = false;

    private errors: ValidationError = new ValidationError({});

    private showPassword: boolean = false;

    private toggleShowPassword() {
      this.showPassword = !this.showPassword;
    }

    private async login(username: string, password: string) {
      const payload = {
        username,
        password,
      };

      let resp: AxiosResponse;

      resp = await apiService.post('auth/login', payload);

      if (resp.status === 400) {
        throw new ValidationError(resp.data);
      } else if (resp.status === 200) {

      } else {
        throw new UnexpectedStatusError(resp.status);
      }
    }

    private async submit() {
      this.loading = true;
      try {
        await this.login(this.username, this.password);
        this.password = '';

        let path = '/';
        if (this.$route.query.next !== undefined) {
          path = this.$route.query.next[0] || '/';
        }
        this.$router.push({ path });
      } catch (e) {
        if (e instanceof ValidationError) {
          this.errors = e;
        } else {
          this.$router.push({ name: 'error' });
        }
      }
      this.loading = false;
    }
}
</script>

<style lang="scss" scoped>
  .error-spacer {
    padding: 10px;
  }
</style>
