<script >


export default {
  data() {
    return {
      ui: {}
    }
  },

  methods: {
    getUI() {
      const url = 'http://localhost:8080/api/v1/ui/page'+this.$route.path
      console.log(url)
      fetch(url)
          .then(response => response.json())
          .then(data => this.ui = data)
    }
  },
  mounted() {
    this.getUI()
  }
}


</script>

<template>
  <h2>!{{ ui.value }}!</h2>
  <button @click="getUI">You clicked me {{ count }} times.</button>
</template>