<script lang="ts">
  import { onMount } from "svelte";
  import CertificateItem from "$lib/CertificateItem.svelte";
  interface Certificate {
    _id: string;
    certificatenumber: string;
    certificatereport: number;
    certificationnotlicensed: boolean;
    certificationstatusid: number;
    companyid: number;
    id: number;
    notlicensed: boolean;
    notlicensedmessage: null | string;
    product: string;
    sectorid: number;
    standard: string;
    statusid: number;
    suspended: boolean;
    technicalspecification: null | string;
  }

  let data: Certificate[] = [];
  let currentPage = 1;
  let searchQuery = "";

  function fetchData() {
    fetch(`http://localhost:8080/api/certificates?page=${currentPage}`)
      .then((response) => response.json())
      .then((json: Certificate[]) => {
        data = json;
      });
  }

  function goToPage(page: number) {
    currentPage = page;
    fetchData();
  }

  function handleSearchSubmit(event: Event) {
    event.preventDefault();

    if (!searchQuery.trim()) {
      currentPage = 1;
      fetchData();
      return;
    }

    fetch(
      `http://localhost:8080/api/certificates?page=${currentPage}&q=${searchQuery}`
    )
      .then((response) => response.json())
      .then((json: Certificate[]) => {
        data = json;
      });
  }

  export function initSearchForm(submitHandler: (event: Event) => void): void {
    const searchForm = document.querySelector(".searchContainer form");
    if (!searchForm) return;

    searchForm.addEventListener("submit", submitHandler);
  }

  onMount(() => {
    fetchData();
    initSearchForm(handleSearchSubmit);
  });
</script>

<header>
  <img src="/src/assets/logo.svg" alt="logo" />
  <div class="searchContainer">
    <form>
      <input type="text" placeholder="> Wat zoek je" bind:value={searchQuery} />
      <button type="submit">
        <img src="/src/assets/search.svg" alt="search icon" />
      </button>
    </form>
  </div>
  <p>Hello thiery</p>
</header>
<main>
  <div id="filter-container">
    <h3>Filters</h3>
    <h4>Product</h4>
    <label
      ><input type="checkbox" name="filter1" value="filter1" />Filter 1</label
    ><br />
    <label
      ><input type="checkbox" name="filter2" value="filter2" />Filter 2</label
    ><br />
    <label
      ><input type="checkbox" name="filter3" value="filter3" />Filter 3</label
    ><br />
    <label
      ><input type="checkbox" name="filter4" value="filter4" />Filter 4</label
    ><br />
    <label
      ><input type="checkbox" name="filter5" value="filter5" />Filter 5</label
    ><br />
    <h4>Toepassing</h4>
    <label
      ><input type="checkbox" name="filter6" value="filter6" />Filter 6</label
    ><br />
    <label
      ><input type="checkbox" name="filter7" value="filter7" />Filter 7</label
    ><br />
    <label
      ><input type="checkbox" name="filter8" value="filter8" />Filter 8</label
    ><br />
    <label
      ><input type="checkbox" name="filter9" value="filter9" />Filter 9</label
    ><br />
    <label
      ><input type="checkbox" name="filter10" value="filter10" />Filter 10</label
    ><br />
    <h4>Licentiestatus</h4>
    <label
      ><input
        type="checkbox"
        name="filter11"
        value="filter11"
      />gecertificeerd</label
    ><br />
    <label
      ><input
        type="checkbox"
        name="filter12"
        value="filter12"
      />geschorst</label
    ><br />
  </div>

  <div id="certificate-container">
    <div class="tabs">
      <span class="tab active">All Certificates</span>
      <span class="tab">Favorites</span>
    </div>
    <div>
      <button
        disabled={currentPage === 1}
        on:click={() => goToPage(currentPage - 1)}>Prev</button
      >
      <button
        disabled={currentPage === 86}
        on:click={() => goToPage(currentPage + 1)}>Next</button
      >
      <span>{currentPage}</span>
    </div>
    {#each data as certificate}
      <CertificateItem {certificate} />
    {/each}
  </div>
</main>

<style lang="scss">
  @import "/src/styles/main.scss";
</style>
