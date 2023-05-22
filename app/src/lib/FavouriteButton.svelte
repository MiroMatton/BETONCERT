<script lang="ts">
  export let Id: number;
  export let IsFavourite: boolean;
  import favourite from "/src/assets/favourite.svg";
  import unFavourite from "/src/assets/unFavourite.svg";

  async function toggleFavourite() {
    IsFavourite = !IsFavourite; // toggle the value of IsFavourite

    try {
      // send PUT request to update the favourite status on the server
      const url = `http://localhost:8080/api/favourite/${Id}`;
      const response = await fetch(url, {
        method: "PUT",
        body: JSON.stringify({ IsFavourite }),
        headers: {
          "Content-Type": "application/json",
        },
        mode: "cors",
      });

      // check if request was successful
      if (response.ok) {
        const responseData = await response.json();
        console.log(responseData.message);
      } else {
        console.error(response.statusText);
      }
    } catch (error) {
      console.error(error);
    }
  }
</script>

<img
  src={IsFavourite ? favourite : unFavourite}
  alt="favourite button"
  on:click={toggleFavourite}
  on:keydown={(event) => {
    if (event.key === "Enter" || event.key === " ") {
      toggleFavourite();
    }
  }}
/>

<style lang="scss">
  img {
    margin: 0 0.5rem;
    cursor: pointer;
  }
</style>
