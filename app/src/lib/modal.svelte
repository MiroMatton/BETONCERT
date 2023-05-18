<script lang="ts">
  export let showModal = false;
  export let company: any;
</script>

{#if showModal}
  <div class="modal-overlay" on:click|self>
    <div class="modal">
      <h1>{company.Name}</h1>
    </div>
  </div>
{/if}

<style lang="scss">
  @import "/src/styles/_variables.scss";
  .modal-overlay {
    position: fixed;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    z-index: 100;
  }
  .modal {
    position: absolute;
    width: 50%;
    height: auto; /* or specify a fixed height if you prefer */
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    background-color: $white-color;
    align-items: center;
    justify-content: center;
    padding: 4rem 0;
    border-radius: 32px;
  }
</style>
