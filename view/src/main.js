import Vue from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify';
import VueRouter from 'vue-router';
import SignIn from './components/SignIn.vue';
import Home from './components/Home.vue';

Vue.config.productionTip = false
Vue.use(VueRouter)

const routes = [
  { path: '/signin', component: SignIn },
  { path: '/', component: Home }
]

const router = new VueRouter({
  routes,
  mode: 'history'
})

new Vue({
  vuetify,
  router,
  render: h => h(App)
}).$mount('#app')
