<script lang="ts">
  import { onMount } from "svelte";
  import CertificateItem from "$lib/CertificateItem.svelte";
  import Modal from "$lib/modal.svelte";
  interface Certificate {
    _id: string;
    certificateNumber: string;
    certificateReport: number;
    certificationNotLicensed: boolean;
    certificationStatusID: number;
    companyId: number;
    id: number;
    notLicensed: boolean;
    notLicensedMessage: string | null;
    product: string;
    sectorId: number;
    standard: string;
    statusID: number;
    suspended: boolean;
    technicalSpecification: string | null;
    isFavourite?: boolean;
  }

  interface Company {
    Id: number;
    Name: string;
    Address: string;
    Zip: string;
    City: string;
    CountryId: number;
    Tel: string;
    VAT: string;
    categoryId: number;
    ProductionEntities: ProductionEntity[];
  }

  interface ProductionEntity {
    Id: number;
    Name: string;
    Address: string;
    Zip: string;
    City: string;
    Tel: string;
  }

  let certificates: Certificate[] = [];
  let currentPage = 1;
  let perPage = 25;
  let searchQuery = "";
  let totalPages: number;
  let activeTab = "all";
  let activeCategories: string[] = [];
  let showModal = false;
  let company: Company;

  const toggleModal = () => {
    showModal = !showModal;
  };

  function handleCategoryChange(event: Event): void {
    const category = event.target;
    if (category instanceof HTMLInputElement && category.checked) {
      activeCategories.push(category.value);
    } else if (category instanceof HTMLElement) {
      const index = activeCategories.indexOf(category.getAttribute("value")!);
      activeCategories.splice(index, 1);
    }
    currentPage = 1;
    fetchData();
  }

  async function fetchData(): Promise<void> {
    const url = new URL(`http://localhost:8080/api/certificates`);

    url.searchParams.set("mode", String(activeTab));
    url.searchParams.set("page", String(currentPage));
    url.searchParams.set("perPage", String(perPage));

    if (activeCategories.length > 0) {
      url.searchParams.set("products", activeCategories.join(","));
    }

    if (searchQuery.trim()) {
      url.searchParams.set("q", searchQuery);
    }

    try {
      const response = await fetch(url.toString());

      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }

      const json = await response.json();
      certificates = json.Certificates;
      totalPages = json.TotalPages;
    } catch (error) {
      console.error(error);
    }
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

    currentPage = 1;
    fetchData();
  }

  export function initSearchForm(submitHandler: (event: Event) => void): void {
    const searchForm = document.querySelector(".searchContainer form");
    if (!searchForm) return;

    searchForm.addEventListener("submit", submitHandler);
  }

  const check = () => {
    if (!("serviceWorker" in navigator)) {
      throw new Error("No Service Worker support!");
    }
    if (!("PushManager" in window)) {
      throw new Error("No Push API Support!");
    }
  };

  const requestNotificationPermission = async () => {
    const permission = await window.Notification.requestPermission();
    if (permission !== "granted") {
      throw new Error("Permission not granted for Notification");
    }
  };

  const main = async () => {
    check();
    await requestNotificationPermission();
  };

  onMount(() => {
    main();
    fetchData();
    initSearchForm(handleSearchSubmit);
  });
</script>

<Modal {company} bind:showModal on:click={toggleModal} />
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
    <h4>Categorie</h4>
    <label
      ><input
        type="checkbox"
        class="category"
        name="mortar"
        value="1"
        on:change={handleCategoryChange}
      /> Mortar</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="concrete"
        value="2"
        on:change={handleCategoryChange}
      /> Concrete</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="fly-ash"
        value="23"
        on:change={handleCategoryChange}
      /> Fly Ash</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="aggregates"
        value="4"
        on:change={handleCategoryChange}
      /> Aggregates</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="rental-company"
        value="20"
        on:change={handleCategoryChange}
      /> Rental Company</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="road-concrete"
        value="19"
        on:change={handleCategoryChange}
      /> Road Concrete</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="hydraulic-road-binders"
        value="18"
        on:change={handleCategoryChange}
      /> Hydraulic Road Binders</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="hydraulically-bound-mixtures"
        value="17"
        on:change={handleCategoryChange}
      /> Hydraulically Bound Mixtures</label
    ><br />
    <label
      ><input type="checkbox" class="category" name="ggbs" value="ggbs" /> GGBS</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="approved-blast-furnace-slag"
        value="15"
        on:change={handleCategoryChange}
      /> Approved Blast Furnace Slag</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="cement-distribution"
        value="14"
        on:change={handleCategoryChange}
      /> Cement Distribution</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="cement"
        value="13"
        on:change={handleCategoryChange}
      /> Cement</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="pigments"
        value="12"
        on:change={handleCategoryChange}
      /> Pigments</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="lime"
        value="11"
        on:change={handleCategoryChange}
      /> Lime</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="admixtures"
        value="10"
        on:change={handleCategoryChange}
      /> Admixtures</label
    ><br />
    <label
      ><input
        type="checkbox"
        class="category"
        name="fly-ash-distribution"
        value="3"
        on:change={handleCategoryChange}
      /> Fly Ash Distribution</label
    ><br />
    <h4>Licentiestatus</h4>
    <label
      ><input type="checkbox" name="filter11" value="filter11" /> geldig</label
    ><br />
    <label
      ><input type="checkbox" name="filter12" value="filter12" /> ongeldig</label
    ><br />
  </div>

  <div id="certificate-container">
    <div class="tabs">
      <span
        class="tab"
        class:selected={activeTab === "all"}
        class:active={activeTab === "all" ? "active-tab" : ""}
        on:click={() => {
          activeTab = "all";
          currentPage = 1;
          fetchData();
        }}>All Certificates</span
      >
      <span
        class="tab"
        class:selected={activeTab === "favorites"}
        class:active={activeTab === "favorites" ? "active-tab" : ""}
        on:click={() => {
          activeTab = "favorites";
          currentPage = 1;
          fetchData();
        }}>Favorites</span
      >
    </div>
    <div class="page-navigation">
      <button
        disabled={currentPage === 1}
        on:click={() => goToPage(currentPage - 1)}
        ><img src="/src/assets/arrowLeft.svg" alt="voorige pagina " />
      </button>
      <span>{currentPage}</span>
      <button
        disabled={currentPage === totalPages}
        on:click={() => goToPage(currentPage + 1)}
        ><img src="/src/assets/arrowRight.svg" alt="volgende pagina" /></button
      >
    </div>
    <div class="certificate-list">
      {#if certificates && certificates.length > 0}
        {#each certificates as certificate}
          <CertificateItem {certificate} bind:showModal bind:company />
        {/each}
      {:else}
        <p>No certificates found.</p>
      {/if}
    </div>
  </div>
</main>

<style lang="scss">
  @import "/src/styles/main.scss";
</style>
