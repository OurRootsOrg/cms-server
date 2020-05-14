<template>
  <div class="test">
    <h1>{{ msg }}</h1>
    <p>
      Select an option to the left. If you have not registered or logged in, do that first.
    </p>
  </div>
</template>

<script>
export default {
  name: "Test",
  props: {
    msg: String
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
